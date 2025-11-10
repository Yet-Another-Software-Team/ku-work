package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ku-work/backend/model"
	"ku-work/backend/providers/ai"
	repo "ku-work/backend/repository"
)

// AIService coordinates automated approval checks for Jobs and Students.
// It no longer depends on a raw *gorm.DB; all persistence is done through repositories.
// Email notifications are delegated to the EventBus (if provided).
type AIService struct {
	approvalAI  ai.ApprovalAI
	jobRepo     repo.JobRepository
	studentRepo repo.StudentRepository
	eventBus    *EventBus
}

// NewAIService constructs an AIService using dependency injection only.
// All dependencies must be non-nil.
func NewAIService(approvalAI ai.ApprovalAI, jobRepo repo.JobRepository, studentRepo repo.StudentRepository, eventBus *EventBus) (*AIService, error) {
	if approvalAI == nil {
		return nil, errors.New("approvalAI is required")
	}
	if jobRepo == nil {
		return nil, errors.New("jobRepo is required")
	}
	if studentRepo == nil {
		return nil, errors.New("studentRepo is required")
	}
	if eventBus == nil {
		return nil, errors.New("eventBus is required")
	}
	return &AIService{
		approvalAI:  approvalAI,
		jobRepo:     jobRepo,
		studentRepo: studentRepo,
		eventBus:    eventBus,
	}, nil
}

// AutoApproveJob performs an AI check on the job identified by jobID.
// If the AI returns accepted/rejected (not pending), it persists the decision and
// publishes an email notification event (when EventBus is configured).
func (s *AIService) AutoApproveJob(ctx context.Context, jobID uint) error {
	job, err := s.jobRepo.FindJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job %d not found", jobID)
	}

	status, reasons := s.approvalAI.CheckJob(job)
	if status == model.JobApprovalPending {
		// Nothing to persist; still under review.
		return nil
	}

	approve := status == model.JobApprovalAccepted
	reason := joinReasons(reasons)

	if err := s.jobRepo.ApproveOrRejectJob(ctx, job.ID, approve, "ai", reason); err != nil {
		return fmt.Errorf("persist job approval: %w", err)
	}

	detail, err := s.jobRepo.GetJobDetail(ctx, job.ID)
	if err != nil || detail == nil {
		// Non-fatal: approval persisted; return nil.
		return nil
	}
	company, err := s.jobRepo.FindCompanyByUserID(ctx, detail.CompanyID)
	if err != nil || company == nil || company.Email == "" {
		return nil
	}

	_ = s.eventBus.PublishEmailJobApproval(EmailJobApprovalEvent{
		CompanyEmail:    company.Email,
		CompanyUsername: detail.CompanyName,
		JobName:         detail.Name,
		JobPosition:     detail.Position,
		Status:          string(status),
		Reason:          reason,
	})

	return nil
}

// AutoApproveJobModel variant that accepts a loaded *model.Job.
// Uses background context for repository calls (caller can supply its own if needed).
func (s *AIService) AutoApproveJobModel(job *model.Job) error {
	if job == nil {
		return errors.New("job is nil")
	}
	return s.AutoApproveJob(context.Background(), job.ID)
}

// AutoApproveStudent performs an AI check on the student identified by userID.
// If accepted/rejected, it persists via repository and publishes an email notification event.
func (s *AIService) AutoApproveStudent(ctx context.Context, userID string) error {
	student, err := s.studentRepo.FindStudentByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find student: %w", err)
	}
	if student == nil {
		return fmt.Errorf("student %s not found", userID)
	}

	status, reasons := s.approvalAI.CheckStudent(student)
	if status == model.StudentApprovalPending {
		return nil
	}

	approve := status == model.StudentApprovalAccepted
	reason := joinReasons(reasons)

	if err := s.studentRepo.ApproveOrRejectStudent(ctx, userID, approve, "ai", reason); err != nil {
		return fmt.Errorf("persist student approval: %w", err)
	}

	if s.eventBus != nil {
		profile, err := s.studentRepo.FindStudentProfileByUserID(ctx, userID)
		if err != nil || profile == nil || profile.Email == "" {
			return nil
		}
		_ = s.eventBus.PublishEmailStudentApproval(EmailStudentApprovalEvent{
			Email:     profile.Email,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Status:    string(status),
			Reason:    reason,
		})
	}

	return nil
}

// AutoApproveStudentModel variant that accepts a loaded *model.Student.
func (s *AIService) AutoApproveStudentModel(student *model.Student) error {
	if student == nil {
		return errors.New("student is nil")
	}
	return s.AutoApproveStudent(context.Background(), student.UserID)
}

// PublishAsyncJobCheck enqueues a job ID for asynchronous AI evaluation using the EventBus AI queue.
// Falls back to synchronous processing if no EventBus is configured.
func (s *AIService) PublishAsyncJobCheck(jobID uint) error {
	if s.eventBus == nil {
		// Fallback: synchronous
		return s.AutoApproveJob(context.Background(), jobID)
	}
	return s.eventBus.PublishAIJobCheck(jobID)
}

// Helper to format reasons list.
func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	return "- " + strings.Join(reasons, "\n- ")
}
