# ===========================
# 1. Builder Stage
# ===========================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

# Download any existing deps
RUN go mod download

# Copy the rest of the source code
COPY . .

# Ensure Prometheus modules are added and synced
RUN go get github.com/prometheus/client_golang@v1.20.4 && \
    go mod tidy

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service .

# ===========================
# 2. Runner Stage
# ===========================
FROM alpine:3.20 AS runner

WORKDIR /app

COPY --from=builder /app/auth-service /app/auth-service

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 4000

ENTRYPOINT ["/app/auth-service"]