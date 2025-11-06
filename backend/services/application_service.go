package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"
)

// ApplicationService handles job application flows.
type ApplicationService struct {
	applicationRepo repo.ApplicationRepository
	jobRepo         repo.JobRepository
	studentRepo     repo.StudentRepository
	identityRepo    repo.IdentityRepository

	fileService  *FileService
	emailService *EmailService

	// Optional email templates
	jobApplicationStatusUpdateEmailTemplate *template.Template
	newApplicantEmailTemplate               *template.Template
}

// NewApplicationService constructs an ApplicationService.
func NewApplicationService(
	appRepo repo.ApplicationRepository,
	jobRepo repo.JobRepository,
	studentRepo repo.StudentRepository,
	identityRepo repo.IdentityRepository,
	fileSvc *FileService,
	emailSvc *EmailService,
	statusTpl *template.Template,
	newApplicantTpl *template.Template,
) *ApplicationService {
	return &ApplicationService{
		applicationRepo:                         appRepo,
		jobRepo:                                 jobRepo,
		studentRepo:                             studentRepo,
		identityRepo:                            identityRepo,
		fileService:                             fileSvc,
		emailService:                            emailSvc,
		jobApplicationStatusUpdateEmailTemplate: statusTpl,
		newApplicantEmailTemplate:               newApplicantTpl,
	}
}

// ApplyToJobParams groups inputs for applying to a job.
type ApplyToJobParams struct {
	// Authenticated user (student) id
	UserID string
	// Target job id
	JobID uint
	// Contact overrides; if empty, caller may pre-populate with defaults (e.g., student's phone/email).
	ContactPhone string
	ContactEmail string
	// Files to attach (e.g., resume, cover letter). Max 2 files is handled by the caller/validator.
	Files []*multipart.FileHeader

	// Framework context is needed by FileService.SaveFile for streaming upload.
	// While this leaks transport concerns into service, it's the current provider contract.
	GinCtx *gin.Context
}

// ApplyToJob creates a new job application.
func (s *ApplicationService) ApplyToJob(ctx context.Context, p ApplyToJobParams) error {
	if s == nil {
		return fmt.Errorf("application service is not initialized")
	}
	if p.GinCtx == nil {
		return fmt.Errorf("gin context is required for file upload")
	}
	if p.UserID == "" {
		return fmt.Errorf("user id is required")
	}
	if p.JobID == 0 {
		return fmt.Errorf("job id is required")
	}

	stu, err := s.studentRepo.FindStudentByUserID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to load student: %w", err)
	}
	if stu.ApprovalStatus != model.StudentApprovalAccepted {
		return fmt.Errorf("student is not approved")
	}

	if p.ContactPhone == "" {
		p.ContactPhone = stu.Phone
	}
	if p.ContactEmail == "" {
		u, err := s.identityRepo.FindUserByID(ctx, p.UserID, false)
		if err != nil {
			return fmt.Errorf("failed to load user: %w", err)
		}
		// Username of OAuth user is their email
		p.ContactEmail = u.Username
	}

	job, err := s.jobRepo.FindJobByID(ctx, p.JobID)
	if err != nil {
		return fmt.Errorf("failed to load job: %w", err)
	}
	if job.ApprovalStatus != model.JobApprovalAccepted {
		return fmt.Errorf("job is not approved")
	}

	app := model.JobApplication{
		UserID:       p.UserID,
		JobID:        job.ID,
		ContactPhone: p.ContactPhone,
		ContactEmail: p.ContactEmail,
		Status:       model.JobApplicationPending,
	}

	for _, fh := range p.Files {
		f, err := s.fileService.SaveFile(p.GinCtx, p.UserID, fh, model.FileCategoryDocument)
		if err != nil {
			return fmt.Errorf("failed to save file %s: %w", fh.Filename, err)
		}
		app.Files = append(app.Files, *f)
	}

	if err := s.applicationRepo.CreateApplication(ctx, &app); err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	// Notify company if configured
	if s.emailService != nil && job.NotifyOnApplication {
		go func(job model.Job, createdAt time.Time) {
			// Resolve recipient (company)
			company, err := s.jobRepo.FindCompanyByUserID(context.Background(), job.CompanyID)
			if err != nil || company == nil || company.Email == "" {
				// Best effort: silently return on failure to build recipient
				return
			}

			subject := fmt.Sprintf("[KU-Work] New Application for %s - %s", job.Name, job.Position)

			// Use template if available, otherwise send a simple fallback body
			if s.newApplicantEmailTemplate != nil {
				type tplCtx struct {
					CompanyUser model.User
					Job         model.Job
					Applicant   struct {
						// Placeholder fields for template compatibility
						UserID    string
						FirstName string
						LastName  string
						Email     string
					}
					Application struct {
						Date time.Time
					}
				}
				var tctx tplCtx
				// We only have the company username if we can infer it; if not available, leave zero value.
				tctx.CompanyUser.ID = job.CompanyID
				tctx.Job = job
				tctx.Application.Date = createdAt.In(time.FixedZone("Asia/Bangkok", 7*60*60)) // BST (+7)
				tctx.Applicant.UserID = p.UserID                                              // If template doesn't use it, harmless.

				var buf bytes.Buffer
				if err := s.newApplicantEmailTemplate.Execute(&buf, tctx); err == nil {
					_ = s.emailService.SendTo(company.Email, subject, buf.String())
					return
				}
				// Fall through to simple body on template error
			}

			body := fmt.Sprintf("You have a new application for \"%s - %s\".\n\nOpen KU-Work to review the applicant's details.", job.Name, job.Position)
			_ = s.emailService.SendTo(company.Email, subject, body)
		}(*job, app.CreatedAt)
	}

	return nil
}

// GetApplicationsForJob lists applications for a job.
func (s *ApplicationService) GetApplicationsForJob(ctx context.Context, jobID uint, params *repo.FetchJobApplicationsParams) ([]repo.ShortApplicationDetail, error) {
	if jobID == 0 {
		return nil, fmt.Errorf("job id is required")
	}
	return s.applicationRepo.GetApplicationsForJob(ctx, jobID, params)
}

// ClearJobApplications deletes applications matching included statuses.
func (s *ApplicationService) ClearJobApplications(ctx context.Context, jobID uint, includePending, includeRejected, includeAccepted bool) (int64, error) {
	if jobID == 0 {
		return 0, fmt.Errorf("job id is required")
	}
	return s.applicationRepo.ClearJobApplications(ctx, jobID, includePending, includeRejected, includeAccepted)
}

// GetApplicationByJobAndEmail returns a single application detail.
func (s *ApplicationService) GetApplicationByJobAndEmail(ctx context.Context, jobID uint, email string) (*repo.FullApplicantDetail, error) {
	if jobID == 0 {
		return nil, fmt.Errorf("job id is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	return s.applicationRepo.GetApplicationByJobAndEmail(ctx, jobID, email)
}

// GetAllApplicationsForUser returns applications visible to the user.
func (s *ApplicationService) GetAllApplicationsForUser(ctx context.Context, userID string, params *repo.FetchAllApplicationsParams) ([]repo.ApplicationWithJobDetails, int64, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("user id is required")
	}
	return s.applicationRepo.GetAllApplicationsForUser(ctx, userID, params)
}

// UpdateStatusParams groups inputs for updating an application's status.
type UpdateStatusParams struct {
	JobID         uint
	StudentUserID string
	NewStatus     model.JobApplicationStatus

	NotifyApplicantEmail string // If provided and email service configured, an email will be sent
	CompanyName          string // Used in email body/template
}

// UpdateJobApplicationStatus updates the application's status and optionally notifies the applicant via email.
func (s *ApplicationService) UpdateJobApplicationStatus(ctx context.Context, p UpdateStatusParams) error {
	if p.JobID == 0 {
		return fmt.Errorf("job id is required")
	}
	if p.StudentUserID == "" {
		return fmt.Errorf("student user id is required")
	}
	if p.NewStatus == "" {
		return fmt.Errorf("new status is required")
	}

	if err := s.applicationRepo.UpdateApplicationStatus(ctx, p.JobID, p.StudentUserID, p.NewStatus); err != nil {
		return err
	}

	if s.emailService != nil && p.NotifyApplicantEmail != "" {

		job, err := s.jobRepo.FindJobByID(ctx, p.JobID)
		if err != nil {

			return nil
		}

		subject := fmt.Sprintf("[KU-Work] Your Application Status for %s - %s", job.Name, job.Position)

		if s.jobApplicationStatusUpdateEmailTemplate != nil {
			type tplCtx struct {
				OAuth struct {
					Email     string
					FirstName string
					LastName  string
				}
				Job         *model.Job
				CompanyName string
				Status      string
			}
			var tctx tplCtx
			tctx.OAuth.Email = p.NotifyApplicantEmail
			tctx.Job = job
			tctx.CompanyName = p.CompanyName
			tctx.Status = string(p.NewStatus)

			var buf bytes.Buffer
			if err := s.jobApplicationStatusUpdateEmailTemplate.Execute(&buf, tctx); err == nil {
				_ = s.emailService.SendTo(p.NotifyApplicantEmail, subject, buf.String())
				return nil
			}
			// Fall through to simple body on template error
		}

		body := fmt.Sprintf("Your application for \"%s - %s\" is now: %s.\n\nThank you for applying on KU-Work.", job.Name, job.Position, p.NewStatus)
		_ = s.emailService.SendTo(p.NotifyApplicantEmail, subject, body)
	}

	return nil
}
