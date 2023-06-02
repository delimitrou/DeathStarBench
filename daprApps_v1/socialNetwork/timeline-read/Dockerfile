# run from root (socialnet) directory
FROM golang:1.17 as builder

COPY timeline-read /src/timeline-read
COPY common /src/common
WORKDIR /src/timeline-read

RUN go mod tidy
RUN go mod vendor

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -tags netgo -mod vendor -o ./service .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /src/timeline-read/service .

ENTRYPOINT ["./service"]
