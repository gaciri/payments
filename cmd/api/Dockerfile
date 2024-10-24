
FROM golang:1.23.2-bullseye AS builder
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o api_service ./cmd/api

FROM debian:bullseye-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/api_service /app/api_service
CMD ["/app/api_service"]


