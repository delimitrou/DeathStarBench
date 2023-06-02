package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	// dapr
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// socialnet
	"dapr-apps/video-sharing/common/info"
	"dapr-apps/video-sharing/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	infoStore      = util.GetEnvVar("INFO_STORE", "info-store")
)
// the maximum attempts to save to post-store. Quit the operation if exceeded
var maxTry = 100

// prometheus metric
var (
	uploadCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "info_upload_total",
			Help: "Number of video-info upload requests received.",
		},
	)
	readCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "info_read_total",
			Help: "Number of video-info read requests received.",
		},
	)
	updateCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "info_update_total",
			Help: "Number of video-info update requests received.",
		},
	)
	uploadReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_upload_lat_hist",
		Help:    "Latency (ms) histogram of video-info upload requests.",
		Buckets: util.LatBuckets(), 
	})
	readFrontReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_frontend_read_lat_hist",
		Help:    "Latency (ms) histogram of video-info read requests (frontend requests).",
		Buckets: util.LatBuckets(), 
	})
	readTrendReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_trending_read_lat_hist",
		Help:    "Latency (ms) histogram of video-info read requests (trending requests).",
		Buckets: util.LatBuckets(), 
	})
	updateReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_update_lat_hist",
		Help:    "Latency (ms) histogram of video-info update requests.",
		Buckets: util.LatBuckets(), 
	})
	readFrontStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_frontend_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading info store (kvs/db), frontend requests.",
		Buckets: util.LatBuckets(), 
	})
	readTrendStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_trending_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading info store (kvs/db), trending requests.",
		Buckets: util.LatBuckets(), 
	})
	writeStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_store_write_lat_hist",
		Help:    "Latency (ms) histogram of writing info store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "info_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) info store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(uploadCtr)
	prometheus.MustRegister(readCtr)
	prometheus.MustRegister(updateCtr)
	prometheus.MustRegister(uploadReqLatHist)
	prometheus.MustRegister(readFrontReqLatHist)
	prometheus.MustRegister(readTrendReqLatHist)
	prometheus.MustRegister(updateReqLatHist)
	prometheus.MustRegister(readFrontStoreLatHist)
	prometheus.MustRegister(readTrendStoreLatHist)
	prometheus.MustRegister(writeStoreLatHist)
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
	if err := s.AddServiceInvocationHandler("upload", uploadHandler); err != nil {
		log.Fatalf("error adding uploadHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("rate", rateHandler); err != nil {
		log.Fatalf("error adding rateHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("view", viewHandler); err != nil {
		log.Fatalf("error adding viewHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("info", infoHandler); err != nil {
		log.Fatalf("error adding infoHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// metaKey returns the video meta data store key given a video id
// meta data includes descption of the video and is only updated once
func metaKey(videoId string) string {
	return fmt.Sprintf("meta-%s", videoId)
}
// rateKey returns the rating data store key given a video id
func rateKey(videoId string) string {
	return fmt.Sprintf("rate-%s", videoId)
}
// viewKey returns the view time data store key given a video id
func viewKey(videoId string) string {
	return fmt.Sprintf("view-%s", videoId)
}

// uploadHanlder saves the video metadata into db
func uploadHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	uploadCtr.Inc()
	var req info.UploadReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("uploadHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("uploadHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	metakey := metaKey(req.VideoId)
	meta := info.Meta {
		UserId: req.UserId,
		Resolutions: req.Resolutions,
		Duration: req.Duration,
		Description: req.Description,
		Date: req.Date,
	}
	metajson, err := json.Marshal(meta)
	if err != nil {
		logger.Printf("uploadHandler json.Marshal (metajson) err:%s", err.Error())
		return nil, err
	}
	// update service latency hist
	epoch := time.Now()
	uploadReqLatHist.Observe(float64(epoch.UnixMilli() - req.SendUnixMilli))
	// update post store
	if err := client.SaveState(ctx, infoStore, metakey, metajson); err != nil {
		logger.Printf("uploadHandler saveState (metajson) err: %s", err.Error())
		return nil, err
	}
	// update store latency hist
	writeStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	// create response
	resp := info.Resp{
		VideoId: req.VideoId,
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

// rateHandler updates the rating of a video
func rateHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req info.RateReq
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
	// generate data
	ratekey := rateKey(req.VideoId)
	var scorediff float64
	var scoresqdiff float64
	var numdiff int64
	if req.Change {
		scorediff = req.Score - req.OriScore
		scoresqdiff = math.Pow(req.Score, 2) - math.Pow(req.OriScore, 2)
		numdiff = 0
	} else {
		scorediff = req.Score
		scoresqdiff = math.Pow(req.Score, 2)
		numdiff = 1
	}
	// loop to update store
	var succ = false
	var loop = 0
	for ; !succ; {
		loop += 1
		// quit if loop exceeds maxTry
		if loop > maxTry {
			err = errors.New(fmt.Sprintf("rateHandler update key:%s loop exceeds %d rounds, quitted",
				ratekey, maxTry))
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, infoStore, ratekey)
		if errl != nil {
			logger.Printf("rateHandler GetState (key: %s) err: %s", ratekey, errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var rating info.Rating
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &rating); err != nil {
				logger.Printf("rateHandler json.Unmarshal Value (key: %s), err: %s", 
					ratekey, errl.Error())
				err = errl
				return nil, err
			}
		} else {
			rating = info.Rating{
				Num: 0,
				Score: 0.0,
				ScoreSq: 0.0,
			}
		}
		// fill in new data
		oriNum := rating.Num
		rating.Num = rating.Num + numdiff
		if rating.Num == 0 {
			logger.Printf("rateHandler err: rating.Num = 0 (key: %s)", ratekey)
			err = errors.New(fmt.Sprintf("rateHandler rating.Num = 0 (key: %s)", ratekey))
			return nil, err
		}
		rating.Score = (float64(oriNum)*rating.Score + scorediff)/float64(rating.Num)
		rating.ScoreSq = (float64(oriNum)*rating.ScoreSq + scoresqdiff)/float64(rating.Num)
		// try update store with etag
		ratingjson, errl := json.Marshal(rating)
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
			Key: ratekey,
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
		errl = client.SaveBulkState(ctx, infoStore, newItem)
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
				ratekey, errl.Error())
			err = errl
			return nil, err
		}

	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := info.Resp{
		VideoId: req.VideoId,
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

// viewHandler updates times of view of a video
func viewHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req info.ViewReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("viewHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// latency metrics
	epoch := time.Now()
	var servLat float64 = float64(epoch.UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("viewHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	viewkey := viewKey(req.VideoId)
	// loop to update store
	var succ = false
	var loop = 0
	for ; !succ; {
		loop += 1
		// quit if loop exceeds maxTry
		if loop > maxTry {
			err = errors.New(fmt.Sprintf("viewHandler update key: %s loop exceeds %d rounds, quitted",
				viewkey, maxTry))
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, infoStore, viewkey)
		if errl != nil {
			logger.Printf("viewHandler GetState (key: %s) err: %s", viewkey, errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var views int64
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &views); err != nil {
				logger.Printf("viewHandler json.Unmarshal Value (key: %s), err: %s", 
					viewkey, errl.Error())
				err = errl
				return nil, err
			}
		} else {
			views = 0
		}
		views += 1
		// try update store with etag
		viewsjson, errl := json.Marshal(views)
		if errl != nil {
			logger.Printf("viewHandler json.Marshal (viewsjson) err:%s", errl.Error())
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
			Key: viewkey,
			// Metadata: map[string]string{
			// 	"created-on": time.Now().UTC().String(),
			// },
			Metadata: nil,
			Value: viewsjson,
			Options: &dapr.StateOptions{
				// Concurrency: dapr.StateConcurrencyLastWrite,
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		}
		errl = client.SaveBulkState(ctx, infoStore, newItem)
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		if errl == nil {
			succ = true
		} else if strings.Contains(errl.Error(), "etag mismatch") {
			// etag mismatch, keeping on trying
			succ = false
		} else {
			// other errors, return
			logger.Printf("viewHandler SaveBulkState (viewsjson) key:%s, err:%s", 
				viewkey, errl.Error())
			err = errl
			return nil, err
		}

	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := info.Resp{
		VideoId: req.VideoId,
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("viewHandler json.Marshal (respdata) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// infoHandler reads meta data of a video
func infoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	readCtr.Inc()
	var req info.InfoReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("infoHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("infoHandler dapr err: %s", err.Error())
		return nil, err
	}
	// query store to get all fields of required videos
	keys := make([]string, 0)
	for _, vid  := range req.VideoIds {
		metakey := metaKey(vid)
		ratekey := rateKey(vid)
		viewkey := viewKey(vid)
		keys = append(keys, metakey, ratekey, viewkey)
	}
	// update service latency hist
	epoch := time.Now()
	servLat := float64(epoch.UnixMilli() - req.SendUnixMilli)
	items, err := client.GetBulkState(ctx, infoStore, keys, nil, int32(len(keys)))
	// update store latency hist
	if req.Upstream == "frontend" {
		readFrontStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	} else if req.Upstream == "trending" {
		readTrendStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	} else {
		logger.Printf("infoHandler Unknown upstream: %s", req.Upstream)
		return nil, err
	}
		
	epoch = time.Now()
	if err != nil {
		logger.Printf("infoHandler getBulkState err: %s", err.Error())
		return nil, err
	}
	if len(items) != len(keys) {
		logger.Printf("infoHandler getBulkState len(items) != len(keys)")
		return nil, errors.New(fmt.Sprintf("store %s: getBulkState len(items) != len(keys)", infoStore))
	}
	// combine fields into complete videos
	videos := make(map[string]info.Info)
	itMap := make(map[string]*dapr.BulkStateItem)
	for _, it := range(items) {
		itMap[it.Key] = it
	}
	for _, vid := range req.VideoIds {
		metakey := metaKey(vid)
		ratekey := rateKey(vid)
		viewkey := viewKey(vid)
		// empty info
		v := info.Info {
			VideoMeta: info.Meta {
				UserId: "missing",
				Resolutions: make([]string, 0),
				Duration: 0.0,
				Description: "missing",
			},
			Views: 0,
			Rate: info.Rating {
				Num: 0,
				Score: 0.0,
			},
		}
		// unmarshal meta
		if string(itMap[metakey].Value) != "" {
			var meta info.Meta
			if err = json.Unmarshal(itMap[metakey].Value, &meta); err != nil {
				logger.Printf("infoHandler json.Unmarshal (meta) for key: %s, err: %s", 
					metakey, err.Error())
				return nil, err
			} else {
				v.VideoMeta = meta
			}
		} else {
			// missing video, return empty info
			videos[vid] = v
			continue
		}
		// unmarshal views
		if string(itMap[viewkey].Value) != "" {
			var view int64
			if err = json.Unmarshal(itMap[viewkey].Value, &view); err != nil {
				logger.Printf("infoHandler json.Unmarshal (view) for key: %s, err: %s", 
					viewkey, err.Error())
				return nil, err
			} else {
				v.Views = view
			}
		} 
		// unmarshal rates
		if string(itMap[ratekey].Value) != "" {
			var rate info.Rating
			if err = json.Unmarshal(itMap[ratekey].Value, &rate); err != nil {
				logger.Printf("infoHandler json.Unmarshal (rate Value) for key: %s, err: %s", 
					ratekey, err.Error())
				return nil, err
			} else {
				v.Rate = rate
			}
		}
		// add this post to the map
		videos[vid] = v
	}
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	if req.Upstream == "frontend" {
		readFrontReqLatHist.Observe(servLat)
	} else if req.Upstream == "trending" {
		readTrendReqLatHist.Observe(servLat)
	} else {
		logger.Printf("infoHandler Unknown upstream: %s", req.Upstream)
		return nil, err
	}
	// create response
	resp := info.InfoResp{
		VideoInfo: videos,
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("infoHandler json.Marshal (resp) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}