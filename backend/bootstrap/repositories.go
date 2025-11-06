package bootstrap

import (
	"fmt"

	"ku-work/backend/repository"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Repositories groups all data access abstractions used by the application.
// Concrete implementations are wired here (GORM for SQL, Redis for revocation lists).
type Repositories struct {
	// SQL-backed repositories (GORM)
	Identity     repository.IdentityRepository
	RefreshToken repository.RefreshTokenRepository
	File         repository.FileRepository
	Audit        repository.AuditRepository
	Job          repository.JobRepository
	Student      repository.StudentRepository
	Company      repository.CompanyRepository
	Application  repository.ApplicationRepository

	// Redis-backed repository
	Revocation repository.JWTRevocationRepository
}

// NewRepositories constructs all repositories using the provided DB and Redis client.
// - DB is required for all GORM repositories.
// - redisClient is optional; when nil, Revocation will remain nil.
func NewRepositories(db *gorm.DB, redisClient *redis.Client) (*Repositories, error) {
	if db == nil {
		return nil, fmt.Errorf("db must not be nil")
	}

	repos := &Repositories{
		Identity:     gormrepo.NewGormIdentityRepository(db),
		RefreshToken: gormrepo.NewGormRefreshTokenRepository(db),
		File:         gormrepo.NewGormFileRepository(db),
		Audit:        gormrepo.NewGormAuditRepository(db),
		Job:          gormrepo.NewGormJobRepository(db),
		Student:      gormrepo.NewGormStudentRepository(db),
		Company:      gormrepo.NewGormCompanyRepository(db),
		Application:  gormrepo.NewGormApplicationRepository(db),
	}

	if redisClient != nil {
		repos.Revocation = redisrepo.NewRedisRevocationRepository(redisClient)
	}

	return repos, nil
}