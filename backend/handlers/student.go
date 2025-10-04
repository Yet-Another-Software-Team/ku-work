package handlers

import (
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ku-work/backend/model"
)

type StudentHandler struct {
	DB           *gorm.DB
	fileHandlers *FileHandlers
}

func NewStudentHandler(db *gorm.DB, fileHandlers *FileHandlers) *StudentHandler {
	return &StudentHandler{
		DB:           db,
		fileHandlers: fileHandlers,
	}
}

func (h *StudentHandler) RegisterHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type StudentRegistrationInput struct {
		Phone             string                `form:"phone" binding:"max=20"`
		BirthDate         string                `form:"birthDate" binding:"max=27"`
		AboutMe           string                `form:"aboutMe" binding:"max=16384"`
		GitHub            string                `form:"github" binding:"max=256"`
		LinkedIn          string                `form:"linkedIn" binding:"max=256"`
		StudentID         string                `form:"studentId" binding:"required,len=10"`
		Major             string                `form:"major" binding:"required,oneof='Software and Knowledge Engineering' 'Computer Engineering'"`
		StudentStatus     string                `form:"studentStatus" binding:"required,oneof='Graduated' 'Current Student'"`
		Photo             *multipart.FileHeader `form:"photo" binding:"required"`
		StudentStatusFile *multipart.FileHeader `form:"statusPhoto" binding:"required"`
	}

	var count int64
	h.DB.Model(&model.Student{}).Where("user_id = ?", userId).Count(&count)
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already registered to be student"})
		return
	}

	input := StudentRegistrationInput{}
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var parsedBirthDate time.Time
	if input.BirthDate != "" {
		parsedBirthDate, err = time.Parse(time.RFC3339, input.BirthDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	tx := h.DB.Begin()
	defer tx.Rollback()

	photo, err := SaveFile(ctx, tx, userId, input.Photo, model.FileCategoryImage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusDocument, err := SaveFile(ctx, tx, userId, input.StudentStatusFile, model.FileCategoryDocument)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	student := model.Student{
		UserID:              userId,
		Approved:            false,
		Phone:               input.Phone,
		PhotoID:             photo.ID,
		BirthDate:           datatypes.Date(parsedBirthDate),
		AboutMe:             input.AboutMe,
		GitHub:              input.GitHub,
		LinkedIn:            input.LinkedIn,
		StudentID:           input.StudentID,
		Major:               input.Major,
		StudentStatus:       input.StudentStatus,
		StudentStatusFileID: statusDocument.ID,
	}
	result := tx.Create(&student)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *StudentHandler) EditProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type StudentEditProfileInput struct {
		Phone         *string               `form:"phone" binding:"omitempty,max=20"`
		BirthDate     *string               `form:"birthDate" binding:"omitempty,max=27"`
		AboutMe       *string               `form:"aboutMe" binding:"omitempty,max=16384"`
		GitHub        *string               `form:"github" binding:"omitempty,max=256"`
		LinkedIn      *string               `form:"linkedIn" binding:"omitempty,max=256"`
		StudentStatus string                `form:"studentStatus" binding:"required,oneof='Graduated' 'Current Student'"`
		Photo         *multipart.FileHeader `form:"photo"`
	}
	input := StudentEditProfileInput{}
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	student := model.Student{
		UserID: userId,
	}
	if err := h.DB.First(&student).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if input.BirthDate != nil {
		parsedBirthDate, err := time.Parse(time.RFC3339, *input.BirthDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		student.BirthDate = datatypes.Date(parsedBirthDate)
	}
	if input.Phone != nil {
		student.Phone = *input.Phone
	}
	if input.AboutMe != nil {
		student.AboutMe = *input.AboutMe
	}
	if input.GitHub != nil {
		student.GitHub = *input.GitHub
	}
	if input.LinkedIn != nil {
		student.LinkedIn = *input.LinkedIn
	}
	if input.StudentStatus != "" {
		student.StudentStatus = input.StudentStatus
	}
	if input.Photo != nil {
		photo, err := SaveFile(ctx, h.DB, userId, input.Photo, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		student.PhotoID = photo.ID
	}
	result := h.DB.Save(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *StudentHandler) ApproveHandler(ctx *gin.Context) {
	type StudentRegistrationApprovalInput struct {
		UserID string `json:"id" binding:"max=128"`
	}
	input := StudentRegistrationApprovalInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	student := model.Student{
		UserID: input.UserID,
	}
	result := h.DB.First(&student)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	student.Approved = true
	result = h.DB.Save(&student)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *StudentHandler) GetProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type GetStudentProfileInput struct {
		UserID string `form:"id" binding:"max=128"`
	}
	input := GetStudentProfileInput{}
	err := ctx.MustBindWith(&input, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.UserID != "" {
		userId = input.UserID
	}
	type StudentInfo struct {
		model.Student
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
	}
	var studentInfo StudentInfo
	if err := h.DB.Model(&model.Student{}).Select("students.*, google_o_auth_details.first_name as first_name, google_o_auth_details.last_name as last_name, google_o_auth_details.email as email").Joins("INNER JOIN google_o_auth_details on google_o_auth_details.user_id = students.user_id").Where("students.user_id = ?", userId).Take(&studentInfo).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"profile": studentInfo,
	})
}
