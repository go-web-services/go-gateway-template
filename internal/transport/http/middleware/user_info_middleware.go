package middleware

import (
	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
	internalError "github.com/go-web-services/go-gateway-template/internal/error"
	"github.com/go-web-services/go-gateway-template/internal/utils"

	"github.com/gin-gonic/gin"
	authClientDTO "github.com/go-web-services/go-service-user/pkg/client/dto"
	userClient "github.com/go-web-services/go-service-user/pkg/client/service"
	"github.com/go-web-services/go-web-platform/logger"
)

// UserInfoMiddleware is a middleware to get user information from auth service.
// It extracts JWT from Authorization header and fingerprint from secure cookie,
// validates the token, and automatically refreshes it if expired.
// If authentication fails, it continues without setting user info (guest mode).
func UserInfoMiddleware(
	log logger.Logger,
	authAPIClient userClient.AuthAPIService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract JWT from Authorization header
		authToken, err := utils.ExtractAuthToken(c)
		if err != nil {
			log.Debug("Authorization header not found or invalid, continuing as guest")
			// Set specific error in context for AuthMiddleware
			c.Set(authConstants.AuthErrorContextKey, internalError.ErrAuthTokenMissing)
			c.Next()
			return
		}

		// Extract fingerprint from cookie
		fingerprint, err := utils.ExtractFingerprint(c)
		if err != nil {
			log.Warn("Fingerprint cookie not found, continuing as guest")
			// Set specific error in context for AuthMiddleware
			c.Set(authConstants.AuthErrorContextKey, internalError.ErrFingerprintMissing)
			c.Next()
			return
		}

		// Try to authorize with current token
		user, newToken, err := authorizeOrRefresh(
			c,
			log,
			authAPIClient,
			authToken,
			fingerprint,
		)
		if err != nil {
			// Store the auth error in context for AuthMiddleware to handle
			c.Set(authConstants.AuthErrorContextKey, err)
			log.Warn("Failed to authorize user, continuing as guest: ", err)
			c.Next()
			return
		}

		// If token was refreshed, return the new token and require retry
		if newToken != "" {
			c.Header(authConstants.AuthNewTokenHeaderKey, newToken)
			log.Info("Token was refreshed, client should retry with new token")
			// Set specific error in context so AuthMiddleware will return TOKEN_REFRESHED error
			c.Set(authConstants.AuthErrorContextKey, internalError.ErrTokenRefreshed)
			c.Next()
			return
		}

		// Set user info in context
		if user != nil {
			c.Set(authConstants.AuthUserIDContextKey, user.ID)
			c.Set(authConstants.AuthUserInfoContextKey, user)
		}

		c.Next()
	}
}

// authorizeOrRefresh attempts to authorize the token, and if it's expired,
// tries to refresh it automatically.
// Returns user info, new token (if refreshed), and error.
func authorizeOrRefresh(
	c *gin.Context,
	log logger.Logger,
	authAPIClient userClient.AuthAPIService,
	authToken string,
	fingerprint string,
) (*authClientDTO.UserDTO, string, error) {
	// Try to authorize
	authorizeReq := authClientDTO.AuthorizeInputDTO{
		AuthToken:   authToken,
		Fingerprint: fingerprint,
	}

	result, err := authAPIClient.AuthorizeV1(c, authorizeReq)
	if err != nil {
		// Check if error is token expired - if so, try to refresh
		if utils.IsTokenExpiredError(err) {
			log.Info("Token expired, attempting to refresh")

			// Try to refresh token
			refreshReq := authClientDTO.RefreshTokenInputDTO{
				AuthToken:   authToken,
				Fingerprint: fingerprint,
			}

			refreshResult, refreshErr := authAPIClient.RefreshTokenV1(c, refreshReq)
			if refreshErr != nil {
				log.Warn("Failed to refresh token: ", refreshErr)
				return nil, "", refreshErr
			}

			// Return nil user info with new token
			// The new token will be sent to client, and they should retry with it
			// This avoids making a second authorize call
			return nil, refreshResult.AuthToken, nil
		}

		// Not a token expired error, return the error
		return nil, "", err
	}

	// Successfully authorized with current token
	return &result.User, "", nil
}
