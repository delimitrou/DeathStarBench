module github.com/harlow/go-micro-services

go 1.21

require (
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.4.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171214222146-0e7658f8ee99
	github.com/hailocab/go-geoindex v0.0.0-20160127134810-64631bfe9711
	github.com/opentracing-contrib/go-stdlib v0.0.0-20180308002341-f6b9967a3c69
	github.com/opentracing/opentracing-go v1.0.2
	github.com/picop-rd/picop-go v0.1.1-0.20231116082607-3419df74ba03
	github.com/rs/zerolog v1.31.0
	github.com/uber/jaeger-client-go v2.11.2+incompatible
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/net v0.18.0
	google.golang.org/grpc v1.59.0
)

require (
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7 // indirect
	github.com/codahale/hdrhistogram v0.9.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/uber-go/atomic v0.0.0-00010101000000-000000000000 // indirect
	github.com/uber/jaeger-lib v1.4.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.opentelemetry.io/otel v1.11.2 // indirect
	go.opentelemetry.io/otel/trace v1.11.2 // indirect
	golang.org/x/crypto v0.15.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace (
	github.com/bradfitz/gomemcache => github.com/picop-rd/gomemcache v1.0.0-picop
	github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.9.0
	github.com/uber-go/atomic => go.uber.org/atomic v1.11.0
	go.mongodb.org/mongo-driver => github.com/picop-rd/mongo-go-driver v1.12.1-picop
	google.golang.org/grpc => github.com/picop-rd/grpc-go v1.0.1-picop
)
