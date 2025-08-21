# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o intelligent-ai-gateway ./cmd/gateway

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates python3 nodejs npm bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/intelligent-ai-gateway .

# Copy configuration files
COPY providers.csv.template ./
COPY web/admin ./web/admin

# Create directories
RUN mkdir -p scripts generated/providers logs

# Install Python dependencies for scripts
RUN pip3 install requests

EXPOSE 3000

CMD ["./intelligent-ai-gateway"]