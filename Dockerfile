FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /app/bin/aogeri ./cmd/api

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/aogeri /usr/local/bin/aogeri

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/aogeri"]
