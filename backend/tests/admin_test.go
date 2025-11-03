package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ku-work/backend/handlers"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"

	"github.com/stretchr/testify/assert"
)

// TestAdminEndpoints covers admin-only endpoints: GET /admin/audits and GET /company (admin company list).
func TestAdminEndpoints(t *testing.T) {
	// Create an admin user for testing
	adminUsername := fmt.Sprintf("admintest-%d", time.Now().UnixNano())
	adminRes, err := CreateUser(UserCreationInfo{
		Username: adminUsername,
		IsAdmin:  true,
	})
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}
	// Ensure cleanup
	defer func() {
		_ = db.Unscoped().Delete(&adminRes.User)
		if adminRes.Admin != nil {
			_ = db.Unscoped().Delete(adminRes.Admin)
		}
	}()

	// Create non-admin company user to test forbidden access
	companyRes, err := CreateUser(UserCreationInfo{
		Username:  fmt.Sprintf("companytest-%d", time.Now().UnixNano()),
		IsCompany: true,
	})
	if err != nil {
		t.Fatalf("failed to create company user: %v", err)
	}
	defer func() { _ = db.Unscoped().Delete(&companyRes.User) }()

	// Prepare JWT tokens
	userRepo := gormrepo.NewGormUserRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
	jwtHandler := handlers.NewJWTHandlers(jwtService)
	adminToken, _, err := jwtHandler.GenerateTokens(adminRes.User.ID)
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}
	companyToken, _, err := jwtHandler.GenerateTokens(companyRes.User.ID)
	if err != nil {
		t.Fatalf("failed to generate company token: %v", err)
	}

	t.Run("FetchAuditLog_AsAdmin", func(t *testing.T) {
		// Insert some audits
		a1 := model.Audit{
			ActorID:    adminRes.User.ID,
			Action:     "create_company",
			ObjectName: "Company",
			Reason:     "test",
			ObjectID:   companyRes.User.ID,
		}
		a2 := model.Audit{
			ActorID:    "system",
			Action:     "system_event",
			ObjectName: "System",
			Reason:     "heartbeat",
			ObjectID:   "0",
		}
		if err := db.Create(&a1).Error; err != nil {
			t.Fatalf("failed to create audit a1: %v", err)
		}
		if err := db.Create(&a2).Error; err != nil {
			t.Fatalf("failed to create audit a2: %v", err)
		}
		// Cleanup inserted audits
		defer func() {
			_ = db.Unscoped().Delete(&a1)
			_ = db.Unscoped().Delete(&a2)
		}()

		req, _ := http.NewRequest("GET", "/admin/audits", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "admin should be able to fetch audit log")

		var audits []model.Audit
		if err := json.Unmarshal(w.Body.Bytes(), &audits); err != nil {
			t.Fatalf("failed to parse audits response: %v", err)
		}

		// Expect at least the two we inserted exist in the returned list (order not important)
		foundA1, foundA2 := false, false
		for _, a := range audits {
			if a.ObjectID == a1.ObjectID && a.Action == a1.Action {
				foundA1 = true
			}
			if a.ObjectID == a2.ObjectID && a.Action == a2.Action {
				foundA2 = true
			}
		}
		assert.True(t, foundA1, "inserted audit a1 should be present in response")
		assert.True(t, foundA2, "inserted audit a2 should be present in response")
	})

	t.Run("FetchAuditLog_AsNonAdminForbidden", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/admin/audits", nil)
		req.Header.Set("Authorization", "Bearer "+companyToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Non-admin should receive 403 Forbidden
		assert.Equal(t, http.StatusForbidden, w.Code, "non-admin should be forbidden from accessing audit log")
	})

	t.Run("CompanyList_AsAdmin", func(t *testing.T) {
		// Create several companies to ensure list has content
		var companyIDs []string
		for i := 0; i < 3; i++ {
			cu, err := CreateUser(UserCreationInfo{
				Username:  fmt.Sprintf("listcompany-%d-%d", i, time.Now().UnixNano()),
				IsCompany: true,
			})
			if err != nil {
				t.Fatalf("failed to create company %d: %v", i, err)
			}
			companyIDs = append(companyIDs, cu.Company.UserID)
			// cleanup later
			defer func(uID string) {
				_ = db.Unscoped().Where("id = ?", uID).Delete(&model.User{})
			}(cu.User.ID)
		}

		req, _ := http.NewRequest("GET", "/company", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "admin should be able to fetch company list")

		// Response is expected to be a JSON array of companies; we can decode into []map[string]interface{}
		var companies []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &companies); err != nil {
			t.Fatalf("failed to parse company list response: %v", err)
		}

		// Ensure at least the created companies appear in the list (match by userId/company id field)
		found := 0
		for _, comp := range companies {
			if idVal, ok := comp["userId"].(string); ok {
				for _, expected := range companyIDs {
					if idVal == expected {
						found++
					}
				}
			}
		}
		assert.True(t, found >= 1, "expected to find created companies in list (found %d)", found)
	})

	t.Run("CompanyList_AsNonAdminForbidden", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/company", nil)
		req.Header.Set("Authorization", "Bearer "+companyToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Non-admin should not be allowed to list companies (AdminPermissionMiddleware)
		assert.Equal(t, http.StatusForbidden, w.Code, "non-admin should be forbidden from listing companies")
	})
}
