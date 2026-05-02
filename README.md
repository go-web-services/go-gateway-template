# Go gateway template

Minimal HTTP gateway that proxies **authentication** and **user** flows to [go-service-user](https://github.com/go-web-services/go-service-user) and **analytics events** to [go-service-event](https://github.com/go-web-services/go-service-event) via their published clients. You can fork this as a starting point for other domains.

Route layout and behavior match the auth/user surface from the historical [`lmk-gateway-website`](https://github.com/Lomank123/lmk-gateway-website) tree (initial `SetupRouter` from commit `7a33a02`, with Turnstile on OTP routes as in `b2320e8`). Handlers use the same `go-web-services/go-web-platform` error patterns as the current gateway refactor.

## Module name

Go module: `github.com/go-web-services/go-gateway-template`. After you copy the template, replace it everywhere (including `go.mod` and imports) with your real module path, then run:

```bash
gocheck
```

## Configuration

| Variable | Purpose |
|----------|---------|
| `APP_PORT` | HTTP listen port |
| `APP_ENV` | Environment (`dev` / production values per platform) |
| `ALLOW_ORIGINS` | Comma-separated CORS origins |
| `USER_SERVICE_URL` | Base URL for the user service |
| `EVENT_SERVICE_URL` | Base URL for the event / analytics service |
| `TURNSTILE_SECRET_KEY` | Cloudflare Turnstile secret (captcha on selected routes) |
| `AUTH_FINGERPRINT_COOKIE_DOMAIN` | Cookie `Domain` for fingerprint / SSO state cookies |
| `AUTH_FINGERPRINT_COOKIE_EXPIRATION_SEC` | Fingerprint cookie max-age |

Set `internal/constants/auth_constants.go` **`AuthProduct`** to the product id your user service expects (default in this template is `main`). Set **`EventProjectID`** in `internal/constants/event_constants.go` for your event service project scope.

## Run locally

```bash
cp .env.sample .env
# set USER_SERVICE_URL, TURNSTILE_SECRET_KEY, GITHUB_TOKEN for private modules if needed
go run ./cmd/app/main.go
```

## Docker

- **Dev**: `docker compose -f docker-compose.yml up` (hot reload via `debug/Dockerfile`; service name `go-gateway-template`).
- **Prod**: `docker compose -f docker-compose-prod.yml up --build` builds binary `go-gateway-template` from the root `Dockerfile`.

Ensure reverse proxy upstreams use the compose service hostname `go-gateway-template` and the same `APP_PORT` as in `.env`.

## API surface (`/api/v1`)

**Auth** (`/api/v1/auth`):

- `POST /login` — captcha
- `POST /logout` — optional auth middleware chain (session + fingerprint)
- `POST /signup` — captcha
- `POST /activate-account`
- `POST /activate-account/resend` — captcha
- `POST /forgot-password/start` — captcha
- `POST /forgot-password/finish` — captcha
- `POST /forgot-password/check-token`
- `POST /google-sso/get-link`
- `POST /google-sso/callback`
- `POST /otp/signup` — captcha
- `POST /otp/login` — captcha

**Users** (`/api/v1/users`; require auth):

- `GET /me`
- `POST /update`
- `POST /providers/list`

**Events** (`/api/v1/events`):

- `POST /send` — optional auth (`UserInfoMiddleware` only); sets `user_id` on the event when the session is valid

Swagger UI is wired through [go-web-platform](https://github.com/go-web-services/go-web-platform) (`/swagger` when enabled for your environment).

## Private dependencies

`go-service-user` and `go-service-event` may be private. For Docker builds, pass `GITHUB_TOKEN` as in compose `args`. Locally, configure git/`GOPRIVATE` as in `.env.sample`.
