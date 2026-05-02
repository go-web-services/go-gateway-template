package dto

import (
	"time"

	"github.com/go-web-services/go-service-user/pkg/client/enum"
)

type GoogleSSOGetLinkOutputDTO struct {
	AuthURL string `json:"auth_url"`
}

type GoogleSSOCallbackInputDTO struct {
	Code  string `json:"code" validate:"required,max=512"`
	State string `json:"state" validate:"required,max=512"`
}

type GoogleSSOCallbackOutputDTO struct {
	User      UserDTO `json:"user"`
	IsNewUser bool    `json:"is_new_user"`
}

type AuthProviderDTO struct {
	Type string `json:"type"`
}

type ListAuthProvidersOutputDTO struct {
	AuthProviders []AuthProviderDTO `json:"auth_providers"`
}

type OTPSignupInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type OTPSignupOutputDTO struct {
	User UserDTO `json:"user"`
}

type OTPLoginInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	OTP                 string `json:"otp" validate:"required,max=6"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type OTPLoginOutputDTO struct {
	User UserDTO `json:"user"`
}

type LoginInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	Password            string `json:"password" validate:"required,max=255"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type LoginOutputDTO struct {
	User UserDTO `json:"user"`
}

type LogoutOutputDTO struct {
	Message string `json:"message"`
}

type ForgotPasswordStartInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type ForgotPasswordStartOutputDTO struct {
	Message string `json:"message"`
}

type ForgotPasswordFinishInputDTO struct {
	ForgotPasswordToken string `json:"forgot_password_token" validate:"required,max=255"`
	NewPassword         string `json:"new_password" validate:"required,strong_password,max=255"`
	DeleteSessions      bool   `json:"delete_sessions"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type ForgotPasswordFinishOutputDTO struct {
	Message string `json:"message"`
}

type SignupInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	Username            string `json:"username" validate:"required,username_alpha,max=64"`
	Password            string `json:"password" validate:"required,strong_password,max=255"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type SignupOutputDTO struct {
	Message string `json:"message"`
}

type ActivateAccountInputDTO struct {
	ActivationToken string `json:"activation_token" validate:"required,max=255"`
}

type ActivateAccountOutputDTO struct {
	User UserDTO `json:"user"`
}

type CheckForgotPasswordTokenInputDTO struct {
	ForgotPasswordToken string `json:"forgot_password_token" validate:"required,max=255"`
}

type CheckForgotPasswordTokenOutputDTO struct {
	Valid bool `json:"valid"`
}

type ResendActivationEmailInputDTO struct {
	Email               string `json:"email" validate:"required,email,max=320"`
	CfTurnstileResponse string `json:"cf_turnstile_response" validate:"required"`
}

type ResendActivationEmailOutputDTO struct {
	Message string `json:"message"`
}

type UserDTO struct {
	ID          string          `json:"id"`
	Email       string          `json:"email"`
	Username    string          `json:"username"`
	CreatedAt   time.Time       `json:"created_at"`
	Status      enum.UserStatus `json:"status"`
	HasPassword bool            `json:"has_password"`
}

type UpdateUserInputDTO struct {
	Username    string `json:"username" validate:"omitempty,username_alpha,max=64"`
	OldPassword string `json:"old_password" validate:"omitempty,max=255"`
	NewPassword string `json:"new_password" validate:"omitempty,strong_password,max=255"`
}

type UpdateUserOutputDTO struct {
	User UserDTO `json:"user"`
}
