package services

import (
	"errors"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"gorm.io/gorm"
)

var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidWebsite     = errors.New("invalid website url")
)


// CompanyService provides read operations for company data.
type CompanyService struct {
	DB *gorm.DB
}

func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{DB: db}
}

func (s *CompanyService) IsAdmin(userID string) bool {
	role := helper.GetRole(userID, s.DB)
	return role == helper.Admin
}

// CompanyResponse mirrors the public response shape for company profiles.
type CompanyResponse struct {
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   string    `json:"photoId"`
	BannerID  string    `json:"bannerId"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Website   string    `json:"website"`
	AboutUs   string    `json:"about"`
	Name      string    `json:"name"`
}

// GetCompanyByUserID fetches company profile and the owner's username.
func (s *CompanyService) GetCompanyByUserID(userID string) (CompanyResponse, error) {
	var company struct {
		model.Company
		Name string `json:"name"`
	}
	if err := s.DB.Model(&model.Company{}).
		Select("companies.*, users.username as name").
		Joins("INNER JOIN users on users.id = companies.user_id").
		Where("companies.user_id = ?", userID).
		Take(&company).Error; err != nil {
		return CompanyResponse{}, err
	}

	resp := CompanyResponse{
		// model.Company uses UserID (string) as primary key.
		CreatedAt: company.CreatedAt,
		UserID:    company.UserID,
		Email:     company.Email,
		Phone:     company.Phone,
		PhotoID:   company.PhotoID,
		BannerID:  company.BannerID,
		Address:   company.Address,
		City:      company.City,
		Country:   company.Country,
		Website:   company.Website,
		AboutUs:   company.AboutUs,
		Name:      company.Name,
	}
	return resp, nil
}

// ListCompanies returns all companies with owner username populated.
func (s *CompanyService) ListCompanies() ([]CompanyResponse, error) {
	var companies []CompanyResponse
	if err := s.DB.Model(&model.Company{}).
		Select("companies.created_at, companies.updated_at, companies.user_id, companies.email, companies.phone, companies.photo_id, companies.banner_id, companies.address, companies.city, companies.country, companies.website, companies.about_us, users.username as name").
		Joins("INNER JOIN users on users.id = companies.user_id").
		Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

// Helper: check if student is registered and approved (used by oauth flow).
// Returns (isRegistered, role).
func isStudentRegisteredAndRole(db *gorm.DB, user model.User) (bool, string, error) {
	var count int64
	if err := db.Model(&model.Student{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		// helper.Role is a named type; convert to string when returning from this function.
		return false, string(helper.Viewer), err
	}
	if count == 0 {
		return false, string(helper.Viewer), nil
	}
	var student model.Student
	if err := db.Model(&student).Where("user_id = ?", user.ID).First(&student).Error; err != nil {
		return true, string(helper.Viewer), err
	}
	if student.ApprovalStatus == model.StudentApprovalAccepted {
		return true, string(helper.Student), nil
	}
	return true, string(helper.Viewer), nil
}

// small helper to compute cookie max age in seconds
func CookieMaxAge() int {
	return int(time.Hour * 24 * 30 / time.Second)
}
