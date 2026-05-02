package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// ValidatePassword validates password requirements:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// - At least one special character/symbol
func ValidatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check minimum length
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return false
	}

	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return false
	}

	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return false
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)
	if !hasSpecial {
		return false
	}

	return true
}

// ValidateUsername validates username requirements:
// - Only uppercase and lowercase letters (A-Z, a-z) and spaces
// - No other whitespace characters (tabs, newlines, etc.)
// - Maximum 64 characters
func ValidateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Check maximum length
	if len(username) > 64 {
		return false
	}

	// Check that it only contains letters and spaces (no tabs, newlines, or other special characters)
	onlyLettersAndSpaces := regexp.MustCompile(`^[A-Za-z ]+$`).MatchString(username)
	return onlyLettersAndSpaces
}

// RegisterCustomValidators registers all custom validators
func RegisterCustomValidators(validate *validator.Validate) {
	validate.RegisterValidation("strong_password", ValidatePassword)
	validate.RegisterValidation("username_alpha", ValidateUsername)
}
