package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	eventClientDTO "github.com/go-web-services/go-service-event/pkg/client/dto"
	eventClient "github.com/go-web-services/go-service-event/pkg/client/service"
	platformError "github.com/go-web-services/go-web-platform/error"
	"github.com/go-web-services/go-web-platform/logger"
	"github.com/google/uuid"

	authConstants "github.com/go-web-services/go-gateway-template/internal/constants"
	"github.com/go-web-services/go-gateway-template/internal/dto"
)

type EventHandler interface {
	SendEventV1(c *gin.Context)
}

type eventHandler struct {
	log            logger.Logger
	eventAPIClient eventClient.EventAPIService
	validate       *validator.Validate
}

func NewEventHandler(
	log logger.Logger,
	eventAPIClient eventClient.EventAPIService,
) EventHandler {
	return &eventHandler{
		log:            log,
		eventAPIClient: eventAPIClient,
		validate:       validator.New(),
	}
}

// SendEventV1
// @Summary Track an analytics event
// @Tags Events
// @Accept json
// @Produce json
// @Param request body dto.EventSendRequestDTO true "Event payload from the client"
// @Success 200 {object} dto.EventSendOutputDTO
// @Router /v1/events/send [post]
func (h *eventHandler) SendEventV1(c *gin.Context) {
	var req dto.EventSendRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	messageID := strings.TrimSpace(req.MessageID)
	if messageID != "" {
		if _, err := uuid.Parse(messageID); err != nil {
			_ = c.Error(platformError.ErrInvalidRequestPayload)
			return
		}
	} else {
		messageID = uuid.New().String()
	}

	var userID *string
	if uid, exists := c.Get(authConstants.AuthUserIDContextKey); exists {
		if s, ok := uid.(string); ok && s != "" {
			userID = &s
		}
	}

	// ClientIP: behind nginx, Gin uses X-Forwarded-For (see proxy/nginx.conf). When the gateway is hit directly
	// from the host in local Docker, the address is typically the Docker bridge/gateway (e.g. 172.19.0.1), not the browser's public IP.
	ip := c.ClientIP()
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}

	ua := c.Request.UserAgent()
	var uaPtr *string
	if ua != "" {
		uaPtr = &ua
	}

	input := eventClientDTO.EventCreateInputDTO{
		ProjectID:  authConstants.EventProjectID,
		MessageID:  messageID,
		DistinctID: req.DistinctID,
		UserID:     userID,
		SessionID:  req.SessionID,
		IP:         ipPtr,
		UserAgent:  uaPtr,
		Name:       req.Name,
		Payload:    req.Payload,
		OccurredAt: req.OccurredAt,
	}

	if err := h.validate.Struct(input); err != nil {
		_ = c.Error(platformError.ErrInvalidRequestPayload)
		return
	}

	if _, err := h.eventAPIClient.CreateV1(c, input); err != nil {
		h.log.Error("Failed to create event: ", err)
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.EventSendOutputDTO{Message: "success"})
}
