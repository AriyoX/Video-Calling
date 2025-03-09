FROM golang:1.19-bookworm AS builder

WORKDIR /app
COPY . .

# Download dependencies
RUN go mod download && go mod verify

# Build the application - specifying the entry point
RUN go build -v -o /app/video-conference ./cmd/server

# Create a minimal production image
FROM debian:bookworm-slim

# Install CA certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/video-conference /app/
COPY --from=builder /app/internal/views /app/internal/views

EXPOSE 8080

# Run the application
CMD ["/app/video-conference"]