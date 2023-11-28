module github.com/harlow/go-micro-services

go 1.21

require (
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/golang/protobuf v1.0.0
	github.com/google/uuid v1.3.1
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171214222146-0e7658f8ee99
	github.com/hailocab/go-geoindex v0.0.0-20160127134810-64631bfe9711
	github.com/opentracing-contrib/go-stdlib v0.0.0-20180308002341-f6b9967a3c69
	github.com/opentracing/opentracing-go v1.0.2
	github.com/rs/zerolog v1.31.0
	github.com/uber/jaeger-client-go v2.11.2+incompatible
	go.mongodb.org/mongo-driver v1.12.1
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b
	google.golang.org/grpc v1.10.0
)

require (
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7 // indirect
	github.com/codahale/hdrhistogram v0.9.0 // indirect
	github.com/golang/glog v1.1.2 // indirect
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
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20180306020942-df60624c1e9b // indirect
)

replace (
	github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.9.0
	github.com/uber-go/atomic => go.uber.org/atomic v1.11.0
)
