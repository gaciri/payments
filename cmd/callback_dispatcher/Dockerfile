
FROM golang:1.23.2-bullseye AS builder
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o dispatcher_service ./cmd/callback_dispatcher

FROM debian:bullseye-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/dispatcher_service /app/dispatcher_service
CMD ["/app/dispatcher_service"]


