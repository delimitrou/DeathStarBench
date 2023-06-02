package main

import (
	"fmt"
	"context"
	"log"
	"os"
	"strings"
	"encoding/json"
	"time"
	// "errors"
	"net/http"

	// dapr
	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/client"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// video-sharing
	"dapr-apps/video-sharing/common/dates"
	"dapr-apps/video-sharing/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	dateStore    = util.GetEnvVar("DATE_STORE", "date-store")
)

// max number of trials to update store, quit if exceeded
var maxTry int = 100

// prometheus metric
var (
	readReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dates_read_total",
			Help: "Number of datestore read requests received.",
		},
	)
	updateReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "dates_update_total",
			Help: "Number of datestore update requests received.",
		},
	)
	readReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dates_read_req_lat_hist",
		Help:    "Latency (ms) histogram of dates read requests",
		Buckets: util.LatBuckets(), 
	})
	updateReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dates_update_req_lat_hist",
		Help:    "Latency (ms) histogram of dates update requests",
		Buckets: util.LatBuckets(), 
	})
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dates_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading date store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dates_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) date store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(readReqCtr)
	prometheus.MustRegister(updateReqCtr)
	prometheus.MustRegister(readReqLatHist)
	prometheus.MustRegister(updateReqLatHist)
	prometheus.MustRegister(readStoreLatHist)
	prometheus.MustRegister(updateStoreLatHist)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	logger.Fatal(http.ListenAndServe(promAddress, nil))
}

func main() {
	// prometheus monitor
	go setup_prometheus()
	// business logic 
	// create serving server
	s, err := daprd.NewService(serviceAddress)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
	// add handlers to the service
	if err := s.AddServiceInvocationHandler("get", getHandler); err != nil {
		log.Fatalf("error adding getHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("upload", uploadHandler); err != nil {
		log.Fatalf("error adding uploadHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
// dateKey converts a date string into store key
func dateKey(d string) string {
	return d
}
func dateKeys(ds []string) []string {
	keys := make([]string, 0)
	for _, d := range(ds) {
		keys = append(keys, dateKey(d))
	}
	return keys
}
// keyDate converts a store key to the corresponding date
func keyDate(k string) string {
	return k
}
// getHandler gets the list of videos uploaded in a date
func getHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	readReqCtr.Inc()
	var req dates.GetReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("getHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("getHandler dapr err: %s", err.Error())
		return nil, err
	}
	// update service latency hist
	epoch := time.Now()
	servLat := float64(epoch.UnixMilli() - req.SendUnixMilli)
	datekeys := dateKeys(req.Dates)
	items, err := client.GetBulkState(ctx, dateStore, datekeys, nil, int32(len(datekeys)))
	if err != nil {
		logger.Printf("getHandler GetBulkState err: %s, store: %s, key: %s", 
			err.Error(), dateStore, strings.Join(datekeys, ","))
		return nil, err
	}
	if len(items) != len(datekeys) {
		logger.Printf("getHandler getBulkState len(items) != len(datekeys)")
		return nil, fmt.Errorf("store %s: getBulkState len(items) != len(datekeys)", dateStore)
	}
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	videos := make(map[string][]string)
	for _, it := range items {
		// unmarshal meta
		if string(it.Value) != "" {
			var dv []string
			if err = json.Unmarshal(it.Value, &dv); err != nil {
				logger.Printf("getHandler json.Unmarshal (dv) for key: %s, err: %s", 
					it.Key, err.Error())
				return nil, err
			} else {
				videos[keyDate(it.Key)] = dv
			}
		} else {
			videos[keyDate(it.Key)] = make([]string, 0)
		}
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	readReqLatHist.Observe(servLat)
	// create response
	resp := dates.GetResp{
		Videos: videos,
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("readHandler json.Umarshal (respdata) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// uploadHandler adds a new video to the list of videos of a date
func uploadHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateReqCtr.Inc()
	var req dates.UploadReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("uploadHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// latency metrics
	epoch := time.Now()
	var servLat float64 = float64(epoch.UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// update store
	datekeys := dateKey(req.Date)
	succ, servl, storel, err := util.UpdateStoreSlice(ctx, dateStore, datekeys, req.VideoId, true, 0, logger)
	servLat += servl
	storeLat += storel
	if !succ {
		return nil, err
	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := dates.UploadResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("uploadHandler json.Marshal (respdata) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}