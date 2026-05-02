package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/go-web-services/go-gateway-template/config"

	platformConsts "github.com/go-web-services/go-web-platform/constants"
)

const (
	turnstileVerifyURL     = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	captchaTimeout         = 30 * time.Second
	bypassCaptchaHeaderKey = "X-Bypass-Captcha"
)

// CaptchaResponse represents the response from Cloudflare Turnstile API
type CaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes,omitempty"`
}

// CaptchaRequest represents the request to Cloudflare Turnstile API
type CaptchaRequest struct {
	Secret         string `json:"secret"`
	Response       string `json:"response"`
	IdempotencyKey string `json:"idempotency_key"`
}

// CaptchaMiddleware creates a middleware that verifies Cloudflare Turnstile captcha
func CaptchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we should bypass captcha verification
		// Only bypass if the header is present with value "true" and we're in dev environment
		bypassHeader := c.GetHeader(bypassCaptchaHeaderKey)
		if bypassHeader == "true" && config.Cfg.App.Env == platformConsts.Development {
			// Skip captcha verification in dev environment with bypass header
			c.Next()
			return
		}

		// Get the captcha response from the request body
		var requestBody map[string]any
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Extract the captcha response
		cfTurnstileResponse, ok := requestBody["cf_turnstile_response"].(string)
		if !ok || cfTurnstileResponse == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing captcha response"})
			return
		}

		// Verify the captcha
		if err := verifyCaptcha(cfTurnstileResponse); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		// Store the original body back for subsequent middleware/handlers
		bodyBytes, _ := json.Marshal(requestBody)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Next()
	}
}

// verifyCaptcha verifies the captcha response with Cloudflare Turnstile API
func verifyCaptcha(cfTurnstileResponse string) error {
	idempotencyKey := uuid.New().String()

	// Create the request payload
	captchaReq := CaptchaRequest{
		Secret:         config.Cfg.App.TurnstileSecretKey,
		Response:       cfTurnstileResponse,
		IdempotencyKey: idempotencyKey,
	}

	// Convert the request to JSON
	reqBody, err := json.Marshal(captchaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal captcha request: %w", err)
	}

	// Create an HTTP client with timeout
	client := &http.Client{
		Timeout: captchaTimeout,
	}

	// Send the request to Cloudflare
	resp, err := client.Post(
		turnstileVerifyURL,
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to send captcha verification request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read captcha verification response: %w", err)
	}

	// Parse the response
	var captchaResp CaptchaResponse
	if err := json.Unmarshal(respBody, &captchaResp); err != nil {
		return fmt.Errorf("failed to unmarshal captcha verification response: %w", err)
	}

	// Check if the captcha verification was successful
	if !captchaResp.Success {
		return fmt.Errorf("wrong captcha")
	}

	return nil
}
