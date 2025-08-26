package handlers

import (
	"mime/multipart"
	"net/http"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ku-work/backend/model"
)

type StudentHandler struct {
	DB *gorm.DB
}

func NewStudentHandler(db *gorm.DB) *StudentHandler {
	return &StudentHandler{
		DB: db,
	}
}

func handleFile(ctx *gin.Context, file *multipart.FileHeader, directoryName string, fileName string) (string, error) {
	path := fmt.Sprintf("./files/%s/%s", directoryName, fileName)
	if err := ctx.SaveUploadedFile(file, path); err != nil {
		return "", err
	}
	return path[1:], nil
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
	photoPath, err := handleFile(ctx, input.Photo, "student_photo", userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusPhotoPath, err := handleFile(ctx, input.StudentStatusFile, "status_photo", userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	student := model.Student{
		UserID:            userId,
		Approved:          false,
		Phone:             input.Phone,
		Photo:             photoPath,
		BirthDate:         datatypes.Date(parsedBirthDate),
		AboutMe:           input.AboutMe,
		GitHub:            input.GitHub,
		LinkedIn:          input.LinkedIn,
		StudentID:         input.StudentID,
		Major:             input.Major,
		StudentStatus:     input.StudentStatus,
		StudentStatusFile: statusPhotoPath,
	}
	result := h.DB.Create(&student)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *StudentHandler) EditProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type StudentEditProfileInput struct {
		Phone         string                `form:"phone" binding:"max=20"`
		BirthDate     string                `form:"birthDate" binding:"max=27"`
		AboutMe       string                `form:"aboutMe" binding:"max=16384"`
		GitHub        string                `form:"github" binding:"max=256"`
		LinkedIn      string                `form:"linkedIn" binding:"max=256"`
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
	if input.BirthDate != "" {
		parsedBirthDate, err := time.Parse(time.RFC3339, input.BirthDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		student.BirthDate = datatypes.Date(parsedBirthDate)
	}
	if input.Phone != "" {
		student.Phone = input.Phone
	}
	if input.AboutMe != "" {
		student.AboutMe = input.AboutMe
	}
	if input.GitHub != "" {
		student.GitHub = input.GitHub
	}
	if input.LinkedIn != "" {
		student.LinkedIn = input.LinkedIn
	}
	if input.StudentStatus != "" {
		student.StudentStatus = input.StudentStatus
	}
	if input.Photo != nil {
		photoPath, err := handleFile(ctx, input.Photo, "student_photo", userId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		student.Photo = photoPath
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
	student := model.Student{
		UserID: userId,
	}
	if err := h.DB.First(&student).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"profile": student,
	})
}
