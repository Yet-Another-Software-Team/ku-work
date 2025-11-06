package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/mail"
	"net/url"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	filehandling "ku-work/backend/providers/file_handling"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// IdentityService merges account lifecycle and user-related operations behind a single, layered service.
// - Handlers call this service.
// - This service depends only on repository and infra services (e.g., FileService).
// - Only repository implementations perform DB access.
type IdentityService struct {
	Repo        repo.IdentityRepository
	FileService *FileService
}

// Error definitions
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrAlreadyDeactivated  = errors.New("account is already deactivated")
	ErrNotDeactivated      = errors.New("account is not deactivated")
	ErrGracePeriodExpired  = errors.New("grace period has expired")
	ErrAnonymizationFailed = errors.New("anonymization failed")

	ErrUsernameExists = errors.New("username already exists")
	ErrInvalidWebsite = errors.New("invalid website url")
)

// GracePeriodExpiredError enriches the error with deadline details for the handler to report.
type GracePeriodExpiredError struct {
	DeletedAt time.Time
	Deadline  time.Time
}

func (e *GracePeriodExpiredError) Error() string {
	return fmt.Sprintf("%s: deleted_at=%s deadline=%s", ErrGracePeriodExpired.Error(), e.DeletedAt.Format(time.RFC3339), e.Deadline.Format(time.RFC3339))
}

// NewIdentityService constructs the service with injected dependencies.
func NewIdentityService(r repo.IdentityRepository, fileSvc *FileService) *IdentityService {
	return &IdentityService{
		Repo:        r,
		FileService: fileSvc,
	}
}

// ---------------------------
// Account lifecycle
// ---------------------------

// DeactivateAccount soft-deletes the user account and performs side effects like disabling company job posts.
// It returns the deletionDate (now + gracePeriodDays) which the handler can present to the user.
func (s *IdentityService) DeactivateAccount(ctx context.Context, userID string, gracePeriodDays int) (time.Time, error) {
	if s == nil || s.Repo == nil {
		return time.Time{}, errors.New("service not initialized")
	}

	// Ensure user exists
	_, err := s.Repo.FindUserByID(ctx, userID, false)
	if err != nil {
		return time.Time{}, ErrUserNotFound
	}

	// Check deactivated state
	deactivated, err := s.Repo.IsUserDeactivated(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}
	if deactivated {
		return time.Time{}, ErrAlreadyDeactivated
	}

	// If user is a company, disable job posts (best-effort)
	if _, err := s.Repo.FindCompanyByUserID(ctx, userID, false); err == nil {
		if n, derr := s.Repo.DisableCompanyJobPosts(ctx, userID); derr != nil {
			log.Printf("Warning: Failed to disable job posts for company %s: %v", userID, derr)
		} else {
			log.Printf("Disabled %d job posts for company: %s", n, userID)
		}
	}

	// Soft delete user (triggers any registered hooks)
	if err := s.Repo.SoftDeleteUserByID(ctx, userID); err != nil {
		return time.Time{}, err
	}

	deletionDate := time.Now().Add(time.Duration(gracePeriodDays) * 24 * time.Hour)
	return deletionDate, nil
}

// ReactivateAccount restores a soft-deactivated account if within the grace period.
// Side effects: restores associated Student/Company/OAuth records that were soft-deleted.
func (s *IdentityService) ReactivateAccount(ctx context.Context, userID string, gracePeriodDays int) error {
	if s == nil || s.Repo == nil {
		return errors.New("service not initialized")
	}

	// Load user including soft-deleted records
	user, err := s.Repo.FindUserByID(ctx, userID, true)
	if err != nil {
		return ErrUserNotFound
	}

	// Must be deactivated to reactivate
	if !user.DeletedAt.Valid {
		return ErrNotDeactivated
	}

	// Check grace period
	gracePeriod := time.Duration(gracePeriodDays) * 24 * time.Hour
	deadline := user.DeletedAt.Time.Add(gracePeriod)
	if time.Now().After(deadline) {
		return &GracePeriodExpiredError{
			DeletedAt: user.DeletedAt.Time,
			Deadline:  deadline,
		}
	}

	// Restore user
	if err := s.Repo.RestoreUserByID(ctx, userID); err != nil {
		return err
	}

	// Restore related student record if exists
	if student, err := s.Repo.FindStudentByUserID(ctx, userID, true); err == nil && student.DeletedAt.Valid {
		if err := s.Repo.RestoreStudentByUserID(ctx, userID); err != nil {
			log.Printf("Warning: Failed to restore student for user %s: %v", userID, err)
		}
	}

	// Restore related company record if exists
	if company, err := s.Repo.FindCompanyByUserID(ctx, userID, true); err == nil && company.DeletedAt.Valid {
		if err := s.Repo.RestoreCompanyByUserID(ctx, userID); err != nil {
			log.Printf("Warning: Failed to restore company for user %s: %v", userID, err)
		}
	}

	// Restore OAuth details if exists
	if oauth, err := s.Repo.FindGoogleOAuthByUserID(ctx, userID, true); err == nil && oauth.DeletedAt.Valid {
		if err := s.Repo.RestoreGoogleOAuthByUserID(ctx, userID); err != nil {
			log.Printf("Warning: Failed to restore OAuth details for user %s: %v", userID, err)
		}
	}

	return nil
}

// ---------------------------
// Anonymization
// ---------------------------

// generateAnonymousID creates a unique anonymous identifier based on the original ID.
func generateAnonymousID(originalID string) string {
	hash := sha256.Sum256([]byte(originalID + time.Now().String()))
	return "ANON-" + hex.EncodeToString(hash[:])[:12]
}

// CheckIfAnonymized checks if an account has already been anonymized by inspecting the username prefix.
func (s *IdentityService) CheckIfAnonymized(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.Repo == nil {
		return false, errors.New("service not initialized")
	}
	user, err := s.Repo.FindUserByID(ctx, userID, true)
	if err != nil {
		return false, err
	}
	return len(user.Username) > 5 && user.Username[:5] == "ANON-", nil
}

// AnonymizeExpiredAccounts finds accounts whose grace period has expired and anonymizes them.
func (s *IdentityService) AnonymizeExpiredAccounts(ctx context.Context, gracePeriodDays int) error {
	if s == nil || s.Repo == nil {
		return errors.New("service not initialized")
	}

	graceDuration := time.Duration(gracePeriodDays) * 24 * time.Hour
	cutoff := time.Now().Add(-graceDuration)

	log.Printf("Starting expired accounts anonymization (grace period: %d days)", gracePeriodDays)

	users, err := s.Repo.ListSoftDeletedUsersBefore(ctx, cutoff)
	if err != nil {
		log.Printf("Error finding expired accounts: %v", err)
		return err
	}
	if len(users) == 0 {
		log.Println("No expired accounts found for anonymization")
		return nil
	}
	log.Printf("Found %d expired accounts to anonymize", len(users))

	anonymizedCount := 0
	errorCount := 0

	for _, u := range users {
		// Skip if already anonymized (idempotent)
		if len(u.Username) > 5 && u.Username[:5] == "ANON-" {
			log.Printf("Account %s already anonymized, skipping", u.ID)
			continue
		}

		if err := s.AnonymizeAccount(ctx, u.ID); err != nil {
			log.Printf("Error anonymizing account %s: %v", u.ID, err)
			errorCount++
			continue
		}
		anonymizedCount++
	}

	log.Printf("Account anonymization completed: %d anonymized, %d errors", anonymizedCount, errorCount)
	if errorCount > 0 {
		log.Printf("Warning: Some accounts failed to anonymize. See errors above.")
	}
	return nil
}

// AnonymizeAccount anonymizes a user account and all associated personal data,
// complying with PDPA while retaining data for analytics.
func (s *IdentityService) AnonymizeAccount(ctx context.Context, userID string) error {
	if s == nil || s.Repo == nil {
		return errors.New("service not initialized")
	}

	log.Printf("Anonymizing account: %s", userID)

	// Ensure user exists (unscoped)
	user, err := s.Repo.FindUserByID(ctx, userID, true)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAnonymizationFailed, err)
	}

	anon := generateAnonymousID(userID)

	// Anonymize user record
	if err := s.Repo.UpdateUserAnonymized(ctx, user.ID, map[string]any{
		"username":      anon,
		"password_hash": "",
	}); err != nil {
		log.Printf("Error anonymizing user record for user %s: %v", userID, err)
		return fmt.Errorf("%w: update user failed", ErrAnonymizationFailed)
	}
	log.Printf("Anonymized user record for: %s", userID)

	// If student, anonymize student data + job applications
	if student, err := s.Repo.FindStudentByUserID(ctx, userID, true); err == nil {
		if err := s.anonymizeStudentData(ctx, student, anon); err != nil {
			return fmt.Errorf("%w: anonymize student failed: %v", ErrAnonymizationFailed, err)
		}
		log.Printf("Anonymized student record for user: %s", userID)

		if err := s.anonymizeJobApplicationsForStudent(ctx, userID, anon); err != nil {
			return fmt.Errorf("%w: anonymize job applications failed: %v", ErrAnonymizationFailed, err)
		}
		log.Printf("Anonymized job applications for student: %s", userID)
	}

	// If company, anonymize company data
	if company, err := s.Repo.FindCompanyByUserID(ctx, userID, true); err == nil {
		if err := s.anonymizeCompanyData(ctx, company, anon); err != nil {
			return fmt.Errorf("%w: anonymize company failed: %v", ErrAnonymizationFailed, err)
		}
		log.Printf("Anonymized company record for user: %s", userID)
	}

	// If OAuth record exists, anonymize it
	if oauth, err := s.Repo.FindGoogleOAuthByUserID(ctx, userID, true); err == nil {
		if err := s.anonymizeGoogleOAuthData(ctx, oauth, anon); err != nil {
			return fmt.Errorf("%w: anonymize oauth failed: %v", ErrAnonymizationFailed, err)
		}
		log.Printf("Anonymized OAuth details for user: %s", userID)
	}

	log.Printf("Successfully anonymized account: %s", userID)
	return nil
}

// anonymizeJobApplicationsForStudent anonymizes and deletes files from all job applications for a student.
func (s *IdentityService) anonymizeJobApplicationsForStudent(ctx context.Context, studentUserID, anon string) error {
	log.Printf("Anonymizing job applications for student: %s", studentUserID)

	apps, err := s.Repo.ListJobApplicationsWithFilesByUserID(ctx, studentUserID)
	if err != nil {
		return fmt.Errorf("failed to list job applications: %w", err)
	}
	if len(apps) == 0 {
		log.Printf("No job applications found for student: %s", studentUserID)
		return nil
	}

	for _, app := range apps {
		// Delete all files associated with this application
		for _, f := range app.Files {
			// Best-effort delete of physical file using storage delete hook
			if derr := model.CallStorageDeleteHook(ctx, f.ID); derr != nil {
				log.Printf("Warning: Failed to delete file %s: %v", f.ID, derr)
			}
			// Delete the DB record unscoped
			if derr := s.Repo.UnscopedDeleteFileRecord(ctx, f.ID); derr != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", f.ID, derr)
			}
		}

		// Anonymize application contact information
		fields := map[string]any{
			"contact_phone": "",
			"contact_email": fmt.Sprintf("%s@anonymized.local", anon),
		}
		if err := s.Repo.UpdateJobApplicationFields(ctx, app.JobID, studentUserID, fields); err != nil {
			log.Printf("Warning: Failed to anonymize application for job %d: %v", app.JobID, err)
		}
	}

	log.Printf("Successfully anonymized %d job applications for student: %s", len(apps), studentUserID)
	return nil
}

// anonymizeStudentData anonymizes all personal data in a student record and deletes related files.
func (s *IdentityService) anonymizeStudentData(ctx context.Context, student *model.Student, anon string) error {
	fields := map[string]any{
		"phone":                  nil,
		"photo_id":               nil,
		"birth_date":             time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		"about_me":               "",
		"git_hub":                "",
		"linked_in":              "",
		"student_id":             anon,
		"major":                  "Anonymized",
		"student_status_file_id": nil,
	}

	if err := s.Repo.UpdateStudentFields(ctx, student.UserID, fields); err != nil {
		return err
	}

	// Delete associated files (photos, documents)
	if student.PhotoID != "" {
		if photo, err := s.Repo.FindFileByID(ctx, student.PhotoID); err == nil && photo != nil {
			if derr := model.CallStorageDeleteHook(ctx, photo.ID); derr != nil {
				log.Printf("Warning: Failed to delete file %s: %v", photo.ID, derr)
			}
			if derr := s.Repo.UnscopedDeleteFileRecord(ctx, photo.ID); derr != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", photo.ID, derr)
			}
		}
	}

	if student.StudentStatusFileID != "" {
		if statusFile, err := s.Repo.FindFileByID(ctx, student.StudentStatusFileID); err == nil && statusFile != nil {
			if derr := model.CallStorageDeleteHook(ctx, statusFile.ID); derr != nil {
				log.Printf("Warning: Failed to delete status file %s: %v", statusFile.ID, derr)
			}
			if derr := s.Repo.UnscopedDeleteFileRecord(ctx, statusFile.ID); derr != nil {
				log.Printf("Warning: Failed to delete status file record %s: %v", statusFile.ID, derr)
			}
		}
	}

	return nil
}

// anonymizeCompanyData anonymizes all personal data in a company record and deletes related images.
func (s *IdentityService) anonymizeCompanyData(ctx context.Context, company *model.Company, anon string) error {
	fields := map[string]any{
		"email":     fmt.Sprintf("%s@anonymized.local", anon),
		"website":   "",
		"phone":     "",
		"photo_id":  nil,
		"banner_id": nil,
		"about_us":  "",
		"address":   "",
		"city":      "Anonymized",
		"country":   "Anonymized",
	}

	if err := s.Repo.UpdateCompanyFields(ctx, company.UserID, fields); err != nil {
		return err
	}

	// Delete associated files (photos, banners)
	if company.PhotoID != "" {
		if photo, err := s.Repo.FindFileByID(ctx, company.PhotoID); err == nil && photo != nil {
			if derr := model.CallStorageDeleteHook(ctx, photo.ID); derr != nil {
				log.Printf("Warning: Failed to delete file %s: %v", photo.ID, derr)
			}
			if derr := s.Repo.UnscopedDeleteFileRecord(ctx, photo.ID); derr != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", photo.ID, derr)
			}
		}
	}

	if company.BannerID != "" {
		if banner, err := s.Repo.FindFileByID(ctx, company.BannerID); err == nil && banner != nil {
			if derr := model.CallStorageDeleteHook(ctx, banner.ID); derr != nil {
				log.Printf("Warning: Failed to delete banner file %s: %v", banner.ID, derr)
			}
			if derr := s.Repo.UnscopedDeleteFileRecord(ctx, banner.ID); derr != nil {
				log.Printf("Warning: Failed to delete banner file record %s: %v", banner.ID, derr)
			}
		}
	}

	return nil
}

// anonymizeGoogleOAuthData anonymizes OAuth details.
func (s *IdentityService) anonymizeGoogleOAuthData(ctx context.Context, oauth *model.GoogleOAuthDetails, anon string) error {
	fields := map[string]any{
		"external_id": anon,
		"first_name":  "Anonymized",
		"last_name":   "User",
		"email":       fmt.Sprintf("%s@anonymized.local", anon),
	}

	return s.Repo.UpdateOAuthFields(ctx, oauth.UserID, fields)
}

// ---------------------------
// Profile edits
// ---------------------------

// CompanyEditProfileInput represents editable fields for a company's profile.
// All fields are optional to support partial updates.
type CompanyEditProfileInput struct {
	Phone    *string               `json:"phone,omitempty"`
	Email    *string               `json:"email,omitempty"`
	Website  *string               `json:"website,omitempty"`
	Address  *string               `json:"address,omitempty"`
	City     *string               `json:"city,omitempty"`
	Country  *string               `json:"country,omitempty"`
	AboutUs  *string               `json:"about,omitempty"`
	Username *string               `json:"username,omitempty"`
	Photo    *multipart.FileHeader `json:"-"`
	Banner   *multipart.FileHeader `json:"-"`
}

// StudentEditProfileInput represents editable fields for a student's profile.
// All fields are optional except student status which can be empty if not updating.
type StudentEditProfileInput struct {
	Phone         *string               `json:"phone,omitempty"`
	BirthDate     *string               `json:"birthDate,omitempty"` // RFC3339 format
	AboutMe       *string               `json:"aboutMe,omitempty"`
	GitHub        *string               `json:"github,omitempty"`
	LinkedIn      *string               `json:"linkedIn,omitempty"`
	StudentStatus string                `json:"studentStatus,omitempty"`
	Photo         *multipart.FileHeader `json:"-"`
}

// UpdateCompanyProfile applies partial updates to a company profile and handles optional file uploads.
// This encapsulates business rules like username uniqueness and email/URL validation.
func (s *IdentityService) UpdateCompanyProfile(ctx *gin.Context, userID string, input CompanyEditProfileInput) error {
	// Handle username change with repository-backed uniqueness check and update.
	if input.Username != nil {
		exists, err := s.Repo.ExistsUsername(ctx.Request.Context(), *input.Username)
		if err != nil {
			return err
		}
		if exists {
			return ErrUsernameExists
		}
		if err := s.Repo.UpdateUserFields(ctx.Request.Context(), userID, map[string]any{
			"username": *input.Username,
		}); err != nil {
			return err
		}
	}

	// Verify company exists (non-alloc result ignored).
	if _, err := s.Repo.FindCompanyByUserID(ctx.Request.Context(), userID, false); err != nil {
		return err
	}

	updates := map[string]any{}

	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.Address != nil {
		updates["address"] = *input.Address
	}
	if input.City != nil {
		updates["city"] = *input.City
	}
	if input.Country != nil {
		updates["country"] = *input.Country
	}
	if input.AboutUs != nil {
		updates["about_us"] = *input.AboutUs
	}
	if input.Email != nil {
		if _, err := mail.ParseAddress(*input.Email); err != nil {
			return fmt.Errorf("invalid email address")
		}
		updates["email"] = *input.Email
	}
	if input.Website != nil {
		u, err := url.Parse(*input.Website)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return ErrInvalidWebsite
		}
		updates["website"] = *input.Website
	}

	// Handle file uploads if present.
	if s.FileService == nil && (input.Photo != nil || input.Banner != nil) {
		// Fallback to global provider if not injected
		if _, err := filehandling.GetProvider(); err != nil {
			return errors.New("file service not configured")
		}
	}
	if input.Photo != nil {
		if s.FileService != nil {
			photo, err := s.FileService.SaveFile(ctx, userID, input.Photo, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["photo_id"] = photo.ID
		} else {
			provider, _ := filehandling.GetProvider()
			photo, err := provider.SaveFile(ctx, nil, userID, input.Photo, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["photo_id"] = photo.ID
		}
	}
	if input.Banner != nil {
		if s.FileService != nil {
			banner, err := s.FileService.SaveFile(ctx, userID, input.Banner, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["banner_id"] = banner.ID
		} else {
			provider, _ := filehandling.GetProvider()
			banner, err := provider.SaveFile(ctx, nil, userID, input.Banner, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["banner_id"] = banner.ID
		}
	}

	if len(updates) > 0 {
		if err := s.Repo.UpdateCompanyFields(ctx.Request.Context(), userID, updates); err != nil {
			return err
		}
	}

	return nil
}

// UpdateStudentProfile applies partial updates to a student profile and handles optional photo upload.
func (s *IdentityService) UpdateStudentProfile(ctx *gin.Context, userID string, input StudentEditProfileInput) error {
	if s == nil || s.Repo == nil {
		return errors.New("service not initialized")
	}

	// Ensure the student record exists.
	if _, err := s.Repo.FindStudentByUserID(ctx.Request.Context(), userID, false); err != nil {
		return err
	}

	updates := map[string]any{}

	if input.BirthDate != nil {
		parsed, err := time.Parse(time.RFC3339, *input.BirthDate)
		if err != nil {
			return err
		}
		updates["birth_date"] = datatypes.Date(parsed)
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.AboutMe != nil {
		updates["about_me"] = *input.AboutMe
	}
	if input.GitHub != nil {
		updates["git_hub"] = *input.GitHub
	}
	if input.LinkedIn != nil {
		updates["linked_in"] = *input.LinkedIn
	}
	if input.StudentStatus != "" {
		updates["student_status"] = input.StudentStatus
	}

	if input.Photo != nil {
		if s.FileService != nil {
			photo, err := s.FileService.SaveFile(ctx, userID, input.Photo, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["photo_id"] = photo.ID
		} else {
			provider, err := filehandling.GetProvider()
			if err != nil {
				return err
			}
			photo, err := provider.SaveFile(ctx, nil, userID, input.Photo, model.FileCategoryImage)
			if err != nil {
				return err
			}
			updates["photo_id"] = photo.ID
		}
	}

	if len(updates) > 0 {
		if err := s.Repo.UpdateStudentFields(ctx.Request.Context(), userID, updates); err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------
// Role and profile helpers
// ---------------------------

// ResolveRole returns the effective role for a user using repository-backed checks.
func (s *IdentityService) ResolveRole(ctx context.Context, userID string) helper.Role {
	if s == nil || s.Repo == nil || userID == "" {
		return helper.Unknown
	}

	if cnt, err := s.Repo.CountAdminByUserID(userID); err == nil && cnt > 0 {
		return helper.Admin
	}
	if cnt, err := s.Repo.CountCompanyByUserID(userID); err == nil && cnt > 0 {
		return helper.Company
	}

	// Attempt to classify student role
	if u, err := s.Repo.FindUserByID(ctx, userID, true); err == nil && u != nil {
		if registered, roleStr, err := s.Repo.IsStudentRegisteredAndRole(*u); err == nil && registered {
			switch roleStr {
			case string(helper.Student):
				return helper.Student
			default:
				return helper.Viewer
			}
		}
	}

	return helper.Viewer
}

// GetUsername fetches a user's username (soft-deleted included for idempotency).
func (s *IdentityService) GetUsername(ctx context.Context, userID string) string {
	if s == nil || s.Repo == nil {
		return "unknown"
	}
	user, err := s.Repo.FindUserByID(ctx, userID, true)
	if err != nil || user == nil {
		return "unknown"
	}
	return user.Username
}
