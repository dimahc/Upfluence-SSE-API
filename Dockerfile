# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

# Final stage
FROM alpine:3.19

# Add ca-certificates for HTTPS requests to SSE stream
RUN apk --no-cache add ca-certificates

WORKDIR /

# Copy binary from builder
COPY --from=builder /server /server

# Non-root user for security
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

ENV ADDR=:8080

ENTRYPOINT ["/server"]
