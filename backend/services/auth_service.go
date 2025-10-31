package services

import (
	"errors"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"mime/multipart"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthService holds auth-related business logic.
// TokenProvider defines the minimal interface required by AuthService to generate tokens.
// This allows decoupling AuthService from the concrete HTTP/handler layer implementation.
type TokenProvider interface {
	HandleToken(user model.User) (string, string, error)
}

// SaveFileFunc defines the signature for saving uploaded files.
// This is injected into AuthService to avoid importing handlers from services
// and to break import cycles.
type SaveFileFunc func(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error)

// AuthService holds auth-related business logic.
type AuthService struct {
	DB            *gorm.DB
	TokenProvider TokenProvider
	SaveFile      SaveFileFunc
}

// NewAuthService constructs an AuthService with injected dependencies.
// The saveFile parameter allows services to call into file-saving logic without
// depending on the handlers package.
func NewAuthService(db *gorm.DB, provider TokenProvider, saveFile SaveFileFunc) *AuthService {
	return &AuthService{DB: db, TokenProvider: provider, SaveFile: saveFile}
}

// RegisterCompany performs
type RegisterCompanyInput struct {
	Username string
	Password string
	Email    string
	Website  string
	Phone    string
	Address  string
	City     string
	Country  string
	Photo    *multipart.FileHeader
	Banner   *multipart.FileHeader
	AboutUs  string
}

func (s *AuthService) RegisterCompany(ctx *gin.Context, input RegisterCompanyInput) (model.User, model.Company, string, string, error) {
	var zeroUser model.User
	var zeroCompany model.Company

	// check existence
	var count int64
	if err := s.DB.Model(&model.User{}).
		Where("username = ? AND user_type = ?", input.Username, "company").
		Count(&count).Error; err != nil {
		return zeroUser, zeroCompany, "", "", err
	}
	if count > 0 {
		return zeroUser, zeroCompany, "", "", ErrUsernameExists
	}

	hashedPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return zeroUser, zeroCompany, "", "", tx.Error
	}
	defer func() {
		// rollback will be ignored if commit already happened
		_ = tx.Rollback()
	}()

	newUser := model.User{
		Username:     input.Username,
		UserType:     "company",
		PasswordHash: hashedPassword,
	}

	if err := tx.Create(&newUser).Error; err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	photo, err := s.SaveFile(ctx, tx, newUser.ID, input.Photo, model.FileCategoryImage)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	banner, err := s.SaveFile(ctx, tx, newUser.ID, input.Banner, model.FileCategoryImage)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	if input.Website != "" {
		parsed, err := url.Parse(input.Website)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return zeroUser, zeroCompany, "", "", ErrInvalidWebsite
		}
	}

	newCompany := model.Company{
		UserID:   newUser.ID,
		Email:    input.Email,
		Phone:    input.Phone,
		PhotoID:  photo.ID,
		BannerID: banner.ID,
		Address:  input.Address,
		City:     input.City,
		AboutUs:  input.AboutUs,
		Country:  input.Country,
		Website:  input.Website,
	}

	if err := tx.Create(&newCompany).Error; err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	jwtToken, refreshToken, err := s.TokenProvider.HandleToken(newUser)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	return newUser, newCompany, jwtToken, refreshToken, nil
}

// CompanyLogin validates credentials for a company and returns tokens on success.
func (s *AuthService) CompanyLogin(username, password string) (model.User, string, string, error) {
	var user model.User
	if err := s.DB.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, "company").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, "", "", ErrInvalidCredentials
		}
		return model.User{}, "", "", err
	}

	match, err := helper.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	var companyCount int64
	if err := s.DB.Model(&model.Company{}).Where("user_id = ?", user.ID).Count(&companyCount).Error; err != nil {
		return model.User{}, "", "", err
	}
	if companyCount == 0 {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	jwtToken, refreshToken, err := s.TokenProvider.HandleToken(user)
	if err != nil {
		return model.User{}, "", "", err
	}

	return user, jwtToken, refreshToken, nil
}

// AdminLogin validates admin credentials and returns tokens on success.
func (s *AuthService) AdminLogin(username, password string) (model.User, string, string, error) {
	var user model.User
	if err := s.DB.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, "admin").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, "", "", ErrInvalidCredentials
		}
		return model.User{}, "", "", err
	}

	match, err := helper.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	var adminCount int64
	if err := s.DB.Model(&model.Admin{}).Where("user_id = ?", user.ID).Count(&adminCount).Error; err != nil {
		return model.User{}, "", "", err
	}
	if adminCount == 0 {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	jwtToken, refreshToken, err := s.TokenProvider.HandleToken(user)
	if err != nil {
		return model.User{}, "", "", err
	}

	return user, jwtToken, refreshToken, nil
}

func (s *AuthService) HandleGoogleOAuth(userInfo struct {
	ID         string
	Email      string
	Name       string
	GivenName  string
	FamilyName string
}) (jwtToken string, refreshToken string, username string, role string, userId string, isRegistered bool, statusCode int, err error) {
	// default return values
	jwtToken = ""
	refreshToken = ""
	username = ""
	// helper.Role is a named string type; convert to string for this function's return type.
	role = string(helper.Viewer)
	userId = ""
	isRegistered = false
	statusCode = 200

	var userCount int64
	if err = s.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", userInfo.ID).Count(&userCount).Error; err != nil {
		return
	}

	var oauthDetail model.GoogleOAuthDetails
	if userCount == 0 {
		var newUser model.User
		if err = s.DB.FirstOrCreate(&newUser, model.User{
			Username: userInfo.Email,
			UserType: "oauth",
		}).Error; err != nil {
			return
		}

		oauthDetail = model.GoogleOAuthDetails{
			UserID:     newUser.ID,
			ExternalID: userInfo.ID,
			FirstName:  userInfo.GivenName,
			LastName:   userInfo.FamilyName,
			Email:      userInfo.Email,
		}
		if err = s.DB.Create(&oauthDetail).Error; err != nil {
			return
		}
		statusCode = 201
	}

	// update details
	if oauthDetail.UserID != "" {
		_ = s.DB.Model(&oauthDetail).Updates(model.GoogleOAuthDetails{
			FirstName: userInfo.GivenName,
			LastName:  userInfo.FamilyName,
			Email:     userInfo.Email,
		}).Error
	}

	if err = s.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", userInfo.ID).First(&oauthDetail).Error; err != nil {
		return
	}

	var user model.User
	if err = s.DB.Model(&user).Where("id = ?", oauthDetail.UserID).First(&user).Error; err != nil {
		return
	}

	jwtToken, refreshToken, err = s.TokenProvider.HandleToken(user)
	if err != nil {
		return
	}

	username = oauthDetail.FirstName + " " + oauthDetail.LastName
	userId = user.ID

	if statusCode == 200 {
		var r string
		var reg bool
		reg, r, err = isStudentRegisteredAndRole(s.DB, user)
		if err != nil {
			return
		}
		role = r
		isRegistered = reg
	}

	return
}
