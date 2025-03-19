# Build arguments
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Build stage
FROM --platform=linux/amd64 golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o vm-server

# Final stage
FROM --platform=$TARGETOS/$TARGETARCH alpine:3.18

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/vm-server .

# Copy default template.yaml
COPY template.yaml /app/template.yaml

# Create a non-root user
RUN adduser -D -u 1000 appuser
RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/vm-server"]