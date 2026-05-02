package config

import (
	"strconv"
	"strings"

	platformTypes "github.com/go-web-services/go-web-platform/types"
	platformUtils "github.com/go-web-services/go-web-platform/utils"
)

type AuthConfig struct {
	FingerprintCookieDomain        string
	FingerprintCookieExpirationSec int
}

type AppConfig struct {
	Port               int
	Env                platformTypes.Environment
	SwaggerBasePath    string
	Auth               AuthConfig
	AllowOrigins       []string
	TurnstileSecretKey string
}

type ServicesConfig struct {
	UserServiceURL  string
	EventServiceURL string
}

type Config struct {
	App      AppConfig
	Services ServicesConfig
}

var Cfg Config

func LoadConfig() (*Config, error) {

	portStr := platformUtils.GetEnv("APP_PORT", "8080")
	port, _ := strconv.Atoi(portStr)

	authFingerprintCookieExpirationStr := platformUtils.GetEnv("AUTH_FINGERPRINT_COOKIE_EXPIRATION_SEC", "2592000")
	authFingerprintCookieExpirationSec, _ := strconv.Atoi(authFingerprintCookieExpirationStr)

	allowOriginsValue := platformUtils.GetEnv(
		"ALLOW_ORIGINS",
		"http://localhost:3000,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:8080",
	)
	corsOrigins := strings.Split(allowOriginsValue, ",")

	Cfg = Config{
		App: AppConfig{
			Port: port,
			Env:  platformTypes.Environment(platformUtils.GetEnv("APP_ENV", "dev")),
			Auth: AuthConfig{
				FingerprintCookieDomain:        platformUtils.GetEnv("AUTH_FINGERPRINT_COOKIE_DOMAIN", ""),
				FingerprintCookieExpirationSec: authFingerprintCookieExpirationSec,
			},
			AllowOrigins:       corsOrigins,
			TurnstileSecretKey: platformUtils.GetEnv("TURNSTILE_SECRET_KEY", ""),
		},
		Services: ServicesConfig{
			UserServiceURL:  platformUtils.GetEnv("USER_SERVICE_URL", "http://host.docker.internal:8007"),
			EventServiceURL: platformUtils.GetEnv("EVENT_SERVICE_URL", "http://host.docker.internal:8010"),
		},
	}

	return &Cfg, nil
}
