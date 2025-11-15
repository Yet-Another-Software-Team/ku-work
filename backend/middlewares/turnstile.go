package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// TurnstileExceptionMiddleware creates a middleware to tell TurnstileMiddleware to not verify by asking associated handler first
func TurnstileExceptionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("ShouldCF", true)
	}
}

// TurnstileMiddleware creates a middleware to protect endpoints with cloudflare turnstile
func TurnstileMiddleware() gin.HandlerFunc {
	secret, hasSecret := os.LookupEnv("TURNSTILE_SECRET")
	if !hasSecret {
		log.Print("TURNSTILE_SECRET is missing, not verifying turnstile")
	}
	// Create http client to request cloudflare
	client := &http.Client{}
	return func(ctx *gin.Context) {
		// Only post or patch and exception
		if !hasSecret || (ctx.Request.Method != "POST" && ctx.Request.Method != "PATCH") || (ctx.GetBool("ShouldCF") && !ctx.GetBool("DoCF")) {
			delete(ctx.Keys, "ShouldCF")
			ctx.Next()
			return
		}

		// Get the token from HTTP header
		turnstileToken := ctx.GetHeader("X-Turnstile-Token")
		if turnstileToken == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Turnstile Token is missing"})
			return
		}

		// Encode JSON
		type Request struct {
			Secret   string `json:"secret"`
			Response string `json:"response"`
			RemoteIP string `json:"remoteip"`
		}
		request := Request{
			Secret:   secret,
			Response: turnstileToken,
			RemoteIP: ctx.ClientIP(),
		}
		jsonDataBytes, err := json.Marshal(request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create request https://developers.cloudflare.com/turnstile/get-started/server-side-validation/
		req, err := http.NewRequest("POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewReader(jsonDataBytes))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		// Set content type header
		req.Header.Set("Content-Type", "application/json")

		// Execute the request
		resp, err := client.Do(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		defer (func() {
			_ = resp.Body.Close()
		})()

		// Check if the response is successful
		if resp.StatusCode != http.StatusOK {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Unexpected status code: %d\n", resp.StatusCode)})
			return
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		// Parse the JSON response
		type ResultStruct struct {
			Success            bool     `json:"success"`
			ChallengeTimestamp string   `json:"challenge_ts"`
			HostName           string   `json:"hostname"`
			ErrorCodes         []string `json:"error-codes"`
			Action             string   `json:"action"`
			ClientData         string   `json:"cdata"`
		}
		var result ResultStruct
		err = json.Unmarshal(body, &result)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if !result.Success {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}
		ctx.Next()
	}
}
