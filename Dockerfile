# Build stage
#
#
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o bin/chimera ./examples/chimera

# Runtime stage
#
#
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/chimera .

ARG PORT=8080
ARG CONFIG_PATH=/etc/chimera/config.json

# Run the binary
ENTRYPOINT ["/app/chimera"]
CMD ["-config", "${CONFIG_PATH}", "-port", "${PORT}"]
