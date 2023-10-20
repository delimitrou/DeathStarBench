FROM golang:1.21.3

RUN git config --global http.sslverify false
COPY . /go/src/github.com/harlow/go-micro-services
WORKDIR /go/src/github.com/harlow/go-micro-services
RUN go mod download
RUN go install -ldflags="-s -w" ./cmd/...
