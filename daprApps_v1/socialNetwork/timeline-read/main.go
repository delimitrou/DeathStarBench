package main

import (
	"context"
	// "fmt"
	"log"
	"os"
	// "strings"
	"encoding/json"
	// "errors"
	"net/http"
	"sort"
	"time"

	// dapr
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// util
	"dapr-apps/socialnet/common/timeline"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	timelineStore  = util.GetEnvVar("TIMELINE_STORE", "timeline-store")
)

// prometheus metric
var (
	reqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tl_read_req",
			Help: "Number of timeline read requests received.",
		},
	)
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tl_read_req_lat_hist",
		Help:    "Latency (ms) histogram of timeline read requests (since upstream sends the req), excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tl_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading timeline store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(reqCtr)
	prometheus.MustRegister(reqLatHist)
	prometheus.MustRegister(readStoreLatHist)

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
	if err := s.AddServiceInvocationHandler("read", readHandler); err != nil {
		log.Fatalf("error adding readHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func readHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	reqCtr.Inc()
	var req timeline.ReadReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("readHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("readHandler dapr err: %s", err.Error())
		return nil, err
	}
	key := ""
	if req.UserTl {
		key = timeline.UserTlKey(req.UserId)
	} else {
		key = timeline.HomeTlKey(req.UserId)
	}
	// update service latency hist
	epoch := time.Now()
	servLat := epoch.UnixMilli() - req.SendUnixMilli
	
	// query store to get etag and up-to-date val
	item, err := client.GetState(ctx, timelineStore, key)
	if err != nil {
		logger.Printf("readHandler getState err: %s", err.Error())
		return nil, err
	}
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	var tl []string
	if string(item.Value) != "" {
		if err := json.Unmarshal(item.Value, &tl); err != nil {
			logger.Printf("readHandler json.Umarshal (item.Value) err: %s", err.Error())
			return nil, err
		}
	} else {
		tl = make([]string, 0)
	}
	
	// create response
	resp := timeline.ReadResp{
		PostIds: make([]string, 0),
		SendUnixMilli:  time.Now().UnixMilli(),
	}
	// find the required posts
	i := sort.Search(len(tl), func(i int) bool {
		return util.PostIdTime(tl[i]) >= req.EarlUnixMilli
	})
	if i < len(tl) {
		// required posts exist
		if i + req.Posts >= len(tl) {
			resp.PostIds = tl[i:]
		} else {
			resp.PostIds = tl[i:i+req.Posts]
		}
	}
	// update latency metrics
	servLat += time.Now().UnixMilli() - epoch.UnixMilli()
	reqLatHist.Observe(float64(servLat))	
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}
