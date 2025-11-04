package services

import (
	"errors"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"mime/multipart"
	"net/url"

	repo "ku-work/backend/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthService holds auth-related business logic.
// TokenProvider defines the minimal interface required by AuthService to generate tokens.
// This allows decoupling AuthService from the concrete HTTP/handler layer implementation.
type TokenProvider interface {
	HandleToken(user model.User) (string, string, error)
}

// AuthService holds auth-related business logic.
type AuthService struct {
	tokenProvider TokenProvider
	userRepo      repo.UserRepository
	fileService   FileService
}

// NewAuthService constructs an AuthService with injected dependencies.
// The saveFile parameter allows services to call into file-saving logic without
// depending on the handlers package.
func NewAuthService(provider TokenProvider, userRepo repo.UserRepository, fileService FileService) *AuthService {
	return &AuthService{tokenProvider: provider, userRepo: userRepo, fileService: fileService}
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

	// check existence via repository (non-transactional check)
	exists, err := s.userRepo.ExistsByUsernameAndType(input.Username, "company")
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}
	if exists {
		return zeroUser, zeroCompany, "", "", ErrUsernameExists
	}

	hashedPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}


	newUser := model.User{
		Username:     input.Username,
		UserType:     "company",
		PasswordHash: hashedPassword,
	}

	// Use a repository instance bound to the transaction for all repo calls within the tx.
	repoTx, err := s.userRepo.BeginTx()

	if err := repoTx.CreateUser(&newUser); err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	photo, err := s.fileService.SaveFile(ctx, newUser.ID, input.Photo, model.FileCategoryImage)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	banner, err := s.fileService.SaveFile(ctx, newUser.ID, input.Banner, model.FileCategoryImage)
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

	if err := repoTx.CreateCompany(&newCompany); err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	if err := repoTx.CommitTx(); err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	jwtToken, refreshToken, err := s.tokenProvider.HandleToken(newUser)
	if err != nil {
		return zeroUser, zeroCompany, "", "", err
	}

	return newUser, newCompany, jwtToken, refreshToken, nil
}

// CompanyLogin validates credentials for a company and returns tokens on success.
func (s *AuthService) CompanyLogin(username, password string) (model.User, string, string, error) {
	var user model.User
	userPtr, err := s.userRepo.FindUserByUsernameAndType(username, "company")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, "", "", ErrInvalidCredentials
		}
		return model.User{}, "", "", err
	}
	user = *userPtr

	match, err := helper.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	companyCount, err := s.userRepo.CountCompanyByUserID(user.ID)
	if err != nil {
		return model.User{}, "", "", err
	}
	if companyCount == 0 {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	jwtToken, refreshToken, err := s.tokenProvider.HandleToken(user)
	if err != nil {
		return model.User{}, "", "", err
	}

	return user, jwtToken, refreshToken, nil
}

// AdminLogin validates admin credentials and returns tokens on success.
func (s *AuthService) AdminLogin(username, password string) (model.User, string, string, error) {
	var user model.User
	userPtr, err := s.userRepo.FindUserByUsernameAndType(username, "admin")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, "", "", ErrInvalidCredentials
		}
		return model.User{}, "", "", err
	}
	user = *userPtr

	match, err := helper.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	adminCount, err := s.userRepo.CountAdminByUserID(user.ID)
	if err != nil {
		return model.User{}, "", "", err
	}
	if adminCount == 0 {
		return model.User{}, "", "", ErrInvalidCredentials
	}

	jwtToken, refreshToken, err := s.tokenProvider.HandleToken(user)
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

	// Try to fetch existing OAuth details via repository
	var oauthDetail model.GoogleOAuthDetails
	det, err := s.userRepo.GetGoogleOAuthDetailsByExternalID(userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create user if not exists and attach oauth details
			var newUser model.User
			if err = s.userRepo.FirstOrCreateUser(&newUser, model.User{
				Username: userInfo.Email,
				UserType: "oauth",
			}); err != nil {
				return
			}

			oauthDetail = model.GoogleOAuthDetails{
				UserID:     newUser.ID,
				ExternalID: userInfo.ID,
				FirstName:  userInfo.GivenName,
				LastName:   userInfo.FamilyName,
				Email:      userInfo.Email,
			}
			if err = s.userRepo.CreateGoogleOAuthDetails(&oauthDetail); err != nil {
				return
			}
			statusCode = 201
		} else {
			return
		}
	} else {
		oauthDetail = *det
	}

	// update details
	if oauthDetail.UserID != "" {
		_ = s.userRepo.UpdateGoogleOAuthDetails(&model.GoogleOAuthDetails{
			UserID:    oauthDetail.UserID,
			FirstName: userInfo.GivenName,
			LastName:  userInfo.FamilyName,
			Email:     userInfo.Email,
		})
	}

	det2, err := s.userRepo.GetGoogleOAuthDetailsByExternalID(userInfo.ID)
	if err != nil {
		return
	}
	oauthDetail = *det2

	var user model.User
	userPtr2, err := s.userRepo.FindUserByID(oauthDetail.UserID)
	if err != nil {
		return
	}
	user = *userPtr2

	jwtToken, refreshToken, err = s.tokenProvider.HandleToken(user)
	if err != nil {
		return
	}

	username = oauthDetail.FirstName + " " + oauthDetail.LastName
	userId = user.ID

	if statusCode == 200 {
		var r string
		var reg bool
		reg, r, err = s.userRepo.IsStudentRegisteredAndRole(user)
		if err != nil {
			return
		}
		role = r
		isRegistered = reg
	}

	return
}
