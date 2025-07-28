FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install build dependencies only in builder stage
RUN apk add --no-cache build-base sqlite-dev

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary with CGO enabled for go-sqlite3
RUN CGO_ENABLED=1 go build -o telegram-bot main.go

# Final minimal image
FROM alpine:3.20

# Install only runtime dependencies
RUN apk add --no-cache sqlite-libs

WORKDIR /app

# Create a non-root user for security
RUN adduser -D -s /bin/sh botuser

# Copy the built binary from builder stage
COPY --from=builder /app/telegram-bot .

# Copy environment file if needed
COPY .env .env

# # Set ownership and switch to non-root user
# RUN chown botuser:botuser telegram-bot .env
# USER botuser

# Use exec form for CMD
CMD ["./telegram-bot", "--stage=prod", "--example-data"]