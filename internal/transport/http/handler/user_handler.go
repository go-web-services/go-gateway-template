package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	authClientDTO "github.com/go-web-services/go-service-user/pkg/client/dto"
	"github.com/go-web-services/go-service-user/pkg/client/enum"
	userClient "github.com/go-web-services/go-service-user/pkg/client/service"
	platformConsts "github.com/go-web-services/go-web-platform/constants"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/logger"

	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
	"github.com/go-web-services/go-gateway-template/internal/dto"
	"github.com/go-web-services/go-gateway-template/internal/validation"
)

type UserHandler interface {
	GetCurrentUserV1(c *gin.Context)
	UpdateUserV1(c *gin.Context)
	ListAuthProvidersV1(c *gin.Context)
}

type userHandler struct {
	authAPIClient userClient.AuthAPIService
	userAPIClient userClient.UserAPIService
	log           logger.Logger
	validate      *validator.Validate
}

func NewUserHandler(
	log logger.Logger,
	authAPIClient userClient.AuthAPIService,
	userAPIClient userClient.UserAPIService,
) UserHandler {
	v := validator.New()
	validation.RegisterCustomValidators(v)

	return &userHandler{
		log:           log,
		authAPIClient: authAPIClient,
		userAPIClient: userAPIClient,
		validate:      v,
	}
}

// GetCurrentUserV1
// @Summary Get current authenticated user information.
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} dto.UserDTO
// @Router /v1/users/me [get]
func (h *userHandler) GetCurrentUserV1(c *gin.Context) {
	userInfoInterface, exists := c.Get(authConstants.AuthUserInfoContextKey)
	if !exists {
		_ = c.Error(platformError.ErrUnauthorized)
		return
	}

	userInfo, ok := userInfoInterface.(*authClientDTO.UserDTO)
	if !ok {
		h.log.Error("Failed to cast user info to UserDTO")
		_ = c.Error(platformError.ErrInternalServerError)
		return
	}

	response := dto.UserDTO{
		ID:          userInfo.ID,
		Email:       userInfo.Email,
		Username:    userInfo.Username,
		CreatedAt:   userInfo.CreatedAt,
		Status:      userInfo.Status,
		HasPassword: userInfo.HasPassword,
	}
	c.JSON(http.StatusOK, response)
}

// UpdateUserV1
// @Summary Update user information.
// @Tags Users
// @Accept json
// @Produce json
// @Param SendRequest body dto.UpdateUserInputDTO true "Update user request"
// @Success 200 {object} dto.UpdateUserOutputDTO
// @Router /v1/users/update [post]
func (h *userHandler) UpdateUserV1(c *gin.Context) {
	var req dto.UpdateUserInputDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		_ = c.Error(err)
		return
	}

	userID, exists := c.Get(authConstants.AuthUserIDContextKey)
	if !exists {
		_ = c.Error(platformError.ErrUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.log.Error("Failed to cast user ID to string")
		_ = c.Error(platformError.ErrInternalServerError)
		return
	}

	updateReq := authClientDTO.UserUpdateInputDTO{
		ID: userIDStr,
	}

	if req.Username != "" {
		updateReq.Username = req.Username
	}

	result, err := h.userAPIClient.UpdateV1(c, updateReq)
	if err != nil {
		h.log.Error("Failed to update user: ", err)

		var baseErr *platformError.BaseError
		if errors.As(err, &baseErr) && baseErr.Code == platformConsts.EntityNotFound {
			_ = c.Error(platformError.ErrEntityNotFound)
			return
		}

		_ = c.Error(err)
		return
	}

	response := dto.UpdateUserOutputDTO{
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

// ListAuthProvidersV1
// @Summary List user's auth providers.
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} dto.ListAuthProvidersOutputDTO
// @Router /v1/users/providers/list [post]
func (h *userHandler) ListAuthProvidersV1(c *gin.Context) {
	providers := []dto.AuthProviderDTO{
		{Type: string(enum.AuthProviderGoogle)},
	}

	response := dto.ListAuthProvidersOutputDTO{
		AuthProviders: providers,
	}
	c.JSON(http.StatusOK, response)
}
