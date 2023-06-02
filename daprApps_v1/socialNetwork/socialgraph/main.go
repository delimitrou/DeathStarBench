package main

import (
	"fmt"
	"context"
	"log"
	"os"
	// "strings"
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
	"dapr-apps/socialnet/common/socialgraph"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	graphStore     = util.GetEnvVar("SOCIAL_GRAPH_STORE", "social-grpah-store")
)

// prometheus metric
var (
	readCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "socialgraph_read_req",
			Help: "Number of socialgraph read requests received.",
		},
	)
	recmdCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "socialgraph_recmd_req",
			Help: "Number of socialgraph recmd requests received.",
		},
	)
	updateCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "socialgraph_update_req",
			Help: "Number of socialgraph update requests received.",
		},
	)
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "socialgraph_req_lat_hist",
		Help:    "Latency (ms) histogram of socialgraph requests (since upstream sends the req), excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	recmdLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "socialgraph_recmd_lat_hist",
		Help:    "Latency (ms) histogram of socialgraph recmd requests (since upstream sends the req), excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "socialgraph_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading socialgraph store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "socialgraph_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) socialgraph store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(readCtr)
	prometheus.MustRegister(recmdCtr)
	prometheus.MustRegister(updateCtr)
	prometheus.MustRegister(reqLatHist)
	prometheus.MustRegister(recmdLatHist)
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
	if err := s.AddServiceInvocationHandler("getfollow", getFollowHandler); err != nil {
		log.Fatalf("error adding getFollowHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("getrecmd", getRecmdHandler); err != nil {
		log.Fatalf("error adding getRecmdHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("getfollower", getFollowerHandler); err != nil {
		log.Fatalf("error adding getFollowerHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("follow", followHandler); err != nil {
		log.Fatalf("error adding followHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("unfollow", unfollowHandler); err != nil {
		log.Fatalf("error adding unfollowHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getFollowHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req socialgraph.GetReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("getFollowHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	if len(req.UserIds) == 0 {
		resp := socialgraph.GetFollowResp{
			SendUnixMilli:  time.Now().UnixMilli(),
			FollowIds: make(map[string][]string),
		}
		respdata, _ := json.Marshal(resp)
		out = &common.Content{
			ContentType: "application/json",
			Data:        respdata,
		}
		return
	}
	readCtr.Inc()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("getFollowHanlder dapr err: %s", err.Error())
		return nil, err
	}
	// query store, first change keys to query follow
	keys := make([]string, len(req.UserIds))
	for i, _ := range req.UserIds {
		keys[i] = util.FollowKey(req.UserIds[i])
	}
	// update service latency hist
	epoch := time.Now()
	servLat := float64(epoch.UnixMilli() - req.SendUnixMilli)
	items, err := client.GetBulkState(ctx, graphStore, keys, nil, int32(len(req.UserIds)))
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	if err != nil {
		logger.Printf("getFollowHanlder getBulkState err: %s", err.Error())
		return nil, err
	}
	if len(items) != len(req.UserIds) {
		logger.Printf("getFollowHanlder getBulkState len(items) != len(req.UserIds)")
		return nil, errors.New(fmt.Sprintf("store %s: getBulkState len(items) != len(req.UserIds)", graphStore))
	}
	itMap := make(map[string]*dapr.BulkStateItem)
	for _, it := range items {
		itMap[it.Key] = it
	}
	followIds := make(map[string][]string)
	for _, userid:= range(req.UserIds) {
		key := util.FollowKey(userid)
		if string(itMap[key].Value) != "" {
			var flws []string
			if err := json.Unmarshal(itMap[key].Value, &flws); err != nil {
				logger.Printf("getFollowHanlder json unmarshal (item.Value) for key: %s, err: %s", 
					key, err.Error())
				followIds[userid] = make([]string, 0)
			} else {
				followIds[userid] = flws
			}
		} else {
			followIds[userid] = make([]string, 0)
		}
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	reqLatHist.Observe(servLat)
	// create response
	resp := socialgraph.GetFollowResp{
        SendUnixMilli:  time.Now().UnixMilli(),
        FollowIds: followIds,
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

func getRecmdHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req socialgraph.GetRecmdReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("getRecmdHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	if len(req.UserIds) == 0 {
		resp := socialgraph.GetRecmdResp{
			SendUnixMilli:  time.Now().UnixMilli(),
			FollowIds: make(map[string][]string),
			Latency: int64(0),
		}
		respdata, _ := json.Marshal(resp)
		out = &common.Content{
			ContentType: "application/json",
			Data:        respdata,
		}
		return
	}
	recmdCtr.Inc()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("getRecmdHanlder dapr err: %s", err.Error())
		return nil, err
	}
	// query store, first change keys to query follow
	keys := make([]string, len(req.UserIds))
	for i, _ := range req.UserIds {
		keys[i] = util.FollowKey(req.UserIds[i])
	}
	// update service latency hist
	epoch := time.Now()
	servLat := epoch.UnixMilli() - req.SendUnixMilli
	if req.Record {
		servLat += req.Latency
	}
	items, err := client.GetBulkState(ctx, graphStore, keys, nil, int32(len(req.UserIds)))
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	if err != nil {
		logger.Printf("getRecmdHanlder getBulkState err: %s", err.Error())
		return nil, err
	}
	if len(items) != len(req.UserIds) {
		logger.Printf("getRecmdHanlder getBulkState len(items) != len(req.UserIds)")
		return nil, errors.New(fmt.Sprintf("store %s: getBulkState len(items) != len(req.UserIds)", graphStore))
	}
	itMap := make(map[string]*dapr.BulkStateItem)
	for _, it := range items {
		itMap[it.Key] = it
	}
	followIds := make(map[string][]string)
	for _, userid:= range(req.UserIds) {
		key := util.FollowKey(userid)
		if string(itMap[key].Value) != "" {
			var flws []string
			if err := json.Unmarshal(itMap[key].Value, &flws); err != nil {
				logger.Printf("getRecmdHanlder json unmarshal (item.Value) for key: %s, err: %s", 
					key, err.Error())
				followIds[userid] = make([]string, 0)
			} else {
				followIds[userid] = flws
			}
		} else {
			followIds[userid] = make([]string, 0)
		}
	}
	// update service latency hist
	servLat += time.Now().UnixMilli() - epoch.UnixMilli()
	if req.Record {
		recmdLatHist.Observe(float64(servLat))
	}
	// create response
	resp := socialgraph.GetRecmdResp{
        SendUnixMilli:  time.Now().UnixMilli(),
        FollowIds: followIds,
		Latency: servLat,
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

func getFollowerHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	readCtr.Inc()
	var req socialgraph.GetReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("getFollowerHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("getFollowerHanlder dapr err: %s", err.Error())
		return nil, err
	}
	// query store, first change keys to query followers
	keys := make([]string, len(req.UserIds))
	for i, _ := range req.UserIds {
		keys[i] = util.FollowerKey(req.UserIds[i])
	}
	// update service latency hist
	epoch := time.Now()
	servLat := float64(epoch.UnixMilli() - req.SendUnixMilli)
	items, err := client.GetBulkState(ctx, graphStore, keys, nil, int32(len(req.UserIds)))
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	if err != nil {
		logger.Printf("getFollowerHanlder getBulkState err: %s", err.Error())
		return nil, err
	}
	if len(items) != len(req.UserIds) {
		logger.Printf("getFollowerHanlder getBulkState len(items) != len(req.UserIds)")
		return nil, errors.New(fmt.Sprintf("store %s: getBulkState len(items) != len(req.UserIds)", graphStore))
	}
	itMap := make(map[string]*dapr.BulkStateItem)
	for _, it := range items {
		itMap[it.Key] = it
	}
	followerIds := make(map[string][]string)
	for _, userid:= range(req.UserIds) {
		key := util.FollowerKey(userid)
		if string(itMap[key].Value) != "" {
			var flwers []string
			if err := json.Unmarshal(itMap[key].Value, &flwers); err != nil {
				logger.Printf("getFollowerHanlder json unmarshal (item.Value) for key: %s, err: %s", 
					key, err.Error())
				followerIds[userid] = make([]string, 0)
			} else {
				followerIds[userid] = flwers
			}
		} else {
			followerIds[userid] = make([]string, 0)
		}
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	reqLatHist.Observe(servLat)
	// create response
	resp := socialgraph.GetFollowerResp{
        SendUnixMilli:  time.Now().UnixMilli(),
        FollowerIds: followerIds,
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// followHandler performs the follow operation
// It adds the followees to user's follow list
// and also adds the user to new followees' followers list
func followHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req socialgraph.FollowReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("followHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// update service latency hist
	var servLat float64 = float64(time.Now().UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	
	// update follow list of user
	key := util.FollowKey(req.UserId)
	succ, servl, storel, err := util.UpdateStoreSlice(ctx, graphStore, key, req.FollowId, true, 0, logger)
	servLat += servl
	storeLat += storel
	if !succ {
		return nil, err
	}
	// update follower list of followee
	key = util.FollowerKey(req.FollowId)
	succ, servl, storel, err = util.UpdateStoreSlice(ctx, graphStore, key, req.UserId, true, 0, logger)
	servLat += servl
	storeLat += storel
	// update service latency hist
	reqLatHist.Observe(servLat)
	// update store latency hist
	updateStoreLatHist.Observe(storeLat)
	if !succ {
		return nil, err
	}
	// create response
	resp := socialgraph.UpdateResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("followHanlder json.Marshal (resp) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// followHandler performs the unfollow operation
// It removes the followees from user's follow list
// and also removes the user from the followees' followers list
func unfollowHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req socialgraph.UnfollowReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("unfollowHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// update service latency hist
	var servLat float64 = float64(time.Now().UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	
	// update follow list of user
	key := util.FollowKey(req.UserId)
	succ, servl, storel, err := util.UpdateStoreSlice(ctx, graphStore, key, req.UnfollowId, false, 0, logger)
	servLat += servl
	storeLat += storel
	if !succ {
		return nil, err
	}
	// update follower list of followee
	key = util.FollowerKey(req.UnfollowId)
	succ, servl, storel, err = util.UpdateStoreSlice(ctx, graphStore, key, req.UserId, false, 0, logger)
	servLat += servl
	storeLat += storel
	// update service latency hist
	reqLatHist.Observe(servLat)
	// update store latency hist
	updateStoreLatHist.Observe(storeLat)
	if !succ {
		return nil, err
	}
	// create response
	resp := socialgraph.UpdateResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("unfollowHanlder json.Marshal (resp) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}
