package bootstrap

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	filehandling "ku-work/backend/providers/file_handling"
	"ku-work/backend/services"

	"gorm.io/gorm"
)

// ServicesBundle groups all application services wired from repositories and infra.
type ServicesBundle struct {
	// Infra services
	Email *services.EmailService
	File  *services.FileService
	AI    *services.AIService

	// Core business services
	JWT         *services.JWTService
	Auth        *services.AuthService
	Job         *services.JobService
	Application *services.ApplicationService
	Identity    *services.IdentityService
	Student     *services.StudentService
	Company     *services.CompanyService
	Admin       services.AdminService
}

// BuildServices constructs and wires all services from repositories and environment configuration.
// - db is used only for services that still depend on GORM directly (e.g., AIService).
// - repos must be fully initialized (SQL + Redis where applicable).
// - ctx is used for provider initializations that need context (e.g., GCS).
func BuildServices(ctx context.Context, db *gorm.DB, repos *Repositories) (*ServicesBundle, error) {
	if db == nil {
		return nil, fmt.Errorf("db must not be nil")
	}
	if repos == nil {
		return nil, fmt.Errorf("repositories bundle must not be nil")
	}

	var err error

	// Email service (optional; depends on EMAIL_PROVIDER)
	var emailSvc *services.EmailService
	if repos.Audit != nil {
		if emailSvc, err = services.NewEmailService(repos.Audit); err != nil {
			log.Printf("Email service not configured or failed to init: %v", err)
			emailSvc = nil
		}
	}

	// File service with provider selection (local | gcs)
	fileProvider, err := buildFileProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("file provider init failed: %w", err)
	}
	fileSvc := services.NewFileService(repos.File, fileProvider)
	// Register storage provider for model-level hooks
	fileSvc.RegisterGlobal()

	// JWT service (requires: refresh repo, revocation repo, identity repo)
	if repos.RefreshToken == nil || repos.Revocation == nil || repos.Identity == nil {
		return nil, fmt.Errorf("jwt service requires refresh, revocation, and identity repositories")
	}
	jwtSvc := services.NewJWTService(repos.RefreshToken, repos.Revocation, repos.Identity)

	// Auth service uses a token provider; JWT service implements HandleToken
	authSvc := services.NewAuthService(jwtSvc, repos.Identity, *fileSvc)

	// Job service (+ optional email template)
	jobSvc := services.NewJobService(repos.Job)
	if emailSvc != nil {
		if tpl := parseTemplateIfExists("email_templates/job_approval_status_update.tmpl", "job_approval_status_update.tmpl"); tpl != nil {
			jobSvc.SetEmailConfig(emailSvc, tpl)
		}
	}

	// AI service (optional)
	var aiSvc *services.AIService
	if emailSvc != nil {
		if aiSvc, err = services.NewAIService(db, emailSvc, jobSvc); err != nil {
			log.Printf("AI service not configured or failed to init: %v", err)
			aiSvc = nil
		}
	}

	// Application service (+ optional email templates)
	var appStatusTpl, appNewApplicantTpl *template.Template
	appStatusTpl = parseTemplateIfExists("templates/emails/job_application_status_update.tmpl", "job_application_status_update.tmpl")
	appNewApplicantTpl = parseTemplateIfExists("templates/emails/job_new_applicant.tmpl", "job_new_applicant.tmpl")
	appSvc := services.NewApplicationService(
		repos.Application,
		repos.Job,
		repos.Student,
		repos.Identity,
		fileSvc,
		emailSvc,
		appStatusTpl,
		appNewApplicantTpl,
	)

	// Identity service
	identitySvc := services.NewIdentityService(repos.Identity, fileSvc)

	// Student service (+ optional email template, AI)
	studentApprovalTpl := parseTemplateIfExists("templates/emails/student_approval_status_update.tmpl", "student_approval_status_update.tmpl")
	studentSvc := services.NewStudentService(
		repos.Student,
		repos.Identity,
		fileSvc,
		emailSvc,
		studentApprovalTpl,
		aiSvc,
	)

	// Company service
	companySvc := services.NewCompanyService(repos.Company)

	// Admin service
	adminSvc := services.NewAdminService(repos.Audit)

	return &ServicesBundle{
		Email:       emailSvc,
		File:        fileSvc,
		AI:          aiSvc,
		JWT:         jwtSvc,
		Auth:        authSvc,
		Job:         jobSvc,
		Application: appSvc,
		Identity:    identitySvc,
		Student:     studentSvc,
		Company:     companySvc,
		Admin:       adminSvc,
	}, nil
}

// buildFileProvider selects and initializes the file handling provider from environment variables:
// - FILE_PROVIDER: "local" (default) or "gcs"
// - LOCAL_FILES_DIR: base directory for local provider (default: ./files)
// - GCS_BUCKET, GCS_CREDENTIALS_PATH: for gcs provider
func buildFileProvider(ctx context.Context) (filehandling.FileHandlingProvider, error) {
	providerType := strings.ToLower(strings.TrimSpace(os.Getenv("FILE_PROVIDER")))
	switch providerType {
	case "", "local":
		baseDir := strings.TrimSpace(os.Getenv("LOCAL_FILES_DIR"))
		if baseDir == "" {
			baseDir = "./files"
		}
		// No need to create directory here; local provider will handle write path and error out if missing.
		return filehandling.NewLocalProvider(baseDir), nil
	case "gcs":
		bucket := strings.TrimSpace(os.Getenv("GCS_BUCKET"))
		if bucket == "" {
			return nil, fmt.Errorf("GCS_BUCKET is required for gcs provider")
		}
		creds := os.Getenv("GCS_CREDENTIALS_PATH")
		return filehandling.NewGCSProvider(ctx, bucket, creds)
	default:
		return nil, fmt.Errorf("unsupported FILE_PROVIDER: %s", providerType)
	}
}

// parseTemplateIfExists attempts to parse a template file if it exists.
// Returns nil when the file does not exist or fails to parse.
func parseTemplateIfExists(path string, name string) *template.Template {
	if _, err := os.Stat(path); err == nil {
		tpl, err := template.New(name).ParseFiles(path)
		if err != nil {
			log.Printf("Failed to parse template %s: %v", path, err)
			return nil
		}
		return tpl
	}
	return nil
}
