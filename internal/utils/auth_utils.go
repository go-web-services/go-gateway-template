package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	authClientConstants "github.com/go-web-services/go-service-user/pkg/client/constants"
	platformError "github.com/go-web-services/go-web-platform/error"

	"github.com/go-web-services/go-gateway-template/config"
	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
)

// IsTokenExpiredError checks if the error indicates an expired token
func IsTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}

	var baseErr *platformError.BaseError
	if errors.As(err, &baseErr) {
		return string(baseErr.Code) == string(authClientConstants.AuthTokenExpired) ||
			string(baseErr.Code) == string(authClientConstants.SessionExpired)
	}

	return false
}

// IsAuthenticationError checks if the error is an authentication-related error (401)
func IsAuthenticationError(err error) bool {
	if err == nil {
		return false
	}

	var baseErr *platformError.BaseError
	if errors.As(err, &baseErr) {
		return string(baseErr.Code) == string(authClientConstants.AuthTokenExpired) ||
			string(baseErr.Code) == string(authClientConstants.SessionExpired) ||
			string(baseErr.Code) == string(authClientConstants.InvalidAuthToken) ||
			string(baseErr.Code) == string(authClientConstants.InvalidAuthTokenClaims) ||
			string(baseErr.Code) == string(authClientConstants.FingerprintMismatch) ||
			string(baseErr.Code) == string(authClientConstants.UserNotActive)
	}

	return false
}

// ExtractAuthToken extracts the JWT token from the Authorization header
// Returns the token and an error if extraction fails
func ExtractAuthToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(authConstants.AuthAuthorizationHeaderKey)
	if authHeader == "" {
		return "", platformError.ErrUnauthorized
	}

	// Parse Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != authConstants.AuthBearerPrefix {
		return "", platformError.ErrUnauthorized
	}

	return parts[1], nil
}

// ExtractFingerprint extracts the fingerprint from the cookie
// Returns the fingerprint and an error if extraction fails
func ExtractFingerprint(c *gin.Context) (string, error) {
	fingerprint, err := c.Cookie(authConstants.AuthFingerprintCookieName)
	if err != nil {
		return "", platformError.ErrUnauthorized
	}

	return fingerprint, nil
}

// SetFingerprintCookie sets the fingerprint as a secure, httpOnly cookie
func SetFingerprintCookie(c *gin.Context, fingerprint string) {
	c.SetCookie(
		authConstants.AuthFingerprintCookieName,
		fingerprint,
		config.Cfg.App.Auth.FingerprintCookieExpirationSec,
		"/",
		config.Cfg.App.Auth.FingerprintCookieDomain,
		true, // secure
		true, // httpOnly
	)
}

// DeleteFingerprintCookie deletes the fingerprint cookie
func DeleteFingerprintCookie(c *gin.Context) {
	c.SetCookie(
		authConstants.AuthFingerprintCookieName,
		"",
		-1, // MaxAge -1 deletes the cookie
		"/",
		config.Cfg.App.Auth.FingerprintCookieDomain,
		true, // secure
		true, // httpOnly
	)
}

// GenerateCSRFToken generates a random CSRF token
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetGoogleSSOStateCookie sets the Google SSO state token as a secure, httpOnly cookie
// This cookie expires in 10 minutes (600 seconds) to match the typical OAuth flow timeout
func SetGoogleSSOStateCookie(c *gin.Context, stateToken string) {
	c.SetCookie(
		authConstants.AuthGoogleSSOStateCookieName,
		stateToken,
		600, // 10 minutes
		"/",
		config.Cfg.App.Auth.FingerprintCookieDomain,
		true, // secure
		true, // httpOnly
	)
}

// GetGoogleSSOStateCookie retrieves the Google SSO state token from cookie
func GetGoogleSSOStateCookie(c *gin.Context) (string, error) {
	stateToken, err := c.Cookie(authConstants.AuthGoogleSSOStateCookieName)
	if err != nil {
		return "", platformError.ErrUnauthorized
	}
	return stateToken, nil
}

// DeleteGoogleSSOStateCookie deletes the Google SSO state cookie
func DeleteGoogleSSOStateCookie(c *gin.Context) {
	c.SetCookie(
		authConstants.AuthGoogleSSOStateCookieName,
		"",
		-1, // MaxAge -1 deletes the cookie
		"/",
		config.Cfg.App.Auth.FingerprintCookieDomain,
		true, // secure
		true, // httpOnly
	)
}
