package constants

const (
	// AuthProduct is the user-service product identifier; align with your user API configuration.
	AuthProduct = "main"

	AuthFingerprintCookieName    = "fingerprint"
	AuthGoogleSSOStateCookieName = "google_sso_state"
	AuthNewTokenHeaderKey        = "X-New-Auth-Token"
	AuthAuthorizationHeaderKey   = "Authorization"
	AuthBearerPrefix             = "Bearer"

	// Context keys

	AuthUserIDContextKey   = "user_id"
	AuthUserInfoContextKey = "user_info"
	AuthErrorContextKey    = "auth_error"

	// Error codes

	AuthCSRFErrorCode      = "CSRF_ERROR"
	AuthTokenRefreshedCode = "TOKEN_REFRESHED"
)
