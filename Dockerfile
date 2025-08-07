# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy Go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o ai-assistant ./cmd/main.go

# Stage 2: Run the binary in minimal image
FROM alpine:latest

WORKDIR /root
COPY --from=builder /app/ai-assistant .
COPY .env .

RUN apk add --no-cache ca-certificates

EXPOSE 8081

CMD ["./ai-assistant"]
