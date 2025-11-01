package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"ku-work/backend/model"
	"log"
	"time"

	"gorm.io/gorm"
)

// generateAnonymousID creates a unique anonymous identifier based on the original ID
// This ensures consistency while maintaining anonymity
func generateAnonymousID(originalID string) string {
	hash := sha256.Sum256([]byte(originalID + time.Now().String()))
	return "ANON-" + hex.EncodeToString(hash[:])[:12]
}

// AnonymizeExpiredAccounts anonymizes accounts that have been soft-deleted
func AnonymizeExpiredAccounts(db *gorm.DB, gracePeriodDay int) error {
	gracePeriodDuration := time.Duration(gracePeriodDay) * 24 * time.Hour
	cutoffTime := time.Now().Add(-gracePeriodDuration)

	log.Printf("Starting expired accounts anonymization (grace period: %d days)", gracePeriodDay)

	// Find all users that were soft-deleted before the cutoff time
	var users []model.User
	if err := db.Unscoped().
		Where("deleted_at IS NOT NULL").
		Where("deleted_at < ?", cutoffTime).
		Find(&users).Error; err != nil {
		log.Printf("Error finding expired accounts: %v", err)
		return err
	}

	if len(users) == 0 {
		log.Println("No expired accounts found for anonymization")
		return nil
	}

	log.Printf("Found %d expired accounts to anonymize", len(users))

	// Anonymize each user
	anonymizedCount := 0
	errorCount := 0

	for _, user := range users {
		// Check if already anonymized
		if len(user.Username) > 5 && user.Username[:5] == "ANON-" {
			log.Printf("Account %s already anonymized, skipping", user.ID)
			continue
		}

		if err := AnonymizeAccount(db, user.ID); err != nil {
			log.Printf("Error anonymizing account %s: %v", user.ID, err)
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

// DisableCompanyJobPosts disables all job posts for a deactivated company
func DisableCompanyJobPosts(db *gorm.DB, companyUserID string) error {
	log.Printf("Disabling job posts for company: %s", companyUserID)

	// Update all jobs for this company to set is_open = false
	result := db.Model(&model.Job{}).
		Where("company_id = ? AND is_open = ?", companyUserID, true).
		Update("is_open", false)

	if result.Error != nil {
		log.Printf("Error disabling job posts for company %s: %v", companyUserID, result.Error)
		return fmt.Errorf("failed to disable job posts: %w", result.Error)
	}

	log.Printf("Disabled %d job posts for company: %s", result.RowsAffected, companyUserID)
	return nil
}

// AnonymizeJobApplicationsForStudent anonymizes and deletes files from all job applications for a student
func AnonymizeJobApplicationsForStudent(tx *gorm.DB, studentUserID string) error {
	log.Printf("Anonymizing job applications for student: %s", studentUserID)

	// Find all job applications for this student
	var applications []model.JobApplication
	if err := tx.Unscoped().
		Preload("Files").
		Where("user_id = ?", studentUserID).
		Find(&applications).Error; err != nil {
		return fmt.Errorf("failed to find job applications: %w", err)
	}

	if len(applications) == 0 {
		log.Printf("No job applications found for student: %s", studentUserID)
		return nil
	}

	log.Printf("Found %d job applications to anonymize for student: %s", len(applications), studentUserID)

	anonymousID := generateAnonymousID(studentUserID)

	for _, app := range applications {
		// Delete all files associated with this application
		for _, file := range app.Files {
			// Delete the physical file using the registered storage delete hook
			if err := model.CallStorageDeleteHook(tx.Statement.Context, file.ID); err != nil {
				log.Printf("Warning: Failed to delete file %s: %v", file.ID, err)
			}
			// Delete the file record from database
			if err := tx.Unscoped().Delete(&file).Error; err != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", file.ID, err)
			}
		}

		// Anonymize application contact information
		updates := map[string]any{
			"contact_phone": "",
			"contact_email": fmt.Sprintf("%s@anonymized.local", anonymousID),
		}

		if err := tx.Unscoped().Model(&model.JobApplication{}).
			Where("job_id = ? AND user_id = ?", app.JobID, studentUserID).
			Updates(updates).Error; err != nil {
			log.Printf("Warning: Failed to anonymize application for job %d: %v", app.JobID, err)
		}
	}

	log.Printf("Successfully anonymized %d job applications for student: %s", len(applications), studentUserID)
	return nil
}

// AnonymizeAccount anonymizes a user account and all associated personal data
// This complies with Thailand's PDPA while retaining data for analytics
func AnonymizeAccount(db *gorm.DB, userID string) error {
	log.Printf("Anonymizing account: %s", userID)

	return db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Unscoped().Where("id = ?", userID).First(&user).Error; err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		anonymousID := generateAnonymousID(userID)

		if err := tx.Unscoped().Model(&user).Updates(map[string]any{
			"username":      anonymousID,
			"password_hash": "",
		}).Error; err != nil {
			log.Printf("Error anonymizing user record for user %s: %v", userID, err)
			return fmt.Errorf("failed to anonymize user: %w", err)
		}
		log.Printf("Anonymized user record for: %s", userID)

		// Check if user is a student
		var student model.Student
		isStudent := tx.Unscoped().Where("user_id = ?", userID).First(&student).Error == nil

		// Check if user is a company
		var company model.Company
		isCompany := tx.Unscoped().Where("user_id = ?", userID).First(&company).Error == nil

		// Anonymize Student record if exists
		if isStudent {
			if err := AnonymizeStudentData(tx, &student); err != nil {
				return fmt.Errorf("failed to anonymize student data: %w", err)
			}
			log.Printf("Anonymized student record for user: %s", userID)

			// Anonymize job applications if student
			if err := AnonymizeJobApplicationsForStudent(tx, userID); err != nil {
				return fmt.Errorf("failed to anonymize job applications: %w", err)
			}
			log.Printf("Anonymized job applications for student: %s", userID)
		}

		// Anonymize Company record if exists
		if isCompany {
			if err := AnonymizeCompanyData(tx, &company); err != nil {
				return fmt.Errorf("failed to anonymize company data: %w", err)
			}
			log.Printf("Anonymized company record for user: %s", userID)
		}

		// Anonymize Google OAuth details if exists
		var googleOAuth model.GoogleOAuthDetails
		if err := tx.Unscoped().Where("user_id = ?", userID).First(&googleOAuth).Error; err == nil {
			if err := AnonymizeGoogleOAuthData(tx, &googleOAuth); err != nil {
				return fmt.Errorf("failed to anonymize OAuth data: %w", err)
			}
			log.Printf("Anonymized OAuth details for user: %s", userID)
		}

		log.Printf("Successfully anonymized account: %s", userID)
		return nil
	})
}

// AnonymizeStudentData anonymizes all personal data in a student record
func AnonymizeStudentData(tx *gorm.DB, student *model.Student) error {
	anonymousID := generateAnonymousID(student.UserID)

	updates := map[string]any{
		"phone":                  nil,
		"photo_id":               nil,
		"birth_date":             time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		"about_me":               "",
		"git_hub":                "",
		"linked_in":              "",
		"student_id":             anonymousID,
		"major":                  "Anonymized",
		"student_status_file_id": nil, // Remove document reference
	}

	if err := tx.Unscoped().Model(student).Updates(updates).Error; err != nil {
		return err
	}

	// Delete associated files (photos, documents)
	if student.PhotoID != "" {
		var photo model.File
		if err := tx.Where("id = ?", student.PhotoID).First(&photo).Error; err == nil {
			if err := model.CallStorageDeleteHook(tx.Statement.Context, photo.ID); err != nil {
				log.Printf("Warning: Failed to delete file %s: %v", photo.ID, err)
			}
			if err := tx.Unscoped().Delete(&photo).Error; err != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", photo.ID, err)
			}
		}
	}

	if student.StudentStatusFileID != "" {
		var statusFile model.File
		if err := tx.Where("id = ?", student.StudentStatusFileID).First(&statusFile).Error; err == nil {
			if err := model.CallStorageDeleteHook(tx.Statement.Context, statusFile.ID); err != nil {
				log.Printf("Warning: Failed to delete status file %s: %v", statusFile.ID, err)
			}
			if err := tx.Unscoped().Delete(&statusFile).Error; err != nil {
				log.Printf("Warning: Failed to delete status file record %s: %v", statusFile.ID, err)
			}
		}
	}

	return nil
}

// AnonymizeCompanyData anonymizes all personal data in a company record
func AnonymizeCompanyData(tx *gorm.DB, company *model.Company) error {
	anonymousID := generateAnonymousID(company.UserID)

	updates := map[string]any{
		"email":     fmt.Sprintf("%s@anonymized.local", anonymousID),
		"website":   "",
		"phone":     "",
		"photo_id":  nil,
		"banner_id": nil,
		"about_us":  "",
		"address":   "",
		"city":      "Anonymized",
		"country":   "Anonymized",
	}

	if err := tx.Unscoped().Model(company).Updates(updates).Error; err != nil {
		return err
	}

	// Delete associated files (photos, banners)
	if company.PhotoID != "" {
		var photo model.File
		if err := tx.Where("id = ?", company.PhotoID).First(&photo).Error; err == nil {
			if err := model.CallStorageDeleteHook(tx.Statement.Context, photo.ID); err != nil {
				log.Printf("Warning: Failed to delete file %s: %v", photo.ID, err)
			}
			if err := tx.Unscoped().Delete(&photo).Error; err != nil {
				log.Printf("Warning: Failed to delete file record %s: %v", photo.ID, err)
			}
		}
	}

	if company.BannerID != "" {
		var banner model.File
		if err := tx.Where("id = ?", company.BannerID).First(&banner).Error; err == nil {
			if err := model.CallStorageDeleteHook(tx.Statement.Context, banner.ID); err != nil {
				log.Printf("Warning: Failed to delete banner file %s: %v", banner.ID, err)
			}
			if err := tx.Unscoped().Delete(&banner).Error; err != nil {
				log.Printf("Warning: Failed to delete banner file record %s: %v", banner.ID, err)
			}
		}
	}

	return nil
}

// AnonymizeGoogleOAuthData anonymizes OAuth details
func AnonymizeGoogleOAuthData(tx *gorm.DB, oauth *model.GoogleOAuthDetails) error {
	anonymousID := generateAnonymousID(oauth.UserID)

	updates := map[string]any{
		"external_id": anonymousID,
		"first_name":  "Anonymized",
		"last_name":   "User",
		"email":       fmt.Sprintf("%s@anonymized.local", anonymousID),
	}

	return tx.Unscoped().Model(oauth).Updates(updates).Error
}

// CheckIfAnonymized checks if an account has already been anonymized
func CheckIfAnonymized(db *gorm.DB, userID string) (bool, error) {
	var user model.User
	if err := db.Unscoped().Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}

	// Check if username starts with ANON-
	return len(user.Username) > 5 && user.Username[:5] == "ANON-", nil
}
