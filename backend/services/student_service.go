package services

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"time"

	"ku-work/backend/model"
	filehandling "ku-work/backend/providers/file_handling"
	repo "ku-work/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// ErrAlreadyRegistered is returned when a user attempts to register as a student more than once.
var ErrAlreadyRegistered = errors.New("user already registered as a student")

// StudentService encapsulates business logic for student registration, profile retrieval, and approval workflows.
// Handlers call this service; the service delegates DB access to repositories and uses other infrastructure services
// (AI, Email, File) as collaborators.
type StudentService struct {
	repo         repo.StudentRepository
	identityRepo repo.IdentityRepository
	fileService  *FileService
	eventBus     *EventBus
}

// StudentRegistrationInput represents the payload required to register a user as a student.
type StudentRegistrationInput struct {
	Phone             string                `json:"phone"`
	BirthDate         string                `json:"birthDate"` // RFC3339 format
	AboutMe           string                `json:"aboutMe"`
	GitHub            string                `json:"github"`
	LinkedIn          string                `json:"linkedIn"`
	StudentID         string                `json:"studentId"`
	Major             string                `json:"major"`
	StudentStatus     string                `json:"studentStatus"`
	Photo             *multipart.FileHeader `json:"-"`
	StudentStatusFile *multipart.FileHeader `json:"-"`
}

// NewStudentService constructs a StudentService.
func NewStudentService(
	repo repo.StudentRepository,
	identityRepo repo.IdentityRepository,
	fileService *FileService,
	eventBus *EventBus,
) *StudentService {
	if repo == nil {
		log.Fatal("student repository cannot be nil")
	}
	if identityRepo == nil {
		log.Fatal("identity repository cannot be nil")
	}
	if fileService == nil {
		log.Fatal("file service cannot be nil")
	}
	if eventBus == nil {
		// EventBus is optional; when nil, email/AI events are skipped.
	}
	return &StudentService{
		repo:         repo,
		identityRepo: identityRepo,
		fileService:  fileService,
		eventBus:     eventBus,
	}
}

// RegisterStudent handles the student registration flow.
func (s *StudentService) RegisterStudent(ctx *gin.Context, userID string, input StudentRegistrationInput) error {
	// Ensure service is configured
	if s == nil || s.repo == nil {
		return errors.New("student service not initialized")
	}
	if s.fileService == nil && input.Photo != nil {
		// Fallback to global provider if FileService is not injected
		if _, err := filehandling.GetProvider(); err != nil {
			return errors.New("file service not configured")
		}
	}

	// Check if already registered
	exists, err := s.repo.ExistsByUserID(ctx.Request.Context(), userID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyRegistered
	}

	// Parse birth date
	parsedBirthDate, err := time.Parse(time.RFC3339, input.BirthDate)
	if err != nil {
		return err
	}

	// Save files
	var photo *model.File
	if input.Photo != nil {
		if s.fileService != nil {
			photo, err = s.fileService.SaveFile(ctx, userID, input.Photo, model.FileCategoryImage)
		} else {
			// Use provider directly if FileService is not injected
			provider, pErr := filehandling.GetProvider()
			if pErr != nil {
				return pErr
			}
			photo, err = provider.SaveFile(ctx, nil, userID, input.Photo, model.FileCategoryImage)
		}
		if err != nil {
			return err
		}
	}

	var statusDocument *model.File
	if input.StudentStatusFile != nil {
		if s.fileService != nil {
			statusDocument, err = s.fileService.SaveFile(ctx, userID, input.StudentStatusFile, model.FileCategoryDocument)
		} else {
			// Use provider directly if FileService is not injected
			provider, pErr := filehandling.GetProvider()
			if pErr != nil {
				return pErr
			}
			statusDocument, err = provider.SaveFile(ctx, nil, userID, input.StudentStatusFile, model.FileCategoryDocument)
		}
		if err != nil {
			return err
		}
	}

	// Build student model
	student := model.Student{
		UserID:         userID,
		ApprovalStatus: model.StudentApprovalPending,
		Phone:          input.Phone,
		BirthDate:      datatypes.Date(parsedBirthDate),
		AboutMe:        input.AboutMe,
		GitHub:         input.GitHub,
		LinkedIn:       input.LinkedIn,
		StudentID:      input.StudentID,
		Major:          input.Major,
		StudentStatus:  input.StudentStatus,
	}
	if photo != nil {
		student.PhotoID = photo.ID
		student.Photo = *photo
	}
	if statusDocument != nil {
		student.StudentStatusFileID = statusDocument.ID
		student.StudentStatusFile = *statusDocument
	}

	// Persist
	if err := s.repo.CreateStudent(ctx.Request.Context(), &student); err != nil {
		return err
	}

	if s.eventBus != nil {
		s.eventBus.PublishAIStudentCheck(student.UserID)
	}

	return nil
}

// GetStudentProfile returns a single student's profile (student + oauth fields).
// If the target account is deactivated, personal data is anonymized.
func (s *StudentService) GetStudentProfile(ctx context.Context, userID string) (*repo.StudentProfile, error) {
	profile, err := s.repo.FindStudentProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	deactivated, err := s.identityRepo.IsUserDeactivated(ctx, userID)
	if err != nil {
		return nil, err
	}
	if deactivated {
		anonymizeStudentProfile(profile)
	}

	return profile, nil
}

// ListStudentProfiles returns a list of student profiles according to the given filter
// and anonymizes profiles belonging to deactivated accounts.
func (s *StudentService) ListStudentProfiles(ctx context.Context, filter repo.StudentListFilter) ([]repo.StudentProfile, error) {
	items, err := s.repo.ListStudentProfiles(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Anonymize deactivated accounts
	for i := range items {
		uid := items[i].UserID
		if uid == "" {
			continue
		}
		deactivated, derr := s.identityRepo.IsUserDeactivated(ctx, uid)
		if derr != nil {
			// Best-effort: skip anonymization on error
			continue
		}
		if deactivated {
			anonymizeStudentProfile(&items[i])
		}
	}

	return items, nil
}

// ApproveOrRejectStudent updates the student's approval status and records an audit entry.
// If email wiring is configured, it sends an approval/rejection email using the provided template.
func (s *StudentService) ApproveOrRejectStudent(ctx context.Context, targetUserID string, approve bool, actorID, reason string) error {
	if s == nil || s.repo == nil {
		return errors.New("student service not initialized")
	}

	// Persist approval decision (also creates audit)
	if err := s.repo.ApproveOrRejectStudent(ctx, targetUserID, approve, actorID, reason); err != nil {
		return err
	}
	oauth, err := s.identityRepo.FindGoogleOAuthByUserID(ctx, targetUserID, false)
	if err != nil || oauth == nil {
		// Best-effort: skip emailing if we can't resolve address
		return nil
	}

	var status string
	if approve {
		status = string(model.StudentApprovalAccepted)
	} else {
		status = string(model.StudentApprovalRejected)
	}

	event := EmailStudentApprovalEvent{
		Email:     oauth.Email,
		FirstName: oauth.FirstName,
		LastName:  oauth.LastName,
		Reason:    reason,
		Status:    status,
	}

	if s.eventBus != nil {
		s.eventBus.PublishEmailStudentApproval(event)
	}

	return nil
}

// anonymizeStudentProfile removes PII fields from a StudentProfile for deactivated accounts.
func anonymizeStudentProfile(s *repo.StudentProfile) {
	if s == nil {
		return
	}
	s.FirstName = ""
	s.LastName = ""
	s.Email = ""

	s.Phone = ""
	s.PhotoID = ""
	s.AboutMe = ""
	s.GitHub = ""
	s.LinkedIn = ""
	s.StudentID = ""
	s.StudentStatusFileID = ""
	s.Photo = model.File{}
	s.StudentStatusFile = model.File{}

	s.FullName = "Deactivated Account"
}
