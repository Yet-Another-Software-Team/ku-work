package bootstrap

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"ku-work/backend/providers/ai"
	provideremail "ku-work/backend/providers/email"
	filehandling "ku-work/backend/providers/file_handling"
	"ku-work/backend/services"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// NOTE: gorm import removed because the refactored services no longer directly need *gorm.DB here.

// ServicesBundle groups all application services wired from repositories and infra.
type ServicesBundle struct {
	// Infra services
	Email       *services.EmailService
	File        *services.FileService
	AI          *services.AIService
	EventBus    *services.EventBus
	RateLimiter *services.RateLimiterService
	OAuth       *services.OAuthService

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
func BuildServices(ctx context.Context, _ any, repos *Repositories) (*ServicesBundle, error) {
	// (The first parameter previously was *gorm.DB; kept placeholder to avoid broader signature cascade)
	if repos == nil {
		return nil, fmt.Errorf("repositories bundle must not be nil")
	}

	var (
		err         error
		emailSvc    *services.EmailService
		eventBus    *services.EventBus
		aiSvc       *services.AIService
		aiEngine    ai.ApprovalAI
		jobSvc      *services.JobService
		fileSvc     *services.FileService
		jwtSvc      *services.JWTService
		authSvc     *services.AuthService
		appSvc      *services.ApplicationService
		identity    *services.IdentityService
		student     *services.StudentService
		company     *services.CompanyService
		adminSvc    services.AdminService
		rateLimiter *services.RateLimiterService
		oauthSvc    *services.OAuthService
	)

	// -------------------------
	// Email Service (DI-based)
	// -------------------------
	if repos.Audit != nil {
		emailProviderName := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
		var ep provideremail.EmailProvider
		switch emailProviderName {
		case "smtp":
			if p, e := provideremail.NewSMTPEmailProvider(); e == nil {
				ep = p
			} else {
				log.Printf("SMTP provider init failed: %v", e)
			}
		case "gmail":
			if p, e := provideremail.NewGmailEmailProvider(); e == nil {
				ep = p
			} else {
				log.Printf("Gmail provider init failed: %v", e)
			}
		case "dummy":
			ep = provideremail.NewDummyEmailProvider()
		case "", "none":
			// disabled
		default:
			log.Printf("Unknown EMAIL_PROVIDER '%s' -> email disabled", emailProviderName)
		}
		if ep != nil {
			if emailSvc, err = services.NewEmailService(ep, repos.Audit); err != nil {
				log.Printf("Email service init failed: %v", err)
				emailSvc = nil
			}
		}
	}

	// -------------------------
	// File Service
	// -------------------------
	fileProvider, err := buildFileProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("file provider init failed: %w", err)
	}
	fileSvc = services.NewFileService(repos.File, fileProvider)
	fileSvc.RegisterGlobal()

	// -------------------------
	// JWT + Auth
	// -------------------------
	if repos.RefreshToken == nil || repos.Revocation == nil || repos.Identity == nil {
		return nil, fmt.Errorf("jwt service requires refresh, revocation, and identity repositories")
	}
	jwtSvc = services.NewJWTService(repos.RefreshToken, repos.Revocation, repos.Identity)
	authSvc = services.NewAuthService(jwtSvc, repos.Identity, fileSvc)

	// -------------------------
	// Job Service (initialized after EventBus to ensure non-nil bus)
	// -------------------------
	// Deferred until after EventBus creation; see initialization later.

	// -------------------------
	// Optional AI + EventBus
	// -------------------------
	if emailSvc != nil && repos.Job != nil && repos.Student != nil {
		switch strings.ToLower(strings.TrimSpace(os.Getenv("APPROVAL_AI"))) {
		case "ollama":
			if eng, e := ai.NewOllamaApprovalAI(); e == nil {
				aiEngine = eng
			} else {
				log.Printf("Ollama AI init failed: %v", e)
			}
		case "dummy":
			aiEngine = ai.NewDummyApprovalAI()
		case "", "none":
			// disabled
		default:
			log.Printf("Unknown APPROVAL_AI '%s' -> AI disabled", os.Getenv("APPROVAL_AI"))
		}

		if aiEngine != nil {
			// EventBus (loads templates itself if not provided)
			eventBus, err = services.NewEventBus(services.EventBusOptions{
				AI:          aiEngine,
				JobRepo:     repos.Job,
				StudentRepo: repos.Student,
				Email:       emailSvc,
			})
			if err != nil {
				log.Printf("EventBus init failed: %v", err)
				eventBus = nil
			}

			// AI Service (wraps AI + repos + Bus)
			if eventBus != nil {
				aiSvc, err = services.NewAIService(aiEngine, repos.Job, repos.Student, eventBus)
				if err != nil {
					log.Printf("AI service init failed: %v", err)
					aiSvc = nil
				}

			}
		}
	}

	// -------------------------
	// Initialize Job Service now that EventBus (and possibly AI) are configured
	// -------------------------
	if repos.Job != nil {
		jobSvc = services.NewJobService(repos.Job, eventBus)
	}

	// -------------------------
	// Application Service (email templates optional)
	// -------------------------
	appSvc = services.NewApplicationService(
		repos.Application,
		repos.Job,
		repos.Student,
		repos.Identity,
		fileSvc,
		eventBus,
	)

	// -------------------------
	// Identity Service
	// -------------------------
	identity = services.NewIdentityService(repos.Identity, fileSvc)

	// -------------------------
	// Student Service (EventBus wiring)
	// -------------------------
	student = services.NewStudentService(
		repos.Student,
		repos.Identity,
		fileSvc,
		eventBus,
	)

	// -------------------------
	// Company + Admin
	// -------------------------
	company = services.NewCompanyService(repos.Company)
	adminSvc = services.NewAdminService(repos.Audit)

	// Rate Limiter (repository-backed, fail-open on errors)
	if repos.RateLimit != nil {
		rateLimiter = services.NewRateLimiterService(repos.RateLimit, true)
	}

	// -------------------------
	// OAuth Service (Google) from env config
	// -------------------------
	if strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")) != "" {
		cfg := &oauth2.Config{
			RedirectURL:  "postmessage",
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		}
		oauthSvc = services.NewOAuthService(cfg, nil, "")
	}

	return &ServicesBundle{
		Email:       emailSvc,
		File:        fileSvc,
		AI:          aiSvc,
		EventBus:    eventBus,
		RateLimiter: rateLimiter,
		OAuth:       oauthSvc,
		JWT:         jwtSvc,
		Auth:        authSvc,
		Job:         jobSvc,
		Application: appSvc,
		Identity:    identity,
		Student:     student,
		Company:     company,
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
