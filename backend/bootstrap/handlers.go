package bootstrap

import (
	"fmt"

	"ku-work/backend/handlers"

	"gorm.io/gorm"
)

// HandlersBundle groups all HTTP handlers constructed from service dependencies.
// This decouples HTTP wiring from service and repository composition.
type HandlersBundle struct {
	// Core auth
	JWT       *handlers.JWTHandlers
	LocalAuth *handlers.LocalAuthHandlers
	// OAuth
	OAuth *handlers.OauthHandlers

	// File
	File *handlers.FileHandlers

	// Domain handlers
	Job         *handlers.JobHandlers
	Application *handlers.ApplicationHandlers
	Student     *handlers.StudentHandler
	Company     *handlers.CompanyHandlers
	User        *handlers.UserHandlers
	Admin       *handlers.AdminHandlers
}

// BuildHandlers constructs all HTTP handlers using the provided DB connection and services bundle.
// The DB connection is required for handlers that need it for read-only checks or associations.
func BuildHandlers(db *gorm.DB, svc *ServicesBundle) (*HandlersBundle, error) {
	if db == nil {
		return nil, fmt.Errorf("db must not be nil")
	}
	if svc == nil {
		return nil, fmt.Errorf("services bundle must not be nil")
	}

	// Validate critical services
	if svc.JWT == nil {
		return nil, fmt.Errorf("jwt service is required")
	}
	if svc.Auth == nil {
		return nil, fmt.Errorf("auth service is required")
	}
	if svc.File == nil {
		return nil, fmt.Errorf("file service is required")
	}
	if svc.Job == nil {
		return nil, fmt.Errorf("job service is required")
	}
	if svc.Application == nil {
		return nil, fmt.Errorf("application service is required")
	}
	if svc.Identity == nil {
		return nil, fmt.Errorf("identity service is required")
	}
	if svc.Student == nil {
		return nil, fmt.Errorf("student service is required")
	}
	if svc.Company == nil {
		return nil, fmt.Errorf("company service is required")
	}
	if svc.Admin == nil {
		return nil, fmt.Errorf("admin service is required")
	}

	// JWT + Auth handlers
	jwtHandlers := handlers.NewJWTHandlers(svc.JWT)
	localAuthHandlers := handlers.NewLocalAuthHandlers(svc.Auth)

	// OAuth handlers (optional based on config)
	var oauthHandlers *handlers.OauthHandlers
	if svc.OAuth != nil {
		oauthHandlers = handlers.NewOAuthHandlers(svc.OAuth, svc.Auth)
	}

	// File handlers - requires DB for some associations and checks
	fileHandlers := handlers.NewFileHandlers(svc.File)

	// Job handlers (uses AI service optionally for auto-approve)
	jobHandlers, err := handlers.NewJobHandlers(svc.AI, svc.Job)
	if err != nil {
		return nil, fmt.Errorf("failed to build job handlers: %w", err)
	}

	// Application handlers
	applicationHandlers := handlers.NewApplicationHandlers(svc.Application)

	// Student, Company, User, Admin handlers
	studentHandlers := handlers.NewStudentHandler(svc.Student, svc.Identity)
	companyHandlers := handlers.NewCompanyHandlers(svc.Company)
	userHandlers := handlers.NewUserHandlers(svc.Identity)
	adminHandlers := handlers.NewAdminHandlers(svc.Admin)

	return &HandlersBundle{
		JWT:         jwtHandlers,
		LocalAuth:   localAuthHandlers,
		OAuth:       oauthHandlers,
		File:        fileHandlers,
		Job:         jobHandlers,
		Application: applicationHandlers,
		Student:     studentHandlers,
		Company:     companyHandlers,
		User:        userHandlers,
		Admin:       adminHandlers,
	}, nil
}
