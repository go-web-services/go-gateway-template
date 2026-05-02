package http

import (
	"github.com/gin-gonic/gin"
	"github.com/go-web-services/go-web-platform/logger"

	eventClient "github.com/go-web-services/go-service-event/pkg/client/service"
	userClient "github.com/go-web-services/go-service-user/pkg/client/service"

	"github.com/go-web-services/go-gateway-template/internal/transport/http/handler"
	"github.com/go-web-services/go-gateway-template/internal/transport/http/middleware"
)

// SetupRouter adds all routes to the provided router.
func SetupRouter(
	router *gin.Engine,
	log logger.Logger,
	userAuthAPIClient userClient.AuthAPIService,
	userUserAPIClient userClient.UserAPIService,
	userGoogleSSOAPIClient userClient.GoogleSSOAPIService,
	eventAPIClient eventClient.EventAPIService,
) *gin.Engine {
	authHandler := handler.NewAuthHandler(
		log,
		userAuthAPIClient,
		userGoogleSSOAPIClient,
	)

	userHandler := handler.NewUserHandler(
		log,
		userAuthAPIClient,
		userUserAPIClient,
	)

	eventHandler := handler.NewEventHandler(
		log,
		eventAPIClient,
	)

	website := router.Group("/api/v1")
	{
		auth := website.Group("/auth")
		{
			auth.POST("/login", middleware.CaptchaMiddleware(), authHandler.LoginV1)
			auth.POST("/logout",
				middleware.UserInfoMiddleware(log, userAuthAPIClient),
				middleware.AuthMiddleware(log),
				authHandler.LogoutV1,
			)
			auth.POST("/signup", middleware.CaptchaMiddleware(), authHandler.SignupV1)
			auth.POST("/activate-account", authHandler.ActivateAccountV1)
			auth.POST("/activate-account/resend", middleware.CaptchaMiddleware(), authHandler.ResendActivationEmailV1)
			auth.POST("/forgot-password/start", middleware.CaptchaMiddleware(), authHandler.ForgotPasswordStartV1)
			auth.POST("/forgot-password/finish", middleware.CaptchaMiddleware(), authHandler.ForgotPasswordFinishV1)
			auth.POST("/forgot-password/check-token", authHandler.CheckForgotPasswordTokenV1)
			auth.POST("/google-sso/get-link", authHandler.GoogleSSOGetLinkV1)
			auth.POST("/google-sso/callback", authHandler.GoogleSSOCallbackV1)
			auth.POST("/otp/signup", middleware.CaptchaMiddleware(), authHandler.OTPSignupV1)
			auth.POST("/otp/login", middleware.CaptchaMiddleware(), authHandler.OTPLoginV1)
		}

		users := website.Group("/users")
		{
			users.GET("/me",
				middleware.UserInfoMiddleware(log, userAuthAPIClient),
				middleware.AuthMiddleware(log),
				userHandler.GetCurrentUserV1,
			)
			users.POST("/update",
				middleware.UserInfoMiddleware(log, userAuthAPIClient),
				middleware.AuthMiddleware(log),
				userHandler.UpdateUserV1,
			)
			users.POST("/providers/list",
				middleware.UserInfoMiddleware(log, userAuthAPIClient),
				middleware.AuthMiddleware(log),
				userHandler.ListAuthProvidersV1,
			)
		}

		// Events (optional auth: user_id is set when UserInfoMiddleware succeeds)
		events := website.Group("/events")
		{
			events.POST("/send",
				middleware.UserInfoMiddleware(log, userAuthAPIClient),
				eventHandler.SendEventV1,
			)
		}
	}

	return router
}
