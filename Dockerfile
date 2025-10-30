# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/chimera ./examples/chimera

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/bin/chimera .

ENV PORT=8080
EXPOSE ${PORT}
ENTRYPOINT ["/app/chimera"]
