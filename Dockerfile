# Stage 1: Build Stage
FROM golang:1.26-alpine AS builder

# Define build arguments for private module access
ARG GITHUB_TOKEN

ENV GO111MODULE=on \
    GOPRIVATE=github.com

# Install dependencies and configure git in one layer
RUN apk add --no-cache git ca-certificates tzdata && \
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o go-gateway-template ./cmd/app/main.go

# Stage 2: Final Stage
FROM alpine:3.22 AS final

WORKDIR /app

COPY --from=builder /app/go-gateway-template .

CMD ["./go-gateway-template"]
