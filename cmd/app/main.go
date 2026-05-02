// @title           Go Gateway Template API
// @version         1.0
// @description     Gateway template exposing user, auth, and event HTTP APIs
// @basePath        /api

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	platform "github.com/go-web-services/go-web-platform/entrypoint"
	platformMiddleware "github.com/go-web-services/go-web-platform/middleware"

	eventClient "github.com/go-web-services/go-service-event/pkg/client/service"
	userClient "github.com/go-web-services/go-service-user/pkg/client/service"
	"github.com/go-web-services/go-web-platform/logger"

	"github.com/go-web-services/go-gateway-template/config"
	"github.com/go-web-services/go-gateway-template/docs"

	boHttp "github.com/go-web-services/go-gateway-template/internal/transport/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logg := logger.NewLogger(cfg.App.Env)

	userAuthAPIClient := userClient.NewAuthAPIService(cfg.Services.UserServiceURL)
	userUserAPIClient := userClient.NewUserAPIService(cfg.Services.UserServiceURL)
	userGoogleSSOAPIClient := userClient.NewGoogleSSOAPIService(cfg.Services.UserServiceURL)
	eventAPIClient := eventClient.NewEventAPIService(cfg.Services.EventServiceURL)

	router := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.App.AllowOrigins
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Fingerprint"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	platform.SetupPlatform(
		router,
		logg,
		nil,
		platformMiddleware.DefaultLoggingConfig(),
		cfg.App.Env,
	)

	boHttp.SetupRouter(
		router,
		logg,
		userAuthAPIClient,
		userUserAPIClient,
		userGoogleSSOAPIClient,
		eventAPIClient,
	)

	swaggerBasePath := "/api"
	if cfg.App.SwaggerBasePath != "" {
		swaggerBasePath = "/" + cfg.App.SwaggerBasePath + swaggerBasePath
	}
	docs.SwaggerInfo.BasePath = swaggerBasePath

	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	logg.Info("Starting server on port ", cfg.App.Port)
	go func() {
		if err := router.Run(serverAddr); err != nil {
			logg.Fatal("Failed to start HTTP server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logg.Fatal("Server forced to shutdown: ", err)
	}

	logg.Info("Server stopped.")
}
