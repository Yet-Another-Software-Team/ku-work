package handlers

import (
	"ku-work/backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandlers struct {
	DB *gorm.DB
}

func NewUserHandlers(db *gorm.DB) *UserHandlers {
	return &UserHandlers{
		DB: db,
	}
}

func (h *UserHandlers) GetUsers(ctx *gin.Context) {
	var users []model.User
	result := h.DB.Model(&model.User{}).Find(&users)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.JSON(200, gin.H{
		"users": users,
	})
}

func (h *UserHandlers) CreateUser(ctx *gin.Context) {
	type CreateUserInput struct {
		Username string `json:"user" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	input := CreateUserInput{}
	err := ctx.Bind(&input)
	if err != nil {
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	user := model.User{
		Username: input.Username,
		Password: hashedPassword,
		IsAdmin:  false,
	}
	result := h.DB.Create(&user)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.String(http.StatusOK, "OK ;)")
}

func (h *UserHandlers) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Simple response",
	})
}
