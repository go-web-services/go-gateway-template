package middleware

import (
	"errors"

	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
	internalError "github.com/go-web-services/go-gateway-template/internal/error"
	"github.com/go-web-services/go-gateway-template/internal/utils"

	"github.com/gin-gonic/gin"
	platformConsts "github.com/go-web-services/go-web-platform/constants"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/logger"
)

// AuthMiddleware is a middleware to check if the request is authorized.
// It checks if the user ID is present in the context and returns unauthorized error if not.
// If UserInfoMiddleware was used before this, it will also check for auth errors.
func AuthMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, check if there was an auth error from UserInfoMiddleware
		if authErr, exists := c.Get(authConstants.AuthErrorContextKey); exists && authErr != nil {
			err := authErr.(error)
			handleAuthError(c, log, err)
			return
		}

		// Check if user ID is in context
		userID, exists := c.Get(authConstants.AuthUserIDContextKey)
		if !exists || userID == "" {
			log.Error("User ID not found in context")
			_ = c.Error(platformError.ErrUnauthorized)
			return
		}

		// Successfully authenticated - continue with request
		c.Next()
	}
}

// handleAuthError handles authentication errors by returning appropriate HTTP status codes
func handleAuthError(c *gin.Context, log logger.Logger, err error) {
	if err == nil {
		return
	}

	log.Warn("Authentication failed: ", err)

	// Check for token refreshed error - return with specific error code
	if errors.Is(err, internalError.ErrTokenRefreshed) {
		_ = c.Error(platformError.NewError(authConstants.AuthTokenRefreshedCode, internalError.ErrTokenRefreshed.Error()))
		return
	}

	// Check for specific missing credentials errors
	if errors.Is(err, internalError.ErrAuthTokenMissing) {
		_ = c.Error(platformError.ErrUnauthorized)
		return
	}

	if errors.Is(err, internalError.ErrFingerprintMissing) {
		_ = c.Error(platformError.ErrUnauthorized)
		return
	}

	// Check if it's an authentication error (401)
	if utils.IsAuthenticationError(err) {
		_ = c.Error(platformError.NewError(platformConsts.UnauthorizedError, err.Error()))
		return
	}

	// Check if it's a BaseError with status code
	var baseErr *platformError.BaseError
	if errors.As(err, &baseErr) {
		if baseErr.Status == 400 {
			if baseErr.Code == platformConsts.EntityNotFound {
				_ = c.Error(platformError.ErrEntityNotFound)
				return
			}
			_ = c.Error(platformError.ErrUnauthorized)
			return
		}

		_ = c.Error(platformError.ErrInternalServerError)
		return
	}

	// For any other error, return unauthorized (don't expose internal errors)
	log.Warn("Unexpected error during authorization: ", err)
	_ = c.Error(platformError.ErrUnauthorized)
}
