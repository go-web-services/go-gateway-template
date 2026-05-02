package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	authClientDTO "github.com/go-web-services/go-service-user/pkg/client/dto"
	userClient "github.com/go-web-services/go-service-user/pkg/client/service"
	platformConsts "github.com/go-web-services/go-web-platform/constants"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/logger"

	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
	"github.com/go-web-services/go-gateway-template/internal/dto"
	internalError "github.com/go-web-services/go-gateway-template/internal/error"
	"github.com/go-web-services/go-gateway-template/internal/utils"
	"github.com/go-web-services/go-gateway-template/internal/validation"
)

type AuthHandler interface {
	LoginV1(c *gin.Context)
	LogoutV1(c *gin.Context)
	ForgotPasswordFinishV1(c *gin.Context)
	ForgotPasswordStartV1(c *gin.Context)
	SignupV1(c *gin.Context)
	ActivateAccountV1(c *gin.Context)
	ResendActivationEmailV1(c *gin.Context)
	CheckForgotPasswordTokenV1(c *gin.Context)
	GoogleSSOGetLinkV1(c *gin.Context)
	GoogleSSOCallbackV1(c *gin.Context)
	OTPSignupV1(c *gin.Context)
	OTPLoginV1(c *gin.Context)
}

type authHandler struct {
	authAPIClient      userClient.AuthAPIService
	googleSSOAPIClient userClient.GoogleSSOAPIService
	log                logger.Logger
	validate           *validator.Validate
}

func NewAuthHandler(
	log logger.Logger,
	authAPIClient userClient.AuthAPIService,
	googleSSOAPIClient userClient.GoogleSSOAPIService,
) AuthHandler {
	v := validator.New()
	validation.RegisterCustomValidators(v)

	return &authHandler{
		log:                log,
		authAPIClient:      authAPIClient,
		googleSSOAPIClient: googleSSOAPIClient,
		validate:           v,
	}
}

// LoginV1
// @Summary Perform user login.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.LoginInputDTO true "Send Login Request"
// @Success 200 {object} dto.LoginOutputDTO
// @Router /v1/auth/login [post]
func (h *authHandler) LoginV1(c *gin.Context) {
	var req dto.LoginInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.LoginInputDTO{
		Email:    req.Email,
		Password: req.Password,
		Product:  authConstants.AuthProduct,
	}
	result, err := h.authAPIClient.LoginV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	utils.SetFingerprintCookie(c, result.Fingerprint)
	c.Header(authConstants.AuthNewTokenHeaderKey, result.AuthToken)

	response := dto.LoginOutputDTO{
		User: dto.UserDTO{
			ID:          result.User.ID,
			Email:       result.User.Email,
			Username:    result.User.Username,
			CreatedAt:   result.User.CreatedAt,
			Status:      result.User.Status,
			HasPassword: result.User.HasPassword,
		},
	}
	c.JSON(http.StatusOK, response)
}

// LogoutV1
// @Summary Perform user logout.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.LogoutOutputDTO
// @Router /v1/auth/logout [post]
func (h *authHandler) LogoutV1(c *gin.Context) {
	authToken, err := utils.ExtractAuthToken(c)
	if err != nil {
		_ = c.Error(platformError.ErrUnauthorized)
		return
	}

	clientReq := authClientDTO.LogoutInputDTO{
		AuthToken: authToken,
	}
	_, err = h.authAPIClient.LogoutV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	utils.DeleteFingerprintCookie(c)

	response := dto.LogoutOutputDTO{
		Message: "Logged out successfully",
	}
	c.JSON(http.StatusOK, response)
}

// ForgotPasswordStartV1
// @Summary Generate password reset token and expiration date (start forgot password flow).
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.ForgotPasswordStartInputDTO true "Send Forgot Password Start Request"
// @Success 200 {object} dto.ForgotPasswordStartOutputDTO
// @Router /v1/auth/forgot-password/start [post]
func (h *authHandler) ForgotPasswordStartV1(c *gin.Context) {
	var req dto.ForgotPasswordStartInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.ForgotPasswordStartInputDTO{
		Email:   req.Email,
		Product: authConstants.AuthProduct,
	}
	_, err := h.authAPIClient.ForgotPasswordStartV1(c, clientReq)
	if err != nil {
		var baseErr *platformError.BaseError
		if errors.As(err, &baseErr) && baseErr.Code == platformConsts.EntityNotFound {
			h.log.Info("User not found for password reset, returning success to prevent enumeration")
			response := dto.ForgotPasswordStartOutputDTO{
				Message: "Password reset email sent",
			}
			c.JSON(http.StatusOK, response)
			return
		}

		_ = c.Error(err)
		return
	}

	response := dto.ForgotPasswordStartOutputDTO{
		Message: "Password reset email sent",
	}
	c.JSON(http.StatusOK, response)
}

// ForgotPasswordFinishV1
// @Summary Update user password (finish forgot password flow).
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.ForgotPasswordFinishInputDTO true "Send Forgot Password Finish Request"
// @Success 200 {object} dto.ForgotPasswordFinishOutputDTO
// @Router /v1/auth/forgot-password/finish [post]
func (h *authHandler) ForgotPasswordFinishV1(c *gin.Context) {
	var req dto.ForgotPasswordFinishInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.ForgotPasswordFinishInputDTO{
		ForgotPasswordToken: req.ForgotPasswordToken,
		NewPassword:         req.NewPassword,
		DeleteSessions:      req.DeleteSessions,
	}
	_, err := h.authAPIClient.ForgotPasswordFinishV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.ForgotPasswordFinishOutputDTO{
		Message: "Password updated successfully",
	}
	c.JSON(http.StatusOK, response)
}

// SignupV1
// @Summary Create a new user account.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.SignupInputDTO true "Send Signup Request"
// @Success 200 {object} dto.SignupOutputDTO
// @Router /v1/auth/signup [post]
func (h *authHandler) SignupV1(c *gin.Context) {
	var req dto.SignupInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.SignupInputDTO{
		Email:    req.Email,
		Product:  authConstants.AuthProduct,
		Username: req.Username,
		Password: req.Password,
	}
	_, err := h.authAPIClient.SignupV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.SignupOutputDTO{
		Message: "Account confirmation email sent",
	}
	c.JSON(http.StatusOK, response)
}

// ActivateAccountV1
// @Summary Confirm user email with activation token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.ActivateAccountInputDTO true "Send Activate Account Request"
// @Success 200 {object} dto.ActivateAccountOutputDTO
// @Router /v1/auth/activate-account [post]
func (h *authHandler) ActivateAccountV1(c *gin.Context) {
	var req dto.ActivateAccountInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.ActivateAccountInputDTO{
		ActivationToken: req.ActivationToken,
	}
	result, err := h.authAPIClient.ActivateAccountV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	utils.SetFingerprintCookie(c, result.Fingerprint)
	c.Header(authConstants.AuthNewTokenHeaderKey, result.AuthToken)

	response := dto.ActivateAccountOutputDTO{
		User: dto.UserDTO{
			ID:          result.User.ID,
			Email:       result.User.Email,
			Username:    result.User.Username,
			CreatedAt:   result.User.CreatedAt,
			Status:      result.User.Status,
			HasPassword: result.User.HasPassword,
		},
	}
	c.JSON(http.StatusOK, response)
}

// ResendActivationEmailV1
// @Summary Resend activation email for pending users.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.ResendActivationEmailInputDTO true "Send Resend Activation Email Request"
// @Success 200 {object} dto.ResendActivationEmailOutputDTO
// @Router /v1/auth/activate-account/resend [post]
func (h *authHandler) ResendActivationEmailV1(c *gin.Context) {
	var req dto.ResendActivationEmailInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.ResendActivationEmailInputDTO{
		Email:   req.Email,
		Product: authConstants.AuthProduct,
	}
	_, err := h.authAPIClient.ResendActivationEmailV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.ResendActivationEmailOutputDTO{
		Message: "Activation email sent",
	}
	c.JSON(http.StatusOK, response)
}

// CheckForgotPasswordTokenV1
// @Summary Check if forgot password token is valid and not expired.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.CheckForgotPasswordTokenInputDTO true "Send Check Forgot Password Token Request"
// @Success 200 {object} dto.CheckForgotPasswordTokenOutputDTO
// @Router /v1/auth/forgot-password/check-token [post]
func (h *authHandler) CheckForgotPasswordTokenV1(c *gin.Context) {
	var req dto.CheckForgotPasswordTokenInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.CheckForgotPasswordTokenInputDTO{
		ForgotPasswordToken: req.ForgotPasswordToken,
	}
	result, err := h.authAPIClient.CheckForgotPasswordTokenV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.CheckForgotPasswordTokenOutputDTO{
		Valid: result.Valid,
	}
	c.JSON(http.StatusOK, response)
}

// GoogleSSOGetLinkV1
// @Summary Get Google SSO authentication link.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} dto.GoogleSSOGetLinkOutputDTO
// @Router /v1/auth/google-sso/get-link [post]
func (h *authHandler) GoogleSSOGetLinkV1(c *gin.Context) {
	clientReq := authClientDTO.GoogleSSOGetLinkInputDTO{
		Product: authConstants.AuthProduct,
	}
	result, err := h.googleSSOAPIClient.GetAuthLinkV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	csrfToken, err := utils.GenerateCSRFToken()
	if err != nil {
		h.log.Error("Failed to generate CSRF token: ", err)
		_ = c.Error(platformError.ErrInternalServerError)
		return
	}

	utils.SetGoogleSSOStateCookie(c, csrfToken)

	parsedURL, err := url.Parse(result.AuthURL)
	if err != nil {
		h.log.Error("Failed to parse auth URL: ", err)
		_ = c.Error(platformError.ErrInternalServerError)
		return
	}

	query := parsedURL.Query()
	existingState := query.Get("state")
	newState := existingState + "|" + csrfToken
	query.Set("state", newState)
	parsedURL.RawQuery = query.Encode()

	response := dto.GoogleSSOGetLinkOutputDTO{
		AuthURL: parsedURL.String(),
	}
	c.JSON(http.StatusOK, response)
}

// GoogleSSOCallbackV1
// @Summary Handle Google SSO callback.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.GoogleSSOCallbackInputDTO true "Send Google SSO Callback Request"
// @Success 200 {object} dto.GoogleSSOCallbackOutputDTO
// @Router /v1/auth/google-sso/callback [post]
func (h *authHandler) GoogleSSOCallbackV1(c *gin.Context) {
	var req dto.GoogleSSOCallbackInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	decodedState, err := url.QueryUnescape(req.State)
	if err != nil {
		h.log.Warn("Failed to decode state parameter: ", err)
		_ = c.Error(platformError.NewError(authConstants.AuthCSRFErrorCode, internalError.ErrInvalidStateFormat.Error()))
		return
	}

	stateParts := strings.Split(decodedState, "|")
	if len(stateParts) != 2 {
		h.log.Warn("Invalid state parameter format")
		_ = c.Error(platformError.NewError(authConstants.AuthCSRFErrorCode, internalError.ErrInvalidStateFormat.Error()))
		return
	}

	receivedCSRFToken := stateParts[1]
	originalState := stateParts[0]

	storedCSRFToken, err := utils.GetGoogleSSOStateCookie(c)
	if err != nil {
		h.log.Warn("CSRF token cookie not found")
		_ = c.Error(platformError.NewError(authConstants.AuthCSRFErrorCode, internalError.ErrCSRFTokenMissing.Error()))
		return
	}

	if receivedCSRFToken != storedCSRFToken {
		h.log.Warn("CSRF token mismatch")
		_ = c.Error(platformError.NewError(authConstants.AuthCSRFErrorCode, internalError.ErrCSRFTokenMismatch.Error()))
		return
	}

	utils.DeleteGoogleSSOStateCookie(c)

	clientReq := authClientDTO.GoogleSSOCallbackInputDTO{
		Code:  req.Code,
		State: originalState,
	}
	result, err := h.googleSSOAPIClient.CallbackV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	utils.SetFingerprintCookie(c, result.Fingerprint)
	c.Header(authConstants.AuthNewTokenHeaderKey, result.AuthToken)

	response := dto.GoogleSSOCallbackOutputDTO{
		User: dto.UserDTO{
			ID:          result.User.ID,
			Email:       result.User.Email,
			Username:    result.User.Username,
			CreatedAt:   result.User.CreatedAt,
			Status:      result.User.Status,
			HasPassword: result.User.HasPassword,
		},
		IsNewUser: result.IsNewUser,
	}
	c.JSON(http.StatusOK, response)
}

// OTPSignupV1
// @Summary Request OTP for signup.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.OTPSignupInputDTO true "Send OTP Signup Request"
// @Success 200 {object} dto.OTPSignupOutputDTO
// @Router /v1/auth/otp/signup [post]
func (h *authHandler) OTPSignupV1(c *gin.Context) {
	var req dto.OTPSignupInputDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.OTPGenerateInputDTO{
		Email:   req.Email,
		Product: authConstants.AuthProduct,
	}

	result, err := h.authAPIClient.OTPSignupV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := dto.OTPSignupOutputDTO{
		User: dto.UserDTO{
			ID:          result.User.ID,
			Email:       result.User.Email,
			Username:    result.User.Username,
			CreatedAt:   result.User.CreatedAt,
			Status:      result.User.Status,
			HasPassword: result.User.HasPassword,
		},
	}
	c.JSON(http.StatusOK, response)
}

// OTPLoginV1
// @Summary Login with OTP.
// @Tags Auth
// @Accept json
// @Produce json
// @Param SendRequest body dto.OTPLoginInputDTO true "Send OTP Login Request"
// @Success 200 {object} dto.OTPLoginOutputDTO
// @Router /v1/auth/otp/login [post]
func (h *authHandler) OTPLoginV1(c *gin.Context) {
	var req dto.OTPLoginInputDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	clientReq := authClientDTO.OTPLoginInputDTO{
		Email:   req.Email,
		OTP:     req.OTP,
		Product: authConstants.AuthProduct,
	}

	result, err := h.authAPIClient.OTPLoginV1(c, clientReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	utils.SetFingerprintCookie(c, result.Fingerprint)
	c.Header(authConstants.AuthNewTokenHeaderKey, result.AccessToken)

	response := dto.OTPLoginOutputDTO{
		User: dto.UserDTO{
			ID:          result.User.ID,
			Email:       result.User.Email,
			Username:    result.User.Username,
			CreatedAt:   result.User.CreatedAt,
			Status:      result.User.Status,
			HasPassword: result.User.HasPassword,
		},
	}
	c.JSON(http.StatusOK, response)
}
