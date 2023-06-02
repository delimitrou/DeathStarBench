# run from root (video-sharing) directory
FROM golang:1.19 as builder

COPY proxy /src/proxy
COPY common /src/common
WORKDIR /src/proxy

RUN go mod tidy
RUN go mod vendor

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo -mod vendor -o ./service .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /src/proxy/service .

ENTRYPOINT ["./service"]
