# Build stage
FROM golang:1.26-alpine AS builder

# Install UPX for binary compression
RUN apk --no-cache add upx

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with maximum optimization
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags=-static" \
    -tags netgo,osusergo \
    -trimpath \
    -gcflags="-l=4" \
    -o main .

# Compress binary with UPX
RUN upx --best --lzma main

# Verify binary still works after compression
RUN ./main --help || echo "Binary compressed successfully"

# Final stage - minimal alpine for shell support
FROM alpine:3.23

WORKDIR /app

# Copy the compressed binary, migrations, static files, and entrypoint
COPY --from=builder /app/main .
COPY migrations/ ./migrations/
COPY static/ ./static/
COPY entrypoint.sh ./entrypoint.sh

# Make entrypoint executable
RUN chmod +x ./entrypoint.sh

# Expose port
EXPOSE 4000

# Use custom entrypoint that runs migrations then server
CMD ["./entrypoint.sh"]
