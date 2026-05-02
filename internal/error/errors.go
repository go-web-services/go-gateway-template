package error

import "errors"

var (
	// ErrCSRFTokenMismatch is returned when the CSRF token validation fails
	ErrCSRFTokenMismatch = errors.New("CSRF token mismatch")
	// ErrCSRFTokenMissing is returned when the CSRF token is missing
	ErrCSRFTokenMissing = errors.New("CSRF token missing")
	// ErrInvalidStateFormat is returned when the state parameter format is invalid
	ErrInvalidStateFormat = errors.New("invalid state parameter format")
	// ErrAuthTokenMissing is returned when the authorization header is missing
	ErrAuthTokenMissing = errors.New("authorization header missing")
	// ErrFingerprintMissing is returned when the fingerprint cookie is missing
	ErrFingerprintMissing = errors.New("fingerprint cookie missing")
	// ErrTokenRefreshed is returned when the auth token has been refreshed and client should retry with new token
	ErrTokenRefreshed = errors.New("token refreshed, retry with new token")
)
