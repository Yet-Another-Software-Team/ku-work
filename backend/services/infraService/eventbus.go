package infraService

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ku-work/backend/model"
	"ku-work/backend/providers/ai"
	repo "ku-work/backend/repository"
)

// EventBus is a small, KISS-style event dispatcher with two worker pools:
// - AI pool: executes Approval AI checks for Jobs and Students, persists results via repositories, then enqueues emails.
// - Email pool: renders templates and sends emails via EmailService.
//
// It is intentionally simple: not a full pub/sub implementation, just two typed queues.
type EventBus struct {
	// Dependencies (DI)
	ai           ai.ApprovalAI
	jobRepo      repo.JobRepository
	studentRepo  repo.StudentRepository
	emailService *EmailService

	// Email templates
	tplJobApproval              *template.Template
	tplStudentApproval          *template.Template
	tplJobApplicationStatus     *template.Template
	tplJobNewApplicant          *template.Template
	emailSubjectPrefix          string
	jobApprovalSubjectFormat    string
	studentApprovalSubject      string
	jobApplicationSubjectFormat string
	jobNewApplicantSubjectFmt   string

	// Queues
	aiQueue    chan any
	emailQueue chan any

	// Lifecycle
	stopCh      chan struct{}
	wg          sync.WaitGroup
	closeOnce   sync.Once
	isClosedMux sync.RWMutex
	isClosed    bool
}

// EventBusOptions configures a new EventBus.
type EventBusOptions struct {
	AI          ai.ApprovalAI
	JobRepo     repo.JobRepository
	StudentRepo repo.StudentRepository
	Email       *EmailService

	// Optional: provide pre-parsed templates (recommended). If nil, the bus will try to parse from default paths.
	JobApprovalTpl          *template.Template
	StudentApprovalTpl      *template.Template
	JobApplicationStatusTpl *template.Template
	JobNewApplicantTpl      *template.Template

	// Worker sizes and buffer sizes (optional; defaults shown below)
	AIWorkers      int
	EmailWorkers   int
	AIQueueSize    int
	EmailQueueSize int
	SubjectPrefix  string // e.g., "[KU-Work]"
}

// NewEventBus constructs and starts the worker pools.
// It prefers injected templates; if not provided, it will attempt to parse from:
// - backend/templates/emails/job_approval_status_update.tmpl
// - backend/templates/emails/student_approval_status_update.tmpl
// - backend/templates/emails/job_application_status_update.tmpl
// - backend/templates/emails/job_new_applicant.tmpl
func NewEventBus(opts EventBusOptions) (*EventBus, error) {
	// Validate DI
	if opts.AI == nil {
		return nil, errors.New("AI provider is required")
	}
	if opts.JobRepo == nil {
		return nil, errors.New("JobRepository is required")
	}
	if opts.StudentRepo == nil {
		return nil, errors.New("StudentRepository is required")
	}
	if opts.Email == nil {
		return nil, errors.New("EmailService is required")
	}

	// Defaults
	if opts.AIWorkers <= 0 {
		opts.AIWorkers = 2
	}
	if opts.EmailWorkers <= 0 {
		opts.EmailWorkers = 2
	}
	if opts.AIQueueSize <= 0 {
		opts.AIQueueSize = 128
	}
	if opts.EmailQueueSize <= 0 {
		opts.EmailQueueSize = 256
	}
	if opts.SubjectPrefix == "" {
		opts.SubjectPrefix = "[KU-Work]"
	}

	// Templates (prefer injected)
	tplJobApproval := opts.JobApprovalTpl
	tplStudentApproval := opts.StudentApprovalTpl
	tplJobApplicationStatus := opts.JobApplicationStatusTpl
	tplJobNewApplicant := opts.JobNewApplicantTpl

	// If any missing, parse from default paths
	// Support running from project root or from backend/ by probing both directories.
	// Priority: backend/templates/emails (monorepo layout) then templates/emails (tests or alt cwd).
	resolveTemplate := func(filename string) (string, error) {
		candidates := []string{
			filepath.Join("backend", "templates", "emails", filename),
			filepath.Join("templates", "emails", filename),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
		return "", fmt.Errorf("template %s not found in known locations", filename)
	}

	if tplJobApproval == nil {
		path, err := resolveTemplate("job_approval_status_update.tmpl")
		if err != nil {
			return nil, fmt.Errorf("parse job approval email template: %w", err)
		}
		t, err := template.New("job_approval_status_update.tmpl").ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parse job approval email template: %w", err)
		}
		tplJobApproval = t
	}
	if tplStudentApproval == nil {
		path, err := resolveTemplate("student_approval_status_update.tmpl")
		if err != nil {
			return nil, fmt.Errorf("parse student approval email template: %w", err)
		}
		t, err := template.New("student_approval_status_update.tmpl").ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parse student approval email template: %w", err)
		}
		tplStudentApproval = t
	}
	if tplJobApplicationStatus == nil {
		path, err := resolveTemplate("job_application_status_update.tmpl")
		if err != nil {
			return nil, fmt.Errorf("parse job application status email template: %w", err)
		}
		t, err := template.New("job_application_status_update.tmpl").ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parse job application status email template: %w", err)
		}
		tplJobApplicationStatus = t
	}
	if tplJobNewApplicant == nil {
		path, err := resolveTemplate("job_new_applicant.tmpl")
		if err != nil {
			return nil, fmt.Errorf("parse job new applicant email template: %w", err)
		}
		t, err := template.New("job_new_applicant.tmpl").ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("parse job new applicant email template: %w", err)
		}
		tplJobNewApplicant = t
	}

	bus := &EventBus{
		ai:           opts.AI,
		jobRepo:      opts.JobRepo,
		studentRepo:  opts.StudentRepo,
		emailService: opts.Email,

		tplJobApproval:              tplJobApproval,
		tplStudentApproval:          tplStudentApproval,
		tplJobApplicationStatus:     tplJobApplicationStatus,
		tplJobNewApplicant:          tplJobNewApplicant,
		emailSubjectPrefix:          opts.SubjectPrefix,
		jobApprovalSubjectFormat:    "%s Your \"%s - %s\" job has been reviewed",
		studentApprovalSubject:      "%s Your student account has been reviewed",
		jobApplicationSubjectFormat: "%s Update on your application for %s (%s) at %s",
		jobNewApplicantSubjectFmt:   "%s New applicant for %s - %s",

		aiQueue:    make(chan any, opts.AIQueueSize),
		emailQueue: make(chan any, opts.EmailQueueSize),
		stopCh:     make(chan struct{}),
	}

	// Start workers
	for i := 0; i < opts.AIWorkers; i++ {
		bus.wg.Add(1)
		go bus.aiWorker()
	}
	for i := 0; i < opts.EmailWorkers; i++ {
		bus.wg.Add(1)
		go bus.emailWorker()
	}

	return bus, nil
}

// Close stops the EventBus and waits for workers to drain queues.
func (b *EventBus) Close() {
	b.closeOnce.Do(func() {
		b.isClosedMux.Lock()
		b.isClosed = true
		b.isClosedMux.Unlock()

		close(b.stopCh)
		close(b.aiQueue)
		close(b.emailQueue)
		b.wg.Wait()
	})
}

// PublishAIJobCheck enqueues a job check event for AI worker pool.
func (b *EventBus) PublishAIJobCheck(jobID uint) error {
	return b.enqueue(b.aiQueue, AIJobCheckEvent{JobID: jobID})
}

// PublishAIStudentCheck enqueues a student check event for AI worker pool.
func (b *EventBus) PublishAIStudentCheck(userID string) error {
	return b.enqueue(b.aiQueue, AIStudentCheckEvent{UserID: userID})
}

// PublishEmailJobApproval enqueues a job approval notification email.
func (b *EventBus) PublishEmailJobApproval(e EmailJobApprovalEvent) error {
	return b.enqueue(b.emailQueue, e)
}

// PublishEmailStudentApproval enqueues a student approval notification email.
func (b *EventBus) PublishEmailStudentApproval(e EmailStudentApprovalEvent) error {
	return b.enqueue(b.emailQueue, e)
}

// PublishEmailJobApplicationStatus enqueues a job application status email to the applicant.
func (b *EventBus) PublishEmailJobApplicationStatus(e EmailJobApplicationStatusEvent) error {
	return b.enqueue(b.emailQueue, e)
}

// PublishEmailJobNewApplicant enqueues an email to the employer about a new applicant.
func (b *EventBus) PublishEmailJobNewApplicant(e EmailJobNewApplicantEvent) error {
	return b.enqueue(b.emailQueue, e)
}

func (b *EventBus) enqueue(queue chan any, v any) error {
	b.isClosedMux.RLock()
	closed := b.isClosed
	b.isClosedMux.RUnlock()
	if closed {
		return errors.New("event bus closed")
	}
	select {
	case queue <- v:
		return nil
	case <-b.stopCh:
		return errors.New("event bus stopped")
	}
}

/*
 * AI worker and Event
 */

type AIJobCheckEvent struct {
	JobID uint
}

type AIStudentCheckEvent struct {
	UserID string
}

func (b *EventBus) aiWorker() {
	defer b.wg.Done()
	for ev := range b.aiQueue {
		switch e := ev.(type) {
		case AIJobCheckEvent:
			b.handleAIJobCheck(e)
		case AIStudentCheckEvent:
			b.handleAIStudentCheck(e)
		}
	}
}

func (b *EventBus) handleAIJobCheck(e AIJobCheckEvent) {
	ctx := context.Background()

	// 1) Load job
	job, err := b.jobRepo.FindJobByID(ctx, e.JobID)
	if err != nil || job == nil {
		return
	}

	// 2) AI check
	status, reasons := b.ai.CheckJob(job)
	if status == model.JobApprovalPending {
		return
	}
	approve := status == model.JobApprovalAccepted
	reason := formatReasons(reasons)

	// 3) Persist via repository (also records audit)
	if err := b.jobRepo.ApproveOrRejectJob(ctx, job.ID, approve, "ai", reason); err != nil {
		return
	}

	// 4) Fetch detail for email (company email + username + job name/position)
	detail, err := b.jobRepo.GetJobDetail(ctx, job.ID)
	if err != nil || detail == nil {
		return
	}
	company, err := b.jobRepo.FindCompanyByUserID(ctx, detail.CompanyID)
	if err != nil || company == nil || company.Email == "" {
		return
	}

	// 5) Enqueue email
	emailEv := EmailJobApprovalEvent{
		CompanyEmail:    company.Email,
		CompanyUsername: detail.CompanyName,
		JobName:         detail.Name,
		JobPosition:     detail.Position,
		Status:          string(status),
		Reason:          reason,
	}
	_ = b.enqueue(b.emailQueue, emailEv)
}

func (b *EventBus) handleAIStudentCheck(e AIStudentCheckEvent) {
	ctx := context.Background()

	// 1) Load student
	student, err := b.studentRepo.FindStudentByUserID(ctx, e.UserID)
	if err != nil || student == nil {
		return
	}

	// 2) AI check
	status, reasons := b.ai.CheckStudent(student)
	if status == model.StudentApprovalPending {
		return
	}
	approve := status == model.StudentApprovalAccepted
	reason := formatReasons(reasons)

	// 3) Persist via repository (also records audit)
	if err := b.studentRepo.ApproveOrRejectStudent(ctx, e.UserID, approve, "ai", reason); err != nil {
		return
	}

	// 4) Load profile for email recipient
	profile, err := b.studentRepo.FindStudentProfileByUserID(ctx, e.UserID)
	if err != nil || profile == nil || profile.Email == "" {
		return
	}

	// 5) Enqueue email
	emailEv := EmailStudentApprovalEvent{
		Email:     profile.Email,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Status:    string(status),
		Reason:    reason,
	}
	_ = b.enqueue(b.emailQueue, emailEv)
}

func formatReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	return "- " + strings.Join(reasons, "\n- ")
}

/*
 * Email Worker and Event
 */

type EmailJobApprovalEvent struct {
	CompanyEmail    string
	CompanyUsername string
	JobName         string
	JobPosition     string
	Status          string
	Reason          string
}

type EmailStudentApprovalEvent struct {
	Email     string
	FirstName string
	LastName  string
	Status    string
	Reason    string
}

type EmailJobApplicationStatusEvent struct {
	Email       string
	FirstName   string
	LastName    string
	JobName     string
	JobPosition string
	CompanyName string
	Status      string
}

type EmailJobNewApplicantEvent struct {
	CompanyEmail       string
	CompanyUsername    string
	JobName            string
	JobPosition        string
	ApplicantFirstName string
	ApplicantLastName  string
	ApplicationDate    time.Time
}

func (b *EventBus) emailWorker() {
	defer b.wg.Done()
	for ev := range b.emailQueue {
		switch e := ev.(type) {
		case EmailJobApprovalEvent:
			b.sendJobApprovalEmail(e)
		case EmailStudentApprovalEvent:
			b.sendStudentApprovalEmail(e)
		case EmailJobApplicationStatusEvent:
			b.sendJobApplicationStatusEmail(e)
		case EmailJobNewApplicantEvent:
			b.sendJobNewApplicantEmail(e)
		}
	}
}

func (b *EventBus) sendJobApprovalEmail(e EmailJobApprovalEvent) {
	// Build template context to match job_approval_status_update.tmpl
	type ctxUser struct {
		Username string
	}
	type ctxJob struct {
		Name     string
		Position string
	}
	type tplCtx struct {
		User   ctxUser
		Job    ctxJob
		Status string
		Reason string
	}
	context := tplCtx{
		User:   ctxUser{Username: e.CompanyUsername},
		Job:    ctxJob{Name: e.JobName, Position: e.JobPosition},
		Status: e.Status,
		Reason: e.Reason,
	}

	var buf bytes.Buffer
	if err := b.tplJobApproval.Execute(&buf, context); err != nil {
		return
	}
	subject := fmt.Sprintf(b.jobApprovalSubjectFormat, b.emailSubjectPrefix, e.JobName, e.JobPosition)
	_ = b.emailService.SendTo(e.CompanyEmail, subject, buf.String())
}

func (b *EventBus) sendStudentApprovalEmail(e EmailStudentApprovalEvent) {
	// Build template context to match student_approval_status_update.tmpl
	type ctxOAuth struct {
		FirstName string
		LastName  string
		Email     string
	}
	type tplCtx struct {
		OAuth  ctxOAuth
		Status string
		Reason string
	}
	context := tplCtx{
		OAuth:  ctxOAuth{FirstName: e.FirstName, LastName: e.LastName, Email: e.Email},
		Status: e.Status,
		Reason: e.Reason,
	}
	var buf bytes.Buffer
	if err := b.tplStudentApproval.Execute(&buf, context); err != nil {
		return
	}
	subject := fmt.Sprintf(b.studentApprovalSubject, b.emailSubjectPrefix)
	_ = b.emailService.SendTo(e.Email, subject, buf.String())
}

func (b *EventBus) sendJobApplicationStatusEmail(e EmailJobApplicationStatusEvent) {
	// Build template context to match job_application_status_update.tmpl
	type ctxOAuth struct {
		FirstName string
		LastName  string
	}
	type ctxJob struct {
		Position string
		Name     string
	}
	type tplCtx struct {
		OAuth       ctxOAuth
		Job         ctxJob
		CompanyName string
		Status      string
	}
	context := tplCtx{
		OAuth:       ctxOAuth{FirstName: e.FirstName, LastName: e.LastName},
		Job:         ctxJob{Position: e.JobPosition, Name: e.JobName},
		CompanyName: e.CompanyName,
		Status:      e.Status,
	}
	var buf bytes.Buffer
	if err := b.tplJobApplicationStatus.Execute(&buf, context); err != nil {
		return
	}
	subject := fmt.Sprintf(b.jobApplicationSubjectFormat, b.emailSubjectPrefix, e.JobPosition, e.JobName, e.CompanyName)
	_ = b.emailService.SendTo(e.Email, subject, buf.String())
}

func (b *EventBus) sendJobNewApplicantEmail(e EmailJobNewApplicantEvent) {
	// Build template context to match job_new_applicant.tmpl
	type ctxCompanyUser struct {
		Username string
	}
	type ctxJob struct {
		Name     string
		Position string
	}
	type ctxApplicant struct {
		FirstName string
		LastName  string
	}
	type ctxApplication struct {
		Date time.Time
	}
	type tplCtx struct {
		CompanyUser ctxCompanyUser
		Job         ctxJob
		Applicant   ctxApplicant
		Application ctxApplication
	}
	context := tplCtx{
		CompanyUser: ctxCompanyUser{Username: e.CompanyUsername},
		Job:         ctxJob{Name: e.JobName, Position: e.JobPosition},
		Applicant:   ctxApplicant{FirstName: e.ApplicantFirstName, LastName: e.ApplicantLastName},
		Application: ctxApplication{Date: e.ApplicationDate},
	}
	var buf bytes.Buffer
	if err := b.tplJobNewApplicant.Execute(&buf, context); err != nil {
		return
	}
	subject := fmt.Sprintf(b.jobNewApplicantSubjectFmt, b.emailSubjectPrefix, e.JobName, e.JobPosition)
	_ = b.emailService.SendTo(e.CompanyEmail, subject, buf.String())
}
