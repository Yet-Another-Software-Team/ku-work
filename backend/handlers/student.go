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

// Handle Registration to be student
//
// Reject if user already registered to be student.
// use by OAuth users.
//
// use multipart/form-data
func (h *StudentHandler) RegisterHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Check if user is already registered as a student
	var count int64
	h.DB.Model(&model.Student{}).Where("user_id = ?", userId).Count(&count)
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already registered to be student"})
		return
	}

	// Bind request body to input struct
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
	// Parse Birth Date to RFC3339 format
	var parsedBirthDate time.Time
	if input.BirthDate != "" {
		parsedBirthDate, err = time.Parse(time.RFC3339, input.BirthDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	tx := h.DB.Begin()
	defer tx.Rollback() // Rollback transaction at function exit (If successful, commit will be executed before rollback)

	// Save Files
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

	// Create student model based on input data
	student := model.Student{
		UserID:              userId,
		ApprovalStatus:      model.StudentApprovalPending,
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

	// Create file Commit the transaction
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

// Edit student profile
//
// # Support partial updates
//
// use multipart/form-data
func (h *StudentHandler) EditProfileHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data to StudentEditProfileInput struct
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

	// Get current data from database.
	student := model.Student{
		UserID: userId,
	}
	if err := h.DB.First(&student).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update student data according to input, maintain same data if not provided
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

	// Save data into database
	result := h.DB.Save(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// Approve student registration
func (h *StudentHandler) ApproveHandler(ctx *gin.Context) {
	// Bind input data to struct
	type StudentRegistrationApprovalInput struct {
		UserID  string `json:"id" binding:"max=128"`
		Approve bool   `json:"approve"`
	}
	input := StudentRegistrationApprovalInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get student data from database
	tx := h.DB.Begin()
	student := model.Student{
		UserID: input.UserID,
	}
	result := tx.Take(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Accept or Reject student based on `approve` paramter
	if input.Approve {
		student.ApprovalStatus = model.StudentApprovalAccepted
	} else {
		student.ApprovalStatus = model.StudentApprovalRejected
	}

	result = tx.Save(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if err := tx.Create(&model.Audit{
		ActorID:    ctx.MustGet("userID").(string),
		Action:     string(student.ApprovalStatus),
		ObjectName: "Student",
		ObjectID:   student.UserID,
	}).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// Get Student Profile from database
func (h *StudentHandler) GetProfileHandler(ctx *gin.Context) {
	// Get userId from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data from request body
	type GetStudentProfileInput struct {
		UserID         string `form:"id" binding:"max=128"`
		Offset         int    `json:"offset" form:"offset"`
		Limit          int    `json:"limit" form:"limit" binding:"max=64"`
		ApprovalStatus string `json:"approvalStatus" form:"approvalStatus" binding:"max=64"`
	}
	input := GetStudentProfileInput{
		Limit: 64,
	}
	err := ctx.MustBindWith(&input, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := h.DB.Model(&model.Student{})

	// If user ID is provided, use the userId from request
	if input.UserID != "" {
		userId = input.UserID
	} else {
		result := h.DB.Limit(1).Find(&model.Admin{
			UserID: userId,
		})
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		} else if result.RowsAffected != 0 {
			var students []model.Student
			if input.ApprovalStatus != "" {
				query = query.Where(&model.Student{
					ApprovalStatus: model.StudentApprovalStatus(input.ApprovalStatus),
				})
			}
			if err := query.Offset(input.Offset).Limit(input.Limit).Find(&students).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusOK, students)
			return
		}
	}

	// Get Student Profile from database
	type StudentInfo struct {
		model.Student
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
	}
	var studentInfo StudentInfo
	if err := query.Select("students.*, google_o_auth_details.first_name as first_name, google_o_auth_details.last_name as last_name, google_o_auth_details.email as email").Joins("INNER JOIN google_o_auth_details on google_o_auth_details.user_id = students.user_id").Where("students.user_id = ?", userId).Take(&studentInfo).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"profile": studentInfo,
	})
}
