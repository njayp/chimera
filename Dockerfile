# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o bin/chimera ./examples/chimera

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/bin/chimera .

ENV PORT=8080
ENV CONFIG_PATH
EXPOSE ${PORT}
ENTRYPOINT ["/app/chimera"]
