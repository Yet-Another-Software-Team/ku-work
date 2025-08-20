package middleware

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("")

func InitAuth() error {
	secretJWTKey, hasJWTKey := os.LookupEnv("JWT_SECRET")
	if !hasJWTKey {
		return errors.New("no JWT secret specified")
	}
	jwtKey = []byte(secretJWTKey)
	return nil
}

type Claims struct {
    UserId uint `json:"userid"`
    jwt.RegisteredClaims
}

func Auth(ctx *gin.Context) {
	token, err := ctx.Cookie("Token")
	if err != nil {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
            ctx.Next()
            return
        }
		ctx.Set("userid", claims.UserId)
	}
	ctx.Next()
}