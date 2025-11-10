package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

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

	fileService *FileService
	eventBus    *EventBus
}

// NewApplicationService constructs an ApplicationService.
func NewApplicationService(
	appRepo repo.ApplicationRepository,
	jobRepo repo.JobRepository,
	studentRepo repo.StudentRepository,
	identityRepo repo.IdentityRepository,
	fileSvc *FileService,
	eventBus *EventBus,
) *ApplicationService {
	// Check nil
	if appRepo == nil || jobRepo == nil || studentRepo == nil || identityRepo == nil || fileSvc == nil {
		log.Fatal("Application service requires non-nil core dependencies")
	}
	return &ApplicationService{
		applicationRepo: appRepo,
		jobRepo:         jobRepo,
		studentRepo:     studentRepo,
		identityRepo:    identityRepo,
		fileService:     fileSvc,
		eventBus:        eventBus,
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
	if p.GinCtx == nil {
		return fmt.Errorf("gin context is required for file upload")
	}
	if p.UserID == "" {
		return fmt.Errorf("user id is required")
	}
	if p.JobID == 0 {
		return fmt.Errorf("job id is required")
	}

	student, err := s.studentRepo.FindStudentByUserID(ctx, p.UserID)
	if err != nil {
		return fmt.Errorf("failed to load student: %w", err)
	}
	if student.ApprovalStatus != model.StudentApprovalAccepted {
		return fmt.Errorf("student is not approved")
	}

	if p.ContactPhone == "" {
		p.ContactPhone = student.Phone
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

	application := model.JobApplication{
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
		application.Files = append(application.Files, *f)
	}

	if err := s.applicationRepo.CreateApplication(ctx, &application); err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	// Use EventBus to sent New application email if company opted in.
	if job.NotifyOnApplication {
		company, err := s.jobRepo.FindCompanyByUserID(context.Background(), job.CompanyID)
		if err != nil || company == nil || company.Email == "" {
			// Best effort: silently return on failure to build recipient
			return nil
		}
		user, err := s.identityRepo.FindUserByID(context.Background(), job.CompanyID, false)
		if err != nil {
			// Best effort: silently return on failure to build recipient
			return nil
		}

		applicant, err := s.identityRepo.FindGoogleOAuthByUserID(context.Background(), application.UserID, false)
		if err != nil {
			// Best effort: silently return on failure to build recipient
			return nil
		}

		event := EmailJobNewApplicantEvent{
			CompanyEmail:       company.Email,
			CompanyUsername:    user.Username,
			JobName:            job.Name,
			JobPosition:        job.Position,
			ApplicantFirstName: applicant.FirstName,
			ApplicantLastName:  applicant.LastName,
			ApplicationDate:    application.CreatedAt,
		}

		if s.eventBus != nil {
			_ = s.eventBus.PublishEmailJobNewApplicant(event)
		}
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
	JobID                uint
	StudentUserID        string
	NewStatus            model.JobApplicationStatus
	CompanyName          string // Used in email body/template
	NotifyApplicantEmail string // If provided, triggers email event
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

	if p.NotifyApplicantEmail != "" {
		// Best-effort enrichment for email event; failures are non-fatal.
		jobDetail, err := s.jobRepo.GetJobDetail(ctx, p.JobID)
		if err == nil && jobDetail != nil {
			oauth, _ := s.identityRepo.FindGoogleOAuthByUserID(ctx, p.StudentUserID, false)

			event := EmailJobApplicationStatusEvent{
				Email:       p.NotifyApplicantEmail,
				FirstName:   "",
				LastName:    "",
				JobName:     jobDetail.Name,
				JobPosition: jobDetail.Position,
				CompanyName: jobDetail.CompanyName,
				Status:      string(p.NewStatus),
			}
			if oauth != nil {
				event.FirstName = oauth.FirstName
				event.LastName = oauth.LastName
			}

			if s.eventBus != nil {
				_ = s.eventBus.PublishEmailJobApplicationStatus(event)
			}
		}
	}

	return nil
}
