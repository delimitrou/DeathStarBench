package main

import (
	"fmt"
	"context"
	"log"
	"os"
	"strings"
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

	// video-sharing
	"dapr-apps/video-sharing/common/rating"
	"dapr-apps/video-sharing/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	ratingStore    = util.GetEnvVar("USER_RATING_STORE", "user-rating-store")
)

// max number of trials to update store, quit if exceeded
var maxTry int = 100

// prometheus metric
var (
	readReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "read_user_rating_total",
			Help: "Number of user-rating read requests received.",
		},
	)
	updateReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "update_user_rating_total",
			Help: "Number of user-rating update requests received.",
		},
	)
	readReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "user_read_rate_req_lat_hist",
		Help:    "Latency (ms) histogram of user-rating read requests",
		Buckets: util.LatBuckets(), 
	})
	updateReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "user_update_rate_req_lat_hist",
		Help:    "Latency (ms) histogram of user-rating update requests",
		Buckets: util.LatBuckets(), 
	})
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "user_rating_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading user-rating store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "user_rating_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) user-rating store (kvs/db).",
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
	if err := s.AddServiceInvocationHandler("rate", rateHandler); err != nil {
		log.Fatalf("error adding rateHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func userKey(userId string) string {
	return userId
}

func getHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	readReqCtr.Inc()
	var req rating.GetReq
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
	userkey := userKey(req.UserId)
	item, err := client.GetState(ctx, ratingStore, userkey)
	if err != nil {
		logger.Printf("getHandler GetState err: %s, store: %s, key: %s", 
			err.Error(), ratingStore, userkey)
		return nil, err
	}
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	var ratings rating.UserRating
	var comment string = ""
	var score float64 = 0.0
	var exist bool = false
	if string(item.Value) != "" {
		if err := json.Unmarshal(item.Value, &ratings); err != nil {
			logger.Printf("readHandler json.Umarshal (item.Value) err: %s", err.Error())
			return nil, err
		}
		rate, ok := ratings.Ratings[req.VideoId]
		if ok {
			comment = rate.Comment
			score = rate.Score
			exist = true
		}
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	readReqLatHist.Observe(servLat)
	// create response
	resp := rating.GetResp{
		Exist: exist,
		Comment: comment,
		Score: score,
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

// rateHandler creates rating of a video, or change the current rating
func rateHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateReqCtr.Inc()
	var req rating.RateReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("rateHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// latency metrics
	epoch := time.Now()
	var servLat float64 = float64(epoch.UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("rateHandler dapr err: %s", err.Error())
		return nil, err
	}
	userkey := userKey(req.UserId)
	// loop to update store
	var succ = false
	var loop = 0
	var exist = false
	var oriScore float64 = 0.0
	for ; !succ; {
		loop += 1
		// quit if loop exceeds maxTry
		if loop > maxTry {
			err = errors.New(fmt.Sprintf("rateHandler update key:%s loop exceeds %d rounds, quitted",
				userkey, maxTry))
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, ratingStore, userkey)
		if errl != nil {
			logger.Printf("rateHandler GetState (key: %s) err: %s", userkey, errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var rat rating.UserRating
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &rat); err != nil {
				logger.Printf("rateHandler json.Unmarshal Value (key: %s), err: %s", 
					userkey, errl.Error())
				err = errl
				return nil, err
			}
		} else {
			rat = rating.UserRating {
				Ratings: make(map[string]rating.Rating),
			}
		}
		// fill in new data
		oriRate, ok := rat.Ratings[userkey]
		if ok {
			exist = true
			oriScore = oriRate.Score
		}
		rat.Ratings[req.VideoId] = rating.Rating {
			Comment: req.Comment,
			Score: req.Score,
		}
		// try update store with etag
		ratingjson, errl := json.Marshal(rat)
		if errl != nil {
			logger.Printf("rateHandler json.Marshal (ratingjson) err:%s", errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// try perform store
		newItem := &dapr.SetStateItem{
			Etag: &dapr.ETag{
				Value: etag,
			},
			Key: userkey,
			// Metadata: map[string]string{
			// 	"created-on": time.Now().UTC().String(),
			// },
			Metadata: nil,
			Value: ratingjson,
			Options: &dapr.StateOptions{
				// Concurrency: dapr.StateConcurrencyLastWrite,
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		}
		errl = client.SaveBulkState(ctx, ratingStore, newItem)
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		if errl == nil {
			succ = true
		} else if strings.Contains(errl.Error(), "etag mismatch") {
			// etag mismatch, keeping on trying
			succ = false
		} else {
			// other errors, return
			logger.Printf("rateHandler SaveBulkState (ratingjson) key:%s, err:%s", 
				userkey, errl.Error())
			err = errl
			return nil, err
		}

	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := rating.RateResp{
		Exist: exist,
		OriScore: oriScore,
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("rateHandler json.Marshal (respdata) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}