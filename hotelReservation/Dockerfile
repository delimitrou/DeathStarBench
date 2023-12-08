FROM golang:1.21 as builder

WORKDIR /workspace

COPY go.sum go.sum
COPY go.mod go.mod
COPY vendor/ vendor/

COPY cmd/ cmd/
COPY dialer/ dialer/
COPY registry/ registry/
COPY services/ services/
COPY tls/ tls/
COPY tracing/ tracing/
COPY tune/ tune/

COPY config.json config.json

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install -ldflags="-s -w" -mod=vendor ./cmd/...

