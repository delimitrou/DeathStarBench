package main

import (
	"fmt"
	"context"
	"log"
	"os"
	// "strings"
	"sort"
	"math/rand"
	"encoding/json"
	"time"
	"errors"
	"net/http"

	// dapr
	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/client"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// socialnet
	"dapr-apps/socialnet/common/recommend"
	"dapr-apps/socialnet/common/socialgraph"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
)

var (
	// max number of follows to take as reference
	maxRef = util.GetEnvVarInt("MAX_REF", 10)
	// max number of users recommended by sorting
	maxRecmdSort = util.GetEnvVarInt("MAX_RECMD_SORT", 10)
	// max number of users recommended by random
	maxRecmdRand = util.GetEnvVarInt("MAX_RECMD_RAND", 10)
)

// prometheus metric
var (
	reqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "recmd_req",
			Help: "Number of recommend requests received",
		},
	)
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "recmd_req_lat_hist",
		Help:    "Latency (ms) histogram of recommend requests (since upstream sends the req), excluding time waiting for rpc resp",
		Buckets: util.LatBuckets(), 
	})
	e2eRecmdLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_recmd_req_lat_hist",
		Help:    "End-to-end latency (ms) histogram of recommend requests",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(reqCtr)
	prometheus.MustRegister(reqLatHist)
	prometheus.MustRegister(e2eRecmdLatHist)

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
	if err := s.AddServiceInvocationHandler("recmd", recmdHandler); err != nil {
		log.Fatalf("error adding recmdHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func recmdHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	reqCtr.Inc()
	var req recommend.Req
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

	// query socialgraph service to get user's follower graph
	userIds := make([]string, 1)
	userIds[0] = req.UserId
	flwReq := socialgraph.GetRecmdReq{
		SendUnixMilli: time.Now().UnixMilli(),
		UserIds: userIds,
		Record: false,
		Latency: int64(0),
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
	flwRespData, err := client.InvokeMethodWithContent(ctx, "dapr-social-graph", "getrecmd", "get", content)
	if err != nil {
		logger.Printf("dapr-social-graph:getrecmd err: %s", err.Error())
		return nil, err
	}	
	// decode flwResp
	var flwResp socialgraph.GetRecmdResp
	if err := json.Unmarshal(flwRespData, &flwResp); err != nil {
		logger.Printf("recmdHanlder json.Unmarshal (flwRespData) err: %s", err.Error())
		return nil, err
	}
	flws, ok := flwResp.FollowIds[req.UserId]
	if !ok {
		logger.Printf("recmdHanlder err: user %s not in flwResp", req.UserId)
		return nil, errors.New(fmt.Sprintf("user %s not in flwResp", req.UserId))
	}
	// select up to maxRef follows to recommend users to follow
	var ref []string
	if len(flws) <= maxRef {
		ref = flws
	} else {
		rand.Shuffle(len(flws), func(i, j int) {
			flws[i], flws[j] = flws[j], flws[i]
		})
		ref = flws[:maxRef]
	}
	// get ref users' follow 
	refReq := socialgraph.GetRecmdReq{
		SendUnixMilli: time.Now().UnixMilli(),
		UserIds: ref,
		Record: true,
		Latency: flwResp.Latency,
	}
	refReqData, err := json.Marshal(refReq)
	if err != nil {
		logger.Printf("recmdHanlder json.Marshal (refReq) err: %s", err.Error())
		return nil, err
	}
	content = &dapr.DataContent{
		ContentType: "application/json",
		Data:        refReqData,
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - flwResp.SendUnixMilli)
	refRespData, err := client.InvokeMethodWithContent(ctx, "dapr-social-graph", "getrecmd", "get", content)	
	if err != nil {
		logger.Printf("dapr-social-graph:getfollow err: %s", err.Error())
		return nil, err
	}
	// decode refResp
	var refResp socialgraph.GetRecmdResp
	if err := json.Unmarshal(refRespData, &refResp); err != nil {
		logger.Printf("recmdHanlder json.Unmarshal (refRespData) err: %s", err.Error())
		return nil, err
	}
	// count number of times each user is followed
	flwCtr := make(map[string]int)
	repeat := make(map[string]bool)
	// won't recommend the user him/herself
	repeat[req.UserId] = true
	cand := make([]string, 0)
	rec := make([]string, 0)
	for _, refFlws := range refResp.FollowIds {
		for _, u := range(refFlws) {
			// exclude users already followed
			if r, ok := repeat[u]; !ok {
				// new user, check if already followed
				in, _ := util.IsValInSlice(u, flws)
				repeat[u] = in
				if !in {
					// _, ok := flwCtr[u]
					// if ok {
					// 	err := errors.New(fmt.Sprintf("user %s exists in flwCtr but not in repeat", u))
					// 	panic(err)
					// }
					flwCtr[u] = 1
					cand = append(cand, u)
				}
			} else if r {
				// ignore already followed user
				continue
			} else {
				// _, ok := flwCtr[u]
				// if !ok {
				// 	err := errors.New(fmt.Sprintf("user %s exists in repeat but not in flwCtr", u))
				// 	panic(err)
				// }
				// update counter
				flwCtr[u] += 1
			}
		}
	}
	// randomly choose recmds
	all := false
	if len(cand) > maxRecmdRand {
		rand.Shuffle(len(cand), func(i, j int) {
			cand[i], cand[j] = cand[j], cand[i]
		})
		rec = cand[:maxRecmdRand]
	} else {
		rec = cand
		all = true
	}
	if len(cand) > maxRecmdSort && !all {
		// sort users by times followed
		type KV struct {
			UserId string
			Ctr int
		}
		var records []KV
		for u, c := range flwCtr {
			records = append(records, KV{UserId: u, Ctr: c})
		}
		// sort in reverse order
		sort.Slice(records, func(i, j int) bool {
			return records[i].Ctr > records[j].Ctr
		})
		for i, kv := range records {
			if i >= maxRecmdSort {
				break
			} else {
				in, _ := util.IsValInSlice(kv.UserId, rec)
				if !in {
					rec = append(rec, kv.UserId)
				}
			}
		}
	} else {
		// if we don't have up to maxRecmdSort users, just recommend all
		rec = cand
	}
	// update service latency hist
	epoch := time.Now()
	servLat += float64(epoch.UnixMilli() - refResp.SendUnixMilli)
	if len(rec) > 0 {
		// update service latency hist
		reqLatHist.Observe(servLat)
		// update end-to-end latency metric
		e2eRecmdLatHist.Observe(float64(epoch.UnixMilli() - req.SendUnixMilli))
	}
	// create response
	resp := recommend.Resp{
		UserIds: rec,
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