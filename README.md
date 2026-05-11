# go-gateway-template

`github.com/go-web-services/go-gateway-template`

Minimal HTTP gateway that proxies authentication and user flows to [go-service-user](https://github.com/go-web-services/go-service-user) and analytics events to [go-service-event](https://github.com/go-web-services/go-service-event) via their published clients. Includes Cloudflare Turnstile captcha on sensitive OTP and auth routes. Production nginx lives in [go-web-infrastructure](https://github.com/go-web-services/go-web-infrastructure) (`proxy/`). Fork this as a starting point for other domain gateways; replace the module path and adjust constants before use.

---

## Responsibilities

- Proxy all auth and user-management requests to go-service-user.
- Proxy analytics event submissions to go-service-event.
- Enforce Cloudflare Turnstile captcha on signup, OTP, and password-reset routes.
- Apply CORS with a configurable allow-list.
- Attach `user_id` to outbound event calls when a valid session is present (`UserInfoMiddleware`).
---

## Configuration

| Variable | Purpose | Default |
|----------|---------|---------|
| `APP_PORT` | HTTP listen port | — |
| `APP_ENV` | Environment (`dev` / `prod`) | — |
| `ALLOW_ORIGINS` | Comma-separated CORS origins | — |
| `USER_SERVICE_URL` | Base URL for go-service-user | — |
| `EVENT_SERVICE_URL` | Base URL for go-service-event | — |
| `TURNSTILE_SECRET_KEY` | Cloudflare Turnstile secret (captcha validation) | — |
| `AUTH_FINGERPRINT_COOKIE_DOMAIN` | Cookie `Domain` for fingerprint / SSO state cookies | — |
| `AUTH_FINGERPRINT_COOKIE_EXPIRATION_SEC` | Fingerprint cookie max-age in seconds | — |

Two constants also need to match your deployment:

- `AuthProduct` in `internal/constants/auth_constants.go` — product ID sent to the user service (default: `main`).
- `EventProjectID` in `internal/constants/event_constants.go` — project scope sent to the event service.

After forking, replace the module path `github.com/go-web-services/go-gateway-template` everywhere (`go.mod` and all imports), then run `gocheck`.

---

## Run locally

```bash
git clone git@github.com:go-web-services/go-gateway-template.git
cd go-gateway-template
cp .env.sample .env
# Set USER_SERVICE_URL, EVENT_SERVICE_URL, TURNSTILE_SECRET_KEY, GITHUB_TOKEN
go run ./cmd/app/main.go
```

---

## Docker

- **Dev** (hot reload via `debug/Dockerfile`):
  ```bash
  docker compose -f docker-compose.yml up
  ```
- **Prod**:
  ```bash
  docker compose -f docker-compose-prod.yml up --build
  ```

Ensure your reverse proxy upstreams use the compose service hostname `go-gateway-template` and the same `APP_PORT` value as in `.env` (see `go-web-infrastructure` for nginx compose and sample configs).

For Docker builds with private Go modules, pass `GITHUB_TOKEN` as a build arg (see `args` in the compose files). Locally, set `GOPRIVATE=github.com/go-web-services/*`.

---

## API surface

### Auth (`/api/v1/auth`)

| Method | Path | Notes |
|--------|------|-------|
| `POST` | `/login` | Captcha |
| `POST` | `/logout` | Session + fingerprint middleware |
| `POST` | `/signup` | Captcha |
| `POST` | `/activate-account` | — |
| `POST` | `/activate-account/resend` | Captcha |
| `POST` | `/forgot-password/start` | Captcha |
| `POST` | `/forgot-password/finish` | Captcha |
| `POST` | `/forgot-password/check-token` | — |
| `POST` | `/google-sso/get-link` | — |
| `POST` | `/google-sso/callback` | — |
| `POST` | `/otp/signup` | Captcha |
| `POST` | `/otp/login` | Captcha |

### Users (`/api/v1/users`) — require auth

| Method | Path | Notes |
|--------|------|-------|
| `GET` | `/me` | Returns current user |
| `POST` | `/update` | Update user profile |
| `POST` | `/providers/list` | List auth providers |

### Events (`/api/v1/events`)

| Method | Path | Notes |
|--------|------|-------|
| `POST` | `/send` | Optional auth; sets `user_id` when session is valid |

Swagger UI is available at `/swagger` when enabled for your environment (wired through go-web-platform).

---

## Private dependencies

This gateway imports `go-service-user/pkg/client`, `go-service-event/pkg/client`, and `go-web-platform`, all of which may be private. Configure access:

```bash
export GOPRIVATE='github.com/go-web-services/*'
```

---

## Author

[Lomank](https://lomank.com)
