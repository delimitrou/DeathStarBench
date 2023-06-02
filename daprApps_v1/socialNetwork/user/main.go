package main

import (
	// "fmt"
	"context"
	"log"
	"os"
	// "strings"
	// "sort"
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

	// socialnet
	"dapr-apps/socialnet/common/user"
	"dapr-apps/socialnet/common/socialgraph"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
)

// prometheus metric
var (
	ReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_req",
			Help: "Number of user requests received",
		},
	)
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "user_req_lat_hist",
		Help:    "Latency (ms) histogram of user requests (since upstream sends the req), excluding time waiting for rpc resp",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(ReqCtr)
	prometheus.MustRegister(reqLatHist)

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
	if err := s.AddServiceInvocationHandler("register", regHandler); err != nil {
		log.Fatalf("error adding regHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func regHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer ReqCtr.Inc()
	var req user.RegisterReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("recmdHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("recmdHandler dapr err: %s", err.Error())
		return nil, err
	}
	// user follows him/herself so that his/her own posts appear in timeline
	flwReq := socialgraph.FollowReq{
		UserId: req.UserId,
		FollowId: req.UserId,
		SendUnixMilli: time.Now().UnixMilli(),
	}
	flwReqData, err := json.Marshal(flwReq)
	if err != nil {
		logger.Printf("recmdHanlder json.Marshal (flwReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        flwReqData,
	}
	// update service latency hist
	servLat := float64(time.Now().UnixMilli() - req.SendUnixMilli)
	// call social graph
	flwRespData, err := client.InvokeMethodWithContent(ctx, "dapr-social-graph", "follow", "post", content)
	if err != nil {
		logger.Printf("dapr-social-graph:follow err: %s", err.Error())
		return nil, err
	}	
	// decode flwResp
	var flwResp socialgraph.UpdateResp
	if err := json.Unmarshal(flwRespData, &flwResp); err != nil {
		logger.Printf("recmdHanlder json.Unmarshal (flwRespData) err: %s", err.Error())
		return nil, err
	}

	servLat += float64(time.Now().UnixMilli() - flwResp.SendUnixMilli)
	// update service latency hist
	reqLatHist.Observe(servLat)
	// create response
	resp := user.RegisterResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("recmdHanlder json.Marshal (respdata) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}