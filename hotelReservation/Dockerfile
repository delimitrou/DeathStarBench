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

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install -ldflags="-s -w" -mod=vendor ./cmd/...

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /go/bin/frontend .
COPY --from=builder /go/bin/geo .
COPY --from=builder /go/bin/profile .
COPY --from=builder /go/bin/rate .
COPY --from=builder /go/bin/recommendation .
COPY --from=builder /go/bin/reservation .
COPY --from=builder /go/bin/search .
COPY --from=builder /go/bin/user .
