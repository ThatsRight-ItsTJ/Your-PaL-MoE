# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go workspace and mod files
COPY go.work go.work.sum ./
COPY core/go.mod core/go.sum ./core/
COPY mcp-servers/go.mod mcp-servers/go.sum ./mcp-servers/ 
COPY openapi-mcp/go.mod openapi-mcp/go.sum ./openapi-mcp/

# Download dependencies
RUN cd core && go mod download

# Copy source code
COPY . .

# Build the application
RUN cd core && go build -o ../intelligent-ai-gateway .

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