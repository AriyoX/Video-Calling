FROM golang:1.19-bookworm AS builder

WORKDIR /app

# Copy only go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Now copy the rest of the application code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -v -o /app/video-conference ./cmd/server

# Create a minimal production image
FROM debian:bookworm-slim

# Install CA certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/video-conference /app/
COPY --from=builder /app/internal/views /app/internal/views

# Add a non-root user for security
RUN useradd -m videoapp
USER videoapp

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/ || exit 1

# Run the application
CMD ["/app/video-conference"]