FROM golang:1.24 as builder

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

ENV SEARCH_AWS_URL=https://k2coxpcvdti43frihmo3gytf6q0ilojc.lambda-url.us-east-2.on.aws/
ENV RESERVATION_AWS_URL=https://pfsu72jzcsqtbwsnrhaadtyh5a0uhpci.lambda-url.us-east-2.on.aws/
ENV RECOMMENDATION_AWS_URL=https://ryeeszilkopelj5yzygg75kt4y0jbwob.lambda-url.us-east-2.on.aws/
ENV PROFILE_AWS_URL=https://de53neoj2avvgjfgvauhmsmlli0eactb.lambda-url.us-east-2.on.aws/
ENV USER_AWS_URL=https://hklpcrrfrfbo3cx2gpayzo6lty0ubgnc.lambda-url.us-east-2.on.aws/
ENV GEO_AWS_URL=https://ivwdmxjk2d7ey73hea7lxbeb2y0dfgmu.lambda-url.us-east-2.on.aws/
ENV RATE_AWS_URL=https://rjybrykceun2wjgz7bj54hx3x40phpei.lambda-url.us-east-2.on.aws/

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on GOBIN=/workspace go install -ldflags="-s -w" -mod=vendor ./cmd/...

