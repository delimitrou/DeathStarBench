module dapr-apps/video-sharing/dapr-trending

go 1.17

require (
	dapr-apps/video-sharing/common v0.0.0-00010101000000-000000000000
	github.com/dapr/go-sdk v1.2.0
	github.com/prometheus/client_golang v1.13.0
)

replace dapr-apps/video-sharing/common => ../common
