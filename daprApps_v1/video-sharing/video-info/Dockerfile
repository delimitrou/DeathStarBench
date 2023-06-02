# run from root (video-sharing) directory
FROM golang:1.17 as builder

COPY video-info /src/video-info
COPY common /src/common
WORKDIR /src/video-info

RUN go mod tidy
RUN go mod vendor

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo -mod vendor -o ./service .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /src/video-info/service .

ENTRYPOINT ["./service"]
