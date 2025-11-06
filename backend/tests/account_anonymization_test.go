package tests

import (
	"context"
	"encoding/json"
	"ku-work/backend/handlers"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupAccountTestRouter creates a test router with account management endpoints
func setupAccountTestRouter(jwtHandlers *handlers.JWTHandlers, userHandlers *handlers.UserHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Protected routes
	protected := router.Group("", middlewares.AuthMiddleware(jwtHandlers.Service.JWTSecret, jwtHandlers.Service))
	protected.POST("/me/deactivate", userHandlers.DeactivateAccount)
	protected.POST("/me/reactivate", userHandlers.ReactivateAccount)
	protected.GET("/me", userHandlers.GetProfileHandler)

	return router
}

// TestAccountReactivation tests the account reactivation endpoint
func TestAccountReactivation(t *testing.T) {
	_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "30")

	// Create test user
	userInfo := UserCreationInfo{
		Username:  "reactivate_test_user",
		IsStudent: true,
	}
	userResult, err := CreateUser(userInfo)
	assert.NoError(t, err, "Should create test user")
	defer cleanupUser(t, userResult.User.ID)

	// Manually soft delete the user (bypass BeforeDelete hooks)
	now := time.Now()
	db.Unscoped().Model(&userResult.User).Update("deleted_at", now)
	if userResult.Student != nil {
		db.Unscoped().Model(userResult.Student).Update("deleted_at", now)
	}

	// Setup handlers and router
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	// removed stray line
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
	jwtHandlers := handlers.NewJWTHandlers(jwtService)
	// grace period handled inside handler
	accountService := services.NewIdentityService(identityRepo, fileService)
	userHandlers := handlers.NewUserHandlers(accountService)
	router := setupAccountTestRouter(jwtHandlers, userHandlers)

	token, _, err := jwtHandlers.GenerateTokens(userResult.User.ID)
	assert.NoError(t, err, "Should generate tokens")

	t.Run("Successfully reactivate account", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/me/reactivate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should reactivate successfully")

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse response")
		assert.Equal(t, "Account reactivated successfully", response["message"])
	})

	t.Run("Verify user is restored", func(t *testing.T) {
		var user model.User
		err := db.First(&user, "id = ?", userResult.User.ID).Error
		assert.NoError(t, err, "Should find user in normal query")
		assert.False(t, user.DeletedAt.Valid, "DeletedAt should be null")
	})

	t.Run("Verify student is restored", func(t *testing.T) {
		var student model.Student
		err := db.First(&student, "user_id = ?", userResult.User.ID).Error
		assert.NoError(t, err, "Should find student in normal query")
		assert.False(t, student.DeletedAt.Valid, "DeletedAt should be null")
	})

	t.Run("Cannot reactivate active account", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/me/reactivate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject")
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse response")
		assert.Equal(t, "Account is not deactivated", response["error"])
	})
}

// TestAccountReactivationGracePeriodExpired tests reactivation after grace period
func TestAccountReactivationGracePeriodExpired(t *testing.T) {
	_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "30")

	// Create test user
	userInfo := UserCreationInfo{
		Username:  "expired_test_user",
		IsStudent: true,
	}
	userResult, err := CreateUser(userInfo)
	assert.NoError(t, err, "Should create test user")
	defer cleanupUser(t, userResult.User.ID)

	// Soft delete the user with old timestamp (past grace period)
	oldDeletedAt := time.Now().Add(-31 * 24 * time.Hour)
	db.Unscoped().Model(&userResult.User).Update("deleted_at", oldDeletedAt)

	// Setup handlers and router
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
	jwtHandlers := handlers.NewJWTHandlers(jwtService)
	// grace period handled inside handler
	accountService := services.NewIdentityService(identityRepo, fileService)
	userHandlers := handlers.NewUserHandlers(accountService)
	router := setupAccountTestRouter(jwtHandlers, userHandlers)

	token, _, err := jwtHandlers.GenerateTokens(userResult.User.ID)
	assert.NoError(t, err, "Should generate tokens")

	t.Run("Cannot reactivate after grace period", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/me/reactivate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Should reject expired account")

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse response")
		assert.Contains(t, response["error"], "Grace period has expired")
		assert.NotEmpty(t, response["deleted_at"])
		assert.NotEmpty(t, response["deadline_was"])
	})
}

// TestAccountAnonymization tests the anonymization process
func TestAccountAnonymization(t *testing.T) {
	_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "1")

	t.Run("Anonymize student account", func(t *testing.T) {
		// Create test student
		userInfo := UserCreationInfo{
			Username:  "anonymize_student_test",
			IsStudent: true,
			IsOAuth:   true,
		}
		userResult, err := CreateUser(userInfo)
		assert.NoError(t, err, "Should create test user")
		defer cleanupUser(t, userResult.User.ID)

		// Set student data
		student := userResult.Student
		student.Phone = "+66812345678"
		student.StudentID = "6410450123"
		student.GitHub = "https://github.com/testuser"
		student.LinkedIn = "https://linkedin.com/in/testuser"
		student.AboutMe = "This is my bio"
		student.Major = "Computer Science"
		db.Save(student)

		// Soft delete with old timestamp (manually)
		oldDeletedAt := time.Now().Add(-2 * 24 * time.Hour)
		db.Unscoped().Model(&userResult.User).Update("deleted_at", oldDeletedAt)

		// Run anonymization
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		identityService := services.NewIdentityService(identityRepo, fileService)
		err = identityService.AnonymizeAccount(context.Background(), userResult.User.ID)
		assert.NoError(t, err, "Should anonymize successfully")

		// Verify user is anonymized
		var user model.User
		db.Unscoped().First(&user, "id = ?", userResult.User.ID)
		assert.True(t, len(user.Username) > 5 && user.Username[:5] == "ANON-", "Username should be anonymized")
		assert.Empty(t, user.PasswordHash, "Password should be cleared")

		// Verify student data is anonymized
		var anonymizedStudent model.Student
		db.Unscoped().First(&anonymizedStudent, "user_id = ?", userResult.User.ID)
		assert.Empty(t, anonymizedStudent.Phone, "Phone should be cleared")
		assert.Empty(t, anonymizedStudent.GitHub, "GitHub should be cleared")
		assert.Empty(t, anonymizedStudent.LinkedIn, "LinkedIn should be cleared")
		assert.Empty(t, anonymizedStudent.AboutMe, "AboutMe should be cleared")
		assert.Equal(t, "Anonymized", anonymizedStudent.Major, "Major should be anonymized")
		assert.True(t, len(anonymizedStudent.StudentID) > 5 && anonymizedStudent.StudentID[:5] == "ANON-", "StudentID should be anonymized")

		// Verify OAuth is anonymized
		var oauth model.GoogleOAuthDetails
		db.Unscoped().First(&oauth, "user_id = ?", userResult.User.ID)
		assert.Equal(t, "Anonymized", oauth.FirstName, "First name should be anonymized")
		assert.Equal(t, "User", oauth.LastName, "Last name should be anonymized")
		assert.Contains(t, oauth.Email, "@anonymized.local", "Email should be anonymized")
	})

	t.Run("Anonymize company account", func(t *testing.T) {
		// Create test company
		userInfo := UserCreationInfo{
			Username:  "anonymize_company_test",
			IsCompany: true,
		}
		userResult, err := CreateUser(userInfo)
		assert.NoError(t, err, "Should create test user")
		defer cleanupUser(t, userResult.User.ID)

		// Set company data
		company := userResult.Company
		company.Email = "contact@company.com"
		company.Phone = "+66812345678"
		company.Website = "https://company.com"
		company.Address = "123 Main St"
		company.City = "Bangkok"
		company.Country = "Thailand"
		company.AboutUs = "Company description"
		db.Save(company)

		// Soft delete with old timestamp
		oldDeletedAt := time.Now().Add(-2 * 24 * time.Hour)
		db.Unscoped().Model(&userResult.User).Update("deleted_at", oldDeletedAt)

		// Run anonymization
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		identityService := services.NewIdentityService(identityRepo, fileService)
		err = identityService.AnonymizeAccount(context.Background(), userResult.User.ID)
		assert.NoError(t, err, "Should anonymize successfully")

		// Verify company data is anonymized
		var anonymizedCompany model.Company
		db.Unscoped().First(&anonymizedCompany, "user_id = ?", userResult.User.ID)
		assert.Contains(t, anonymizedCompany.Email, "@anonymized.local", "Email should be anonymized")
		assert.Empty(t, anonymizedCompany.Phone, "Phone should be cleared")
		assert.Empty(t, anonymizedCompany.Website, "Website should be cleared")
		assert.Empty(t, anonymizedCompany.Address, "Address should be cleared")
		assert.Empty(t, anonymizedCompany.AboutUs, "AboutUs should be cleared")
		assert.Equal(t, "Anonymized", anonymizedCompany.City, "City should be anonymized")
		assert.Equal(t, "Anonymized", anonymizedCompany.Country, "Country should be anonymized")
	})
}

// TestAnonymizeExpiredAccounts tests the scheduled anonymization task
func TestAnonymizeExpiredAccounts(t *testing.T) {
	_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "1")
	gracePeriod := 1

	// Create multiple test users
	users := make([]*UserCreationResult, 3)
	for i := 0; i < 3; i++ {
		userInfo := UserCreationInfo{
			Username:  "batch_anonymize_test_" + string(rune('a'+i)),
			IsStudent: i%2 == 0,
			IsCompany: i%2 == 1,
		}
		var err error
		users[i], err = CreateUser(userInfo)
		assert.NoError(t, err, "Should create test user")
		defer cleanupUser(t, users[i].User.ID)
	}

	// Soft delete users with different timestamps
	// User 0: past grace period (should be anonymized)
	db.Unscoped().Model(&users[0].User).Update("deleted_at", time.Now().Add(-2*24*time.Hour))

	// User 1: past grace period (should be anonymized)
	db.Unscoped().Model(&users[1].User).Update("deleted_at", time.Now().Add(-2*24*time.Hour))

	// User 2: within grace period (should NOT be anonymized)
	db.Unscoped().Model(&users[2].User).Update("deleted_at", time.Now().Add(-12*time.Hour))

	// Run batch anonymization
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	identityService := services.NewIdentityService(identityRepo, fileService)
	err := identityService.AnonymizeExpiredAccounts(context.Background(), gracePeriod)
	assert.NoError(t, err, "Should run anonymization task successfully")

	// Verify users 0 and 1 are anonymized
	for i := 0; i < 2; i++ {
		var user model.User
		db.Unscoped().First(&user, "id = ?", users[i].User.ID)
		assert.True(t, len(user.Username) > 5 && user.Username[:5] == "ANON-",
			"User %d should be anonymized", i)
	}

	// Verify user 2 is NOT anonymized (still within grace period)
	var user2 model.User
	db.Unscoped().First(&user2, "id = ?", users[2].User.ID)
	assert.False(t, len(user2.Username) > 5 && user2.Username[:5] == "ANON-",
		"User 2 should NOT be anonymized yet")
	assert.Equal(t, "batch_anonymize_test_c", user2.Username, "Username should be unchanged")
}

// TestCheckIfAnonymized tests the anonymization check helper
func TestCheckIfAnonymized(t *testing.T) {
	// Create test user
	userInfo := UserCreationInfo{
		Username:  "check_anonymized_test",
		IsStudent: true,
	}
	userResult, err := CreateUser(userInfo)
	assert.NoError(t, err, "Should create test user")
	defer cleanupUser(t, userResult.User.ID)

	t.Run("Check not anonymized", func(t *testing.T) {
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		identityService := services.NewIdentityService(identityRepo, fileService)
		isAnon, err := identityService.CheckIfAnonymized(context.Background(), userResult.User.ID)
		assert.NoError(t, err, "Should check successfully")
		assert.False(t, isAnon, "User should not be anonymized")
	})

	t.Run("Check is anonymized", func(t *testing.T) {
		// Anonymize the user
		db.Unscoped().Model(&userResult.User).Update("deleted_at", time.Now().Add(-2*24*time.Hour))
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		identityService := services.NewIdentityService(identityRepo, fileService)
		err := identityService.AnonymizeAccount(context.Background(), userResult.User.ID)
		assert.NoError(t, err, "Should anonymize successfully")

		isAnon, err := identityService.CheckIfAnonymized(context.Background(), userResult.User.ID)
		assert.NoError(t, err, "Should check successfully")
		assert.True(t, isAnon, "User should be anonymized")
	})
}

// TestGracePeriodConfiguration tests different grace period configurations
func TestGracePeriodConfiguration(t *testing.T) {
	t.Run("Default grace period", func(t *testing.T) {
		_ = os.Unsetenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS")
		gracePeriod := helper.GetGracePeriodDays()
		assert.Equal(t, 30, gracePeriod, "Should default to 30 days")
	})

	t.Run("Custom grace period", func(t *testing.T) {
		_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "60")
		gracePeriod := helper.GetGracePeriodDays()
		assert.Equal(t, 60, gracePeriod, "Should use custom grace period")
	})

	t.Run("Invalid grace period defaults to 30", func(t *testing.T) {
		_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "invalid")
		gracePeriod := helper.GetGracePeriodDays()
		assert.Equal(t, 30, gracePeriod, "Should default to 30 days on invalid input")
	})

	t.Run("Negative grace period defaults to 30", func(t *testing.T) {
		_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "-10")
		gracePeriod := helper.GetGracePeriodDays()
		assert.Equal(t, 30, gracePeriod, "Should default to 30 days on negative input")
	})
}

// TestAnonymizationSkipsAlreadyAnonymized tests that already anonymized accounts are skipped
func TestAnonymizationSkipsAlreadyAnonymized(t *testing.T) {
	_ = os.Setenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS", "1")
	gracePeriod := 1

	// Create and anonymize a user
	userInfo := UserCreationInfo{
		Username:  "skip_anonymized_test",
		IsStudent: true,
	}
	userResult, err := CreateUser(userInfo)
	assert.NoError(t, err, "Should create test user")
	defer cleanupUser(t, userResult.User.ID)

	// Soft delete and anonymize
	db.Unscoped().Model(&userResult.User).Update("deleted_at", time.Now().Add(-2*24*time.Hour))
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	identityService := services.NewIdentityService(identityRepo, fileService)
	err = identityService.AnonymizeAccount(context.Background(), userResult.User.ID)
	assert.NoError(t, err, "First anonymization should succeed")

	// Get the anonymized username
	var user1 model.User
	db.Unscoped().First(&user1, "id = ?", userResult.User.ID)
	username1 := user1.Username

	// Run batch anonymization again - should skip already anonymized account
	identityRepo = gormrepo.NewGormIdentityRepository(db)
	identityService = services.NewIdentityService(identityRepo, fileService)
	err = identityService.AnonymizeExpiredAccounts(context.Background(), gracePeriod)
	assert.NoError(t, err, "Batch anonymization should succeed")

	// Verify username hasn't changed (was skipped)
	var user2 model.User
	db.Unscoped().First(&user2, "id = ?", userResult.User.ID)
	assert.Equal(t, username1, user2.Username, "Username should remain the same (account skipped)")
}

// TestAnonymizationDataRetention tests that non-PII data is retained
func TestAnonymizationDataRetention(t *testing.T) {
	// Create test student
	userInfo := UserCreationInfo{
		Username:  "retention_test_student",
		IsStudent: true,
	}
	userResult, err := CreateUser(userInfo)
	assert.NoError(t, err, "Should create test user")
	defer cleanupUser(t, userResult.User.ID)

	// Store original timestamps
	originalCreatedAt := userResult.User.CreatedAt

	// Anonymize
	db.Unscoped().Model(&userResult.User).Update("deleted_at", time.Now().Add(-2*24*time.Hour))
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	identityService := services.NewIdentityService(identityRepo, fileService)
	err = identityService.AnonymizeAccount(context.Background(), userResult.User.ID)
	assert.NoError(t, err, "Should anonymize successfully")

	// Verify non-PII data is retained
	var user model.User
	db.Unscoped().First(&user, "id = ?", userResult.User.ID)
	assert.Equal(t, userResult.User.ID, user.ID, "User ID should be retained")
	assert.Equal(t, originalCreatedAt.Unix(), user.CreatedAt.Unix(), "CreatedAt should be retained")
	assert.NotNil(t, user.UpdatedAt, "UpdatedAt should exist")

	// Verify deleted_at is still set
	assert.True(t, user.DeletedAt.Valid, "DeletedAt should still be set")
}

// Helper function to cleanup test users
func cleanupUser(t *testing.T, userID string) {
	// Delete all related records using Unscoped
	db.Unscoped().Where("user_id = ?", userID).Delete(&model.Student{})
	db.Unscoped().Where("user_id = ?", userID).Delete(&model.Company{})
	db.Unscoped().Where("user_id = ?", userID).Delete(&model.GoogleOAuthDetails{})
	db.Unscoped().Where("user_id = ?", userID).Delete(&model.Admin{})
	db.Unscoped().Where("id = ?", userID).Delete(&model.User{})
}
