package main

import (
	// "fmt"
	"context"
	"log"
	"os"
	// "strings"
	"encoding/json"
	"time"
	// "errors"
	"net/http"
	"encoding/base64"

	// dapr
	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/client"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// socialnet
	"dapr-apps/socialnet/common/frontend"
	"dapr-apps/socialnet/common/post"
	"dapr-apps/socialnet/common/objectdetect"
	"dapr-apps/socialnet/common/sentiment"
	"dapr-apps/socialnet/common/timeline"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	// image store
	imageStore     = util.GetEnvVar("IMAGE_STORE", "image-store")
	// object-detect
	objectPubsub   = util.GetEnvVar("OBJECT_DETECT_PUBSUB", "object-detect-pubsub")
	objectTopic    = util.GetEnvVar("OBJECT_TOPIC", "object-detect")
	// sentiment
	sentiPubsub    = util.GetEnvVar("SENTIMENT_PUBSUB", "sentiment-pubsub")
	sentiTopic     = util.GetEnvVar("SENTIMENT_TOPIC", "sentiment")
	// timeline-update
	timelinePubsub = util.GetEnvVar("TIMELINE_PUBSUB", "timeline-events")
	timelineTopic  = util.GetEnvVar("TIMELINE_TOPIC", "timeline")
)

// prometheus metric
var (
	// flow counters
	saveReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_save_req",
			Help: "Number of frontend save requests received.",
		},
	)
	updateReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_update_req",
			Help: "Number of frontend update (comment & upvote) requests received.",
		},
	)
	imageReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_image_req",
			Help: "Number of frontend image read requests received.",
		},
	)
	tlReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_tl_req",
			Help: "Number of frontend timeline read requests received.",
		},
	)
	sentiReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_senti_req",
			Help: "Number of sentiment-analysis requests sent by frontend.",
		},
	)
	objDetReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_object_detect_req",
			Help: "Number of object-detect requests sent by frontend.",
		},
	)
	tlUpdateReqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "frontend_tl_update_req",
			Help: "Number of timeline update requests sent by frontend.",
		},
	)
	// latency histograms
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "frontend_req_lat_hist",
		Help:    "Latency (ms) histogram of frontend requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	imgReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "frontend_img_req_lat_hist",
		Help:    "Latency (ms) histogram of frontend image requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	saveReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "frontend_save_req_lat_hist",
		Help:    "Latency (ms) histogram of frontend save requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	saveReqImgLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "frontend_save_img_req_lat_hist",
		Help:    "Latency (ms) histogram of frontend save (with image) requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	readTlReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "frontend_read_tl_req_lat_hist",
		Help:    "Latency (ms) histogram of frontend read timeline requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	// state store latency
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "img_store_front_read_lat_hist",
		Help:    "Latency (ms) histogram of reading img store (kvs/db) by frontend.",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "img_store_front_update_lat_hist",
		Help:    "Latency (ms) histogram of writing img store (kvs/db) by frontend.",
		Buckets: util.LatBuckets(), 
	})
	// end to end latency metrics
	e2ePostSaveLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_post_save_lat_hist",
		Help:    "End-to-end latency (ms) histogram of post save request.",
		Buckets: util.LatBuckets(), 
	})
	e2ePostSaveImgLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_post_save_img_lat_hist",
		Help:    "End-to-end latency (ms) histogram of post save (with images) request.",
		Buckets: util.LatBuckets(), 
	})
	e2ePostUpdateLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_post_update_lat_hist",
		Help:    "End-to-end latency (ms) histogram of post update (comment, upvote, delete) request.",
		Buckets: util.LatBuckets(), 
	})
	e2eImgLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_img_lat_hist",
		Help:    "End-to-end latency (ms) histogram of image request.",
		Buckets: util.LatBuckets(), 
	})
	e2eReadTlLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_read_tl_lat_hist",
		Help:    "End-to-end latency (ms) histogram of timeline read requests.",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(saveReqCtr)
	prometheus.MustRegister(updateReqCtr)
	prometheus.MustRegister(imageReqCtr)
	prometheus.MustRegister(tlReqCtr)
	prometheus.MustRegister(sentiReqCtr)
	prometheus.MustRegister(objDetReqCtr)
	prometheus.MustRegister(tlUpdateReqCtr)
	// service latency
	prometheus.MustRegister(reqLatHist)
	prometheus.MustRegister(imgReqLatHist)
	prometheus.MustRegister(saveReqLatHist)
	prometheus.MustRegister(saveReqImgLatHist)
	prometheus.MustRegister(readTlReqLatHist)
	// state store latency
	prometheus.MustRegister(readStoreLatHist)
	prometheus.MustRegister(updateStoreLatHist)
	// end to end latency
	prometheus.MustRegister(e2ePostSaveLatHist)
	prometheus.MustRegister(e2ePostSaveImgLatHist)
	prometheus.MustRegister(e2ePostUpdateLatHist)
	prometheus.MustRegister(e2eImgLatHist)
	prometheus.MustRegister(e2eReadTlLatHist)

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
	if err := s.AddServiceInvocationHandler("save", saveHandler); err != nil {
		log.Fatalf("error adding saveHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("del", delHandler); err != nil {
		log.Fatalf("error adding delHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("comment", commentHandler); err != nil {
		log.Fatalf("error adding commentHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("upvote", upvoteHandler); err != nil {
		log.Fatalf("error adding upvoteHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("image", imageHandler); err != nil {
		log.Fatalf("error adding imageHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("timeline", timelineHandler); err != nil {
		log.Fatalf("error adding timelineHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// saveHandler saves a new post into store and triggers analysis
func saveHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	saveReqCtr.Inc()
	var req frontend.SaveReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("saveHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("saveHandler dapr err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	epoch := time.Now()
	servLat := epoch.UnixMilli() - req.SendUnixMilli
	// generate a post id
	postId := util.PostId(req.UserId, req.SendUnixMilli)
	// decode & save images if any
	images := make([]string, 0)
	if len(req.Images) > 0 {
		items := make([]*dapr.SetStateItem, 0)
		// decode
		for i, d := range(req.Images) {
			imageId := util.ImageId(postId, i)
			dbytes, err := base64.StdEncoding.DecodeString(d)
			if err != nil {
				logger.Printf("saveHandler err decoding base64 image %s: %s",
					imageId, err.Error())
				return nil, err
			} else {
				images = append(images, imageId)
				newItem := &dapr.SetStateItem{
					Etag: nil,
					Key: imageId,
					// Metadata: map[string]string{
					// 	"created-on": time.Now().UTC().String(),
					// },
					Metadata: nil,
					Value: dbytes,
					Options: &dapr.StateOptions{
						// Concurrency: dapr.StateConcurrencyLastWrite,
						Consistency: dapr.StateConsistencyStrong,
					},
				}
				items = append(items, newItem)
			}
		}
		// update latency metric
		servLat += time.Now().UnixMilli() - epoch.UnixMilli()
		epoch = time.Now()
		// save
		err := client.SaveBulkState(ctx, imageStore, items...)
		if err != nil {
			logger.Printf("saveHandler err saving to %s: %s",
				imageStore, err.Error())
			return nil, err
		}
		// update latency metric
		storeLat := time.Now().UnixMilli() - epoch.UnixMilli()
		updateStoreLatHist.Observe(float64(storeLat))
		epoch = time.Now()
	}

	// save the PostCont
	saveReq := post.SavePostReq{
		PostId: postId,
		PostCont: post.PostCont{
			UserId: req.UserId,
			Text: req.Text,
			Images: images,
		},
		// sender side timestamp in unix millisecond
		SendUnixMilli: time.Now().UnixMilli(),
	}
	saveReqData, err := json.Marshal(saveReq)
	if err != nil {
		logger.Printf("saveHanlder json.Marshal (saveReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        saveReqData,
	}
	// update latency metric
	servLat += time.Now().UnixMilli() - epoch.UnixMilli()
	// invoke dapr-post
	saveRespData, err := client.InvokeMethodWithContent(ctx, "dapr-post", "save", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-post:save err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	var saveResp post.UpdatePostResp
	if err := json.Unmarshal(saveRespData, &saveResp); err != nil {
		logger.Printf("saveHandler json.Unmarshal (saveRespData) err: %s", err.Error())
		return nil, err
	} else {
		servLat += time.Now().UnixMilli() - saveResp.SendUnixMilli
		epoch = time.Now()
	}
	// issue request for timeline update
	timelineReq := timeline.UpdateReq{
		UserId: req.UserId,
		PostId: postId,
		Add: true,
		ImageIncluded: len(images) > 0,
		ClientUnixMilli: req.SendUnixMilli, 
		SendUnixMilli: time.Now().UnixMilli(),
	}
	tlUpdateReqCtr.Inc()
	if err := client.PublishEventfromCustomContent(ctx, timelinePubsub, timelineTopic, timelineReq); err != nil {
		logger.Printf("saveHandler err publish to pubsub %s topic %s: %s", 
			timelinePubsub, timelineTopic, err.Error())
		return nil, err
	}
	// issue request for object detection
	if len(images) > 0 {
		objReq := objectdetect.Req{
			PostId: postId,
			Images: images,
			ClientUnixMilli: req.SendUnixMilli, 
			SendUnixMilli: time.Now().UnixMilli(),
		}
		if err := client.PublishEventfromCustomContent(ctx, objectPubsub, objectTopic, objReq); err != nil {
			logger.Printf("saveHandler err publish to pubsub %s topic %s: %s", 
				objectPubsub, objectTopic, err.Error())
			return nil, err
		}
		objDetReqCtr.Inc()
	}
	// issue request for sentiment analysis
	sentiReq := sentiment.Req{
		PostId: postId,
		Text: req.Text,
		ImageIncluded: len(images) > 0,
		ClientUnixMilli: req.SendUnixMilli, 
		SendUnixMilli: time.Now().UnixMilli(),
	}
	if err := client.PublishEventfromCustomContent(ctx, sentiPubsub, sentiTopic, sentiReq); err != nil {
		logger.Printf("saveHandler err publish to pubsub %s topic %s: %s", 
			sentiPubsub, sentiTopic, err.Error())
		return nil, err
	}
	sentiReqCtr.Inc()
	// update latency metric
	endEpoch := time.Now()
	servLat += endEpoch.UnixMilli() - epoch.UnixMilli()
	if len(images) > 0 {
		saveReqImgLatHist.Observe(float64(servLat))
		// end-to-end latency metric
		e2ePostSaveImgLatHist.Observe(float64(endEpoch.UnixMilli() - req.SendUnixMilli))
	} else {
		saveReqLatHist.Observe(float64(servLat))
		// end-to-end latency metric
		e2ePostSaveLatHist.Observe(float64(endEpoch.UnixMilli() - req.SendUnixMilli))
	}
	
	// create response
	resp := frontend.UpdateResp{
        PostId:  postId,
    }
    respData, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("saveHandler json.Marshal (respData) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respData,
	}
	return
}

// delHandler deletes a post from db
func delHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	// defer reqCtr.Inc()
	var req frontend.DelReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("delHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("delHandler dapr err: %s", err.Error())
		return nil, err
	}
	delReq := post.DelPostReq{
		PostId: req.PostId,
		// sender side timestamp in unix millisecond
		SendUnixMilli: time.Now().UnixMilli(),
	}
	delReqData, err := json.Marshal(delReq)
	if err != nil {
		logger.Printf("delHanlder json.Marshal (saveReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        delReqData,
	}
	// update service latency hist
	servLat := time.Now().UnixMilli() - req.SendUnixMilli
	// invoke dapr-post
	delRespData, err := client.InvokeMethodWithContent(ctx, "dapr-post", "del", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-post:del err: %s", err.Error())
		return nil, err
	}
	// decode dapr-post response
	var delResp post.UpdatePostResp
	if err := json.Unmarshal(delRespData, &delResp); err != nil {
		logger.Printf("delHandler json.Unmarshal (delRespData) err: %s", err.Error())
		return nil, err
	} 
	// issue request for timeline update
	timelineReq := timeline.UpdateReq{
		UserId: req.UserId,
		PostId: req.PostId,
		Add: false,
		ImageIncluded: false,
		ClientUnixMilli: req.SendUnixMilli, 
		SendUnixMilli: time.Now().UnixMilli(),
	}
	if err := client.PublishEventfromCustomContent(ctx, timelinePubsub, timelineTopic, timelineReq); err != nil {
		logger.Printf("delHandler err publish to pubsub %s topic %s: %s", 
			timelinePubsub, timelineTopic, err.Error())
		return nil, err
	}
	// update latency metric
	epoch := time.Now()
	servLat += epoch.UnixMilli() - delResp.SendUnixMilli
	reqLatHist.Observe(float64(servLat))
	// end-to-end latency metric
	e2ePostUpdateLatHist.Observe(float64(epoch.UnixMilli() - req.SendUnixMilli))
	// create response
	resp := frontend.UpdateResp{
        PostId:  req.PostId,
    }
    respData, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("delHandler json.Marshal (respData) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respData,
	}
	return
}

// commentHandler adds comment to a post
func commentHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateReqCtr.Inc()
	var req frontend.CommentReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("commentHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("commentHandler dapr err: %s", err.Error())
		return nil, err
	}
	commReq := post.CommentReq{
		PostId: req.PostId,
		Comm: post.Comment{
			CommentId: util.CommentId(req.UserId, time.Now().UnixMilli()),
			UserId: req.UserId,
			ReplyTo: req.ReplyTo,
			Text: req.Text,
		},
		SendUnixMilli: time.Now().UnixMilli(),
	}
	commReqData, err := json.Marshal(commReq)
	if err != nil {
		logger.Printf("commentHanlder json.Marshal (commReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        commReqData,
	}
	// update service latency hist
	servLat := time.Now().UnixMilli() - req.SendUnixMilli
	// invoke dapr-post
	commRespData, err := client.InvokeMethodWithContent(ctx, "dapr-post", "comment", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-post:comment err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	var commResp post.UpdatePostResp
	if err := json.Unmarshal(commRespData, &commResp); err != nil {
		logger.Printf("commentHandler json.Unmarshal (commRespData) err: %s", err.Error())
		return nil, err
	} else {
		servLat += time.Now().UnixMilli() - commResp.SendUnixMilli
	}
	reqLatHist.Observe(float64(servLat))
	// end-to-end latency metric
	e2ePostUpdateLatHist.Observe(float64(time.Now().UnixMilli() - req.SendUnixMilli))
	// create response
	resp := frontend.UpdateResp{
        PostId:  req.PostId,
    }
    respData, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("commentHandler json.Marshal (respData) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respData,
	}
	return
}

// upvoteHandler upvotes a post
func upvoteHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateReqCtr.Inc()
	var req post.UpvoteReq
	// format check
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("upvoteHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	clientUnixMilli := req.SendUnixMilli
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("upvoteHandler dapr err: %s", err.Error())
		return nil, err
	}
	// update send timestamp
	req.SendUnixMilli = time.Now().UnixMilli()
	upvoteReqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("upvoteHanlder json.Marshal (upvoteReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        upvoteReqData,
	}
	// update service latency hist
	servLat := time.Now().UnixMilli() - clientUnixMilli
	// invoke dapr-post
	upvoteRespData, err := client.InvokeMethodWithContent(ctx, "dapr-post", "upvote", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-post:upvote err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	var upvoteResp post.UpdatePostResp
	if err := json.Unmarshal(upvoteRespData, &upvoteResp); err != nil {
		logger.Printf("upvoteHandler json.Unmarshal (upvoteRespData) err: %s", err.Error())
		return nil, err
	} else {
		servLat += time.Now().UnixMilli() - upvoteResp.SendUnixMilli
	}
	reqLatHist.Observe(float64(servLat))
	// end-to-end latency metric
	e2ePostUpdateLatHist.Observe(float64(time.Now().UnixMilli() - clientUnixMilli))
	// create response
	resp := frontend.UpdateResp{
        PostId:  req.PostId,
    }
    respData, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("upvoteHandler json.Marshal (respData) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respData,
	}
	return
}

// imageHandler returns the give image
func imageHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	imageReqCtr.Inc()
	var req frontend.ImageReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("imageHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("imageHandler dapr err: %s", err.Error())
		return nil, err
	}
	// update service latency hist
	servLat := time.Now().UnixMilli() - req.SendUnixMilli
	epoch := time.Now()
	// query store to get etag and up-to-date val
	item, err := client.GetState(ctx, imageStore, req.Image)
	if err != nil {
		logger.Printf("imageHandler GetState (image: %s) err: %s", req.Image, err.Error())
		return nil, err
	}
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	data := []byte(base64.StdEncoding.EncodeToString(item.Value))
	endEpoch := time.Now()
	servLat += endEpoch.UnixMilli() - epoch.UnixMilli()
	imgReqLatHist.Observe(float64(servLat))
	// end-to-end latency metric
	e2eImgLatHist.Observe(float64(endEpoch.UnixMilli() - req.SendUnixMilli))
	out = &common.Content{
		ContentType: "application/octet-stream",
		Data:        data,
	}
	return
}

// timelineHandler reads all the post in a given timeline (except image data)
func timelineHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	tlReqCtr.Inc()
	var req timeline.ReadReq
	// format check
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("timelineHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	clientUnixMilli := req.SendUnixMilli
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("timelineHandler dapr err: %s", err.Error())
		return nil, err
	}
	// update send timestamp
	req.SendUnixMilli = time.Now().UnixMilli()
	tlReqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("upvoteHanlder json.Marshal (upvoteReq) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        tlReqData,
	}
	// update service latency hist
	servLat := time.Now().UnixMilli() - clientUnixMilli
	// invoke dapr-timeline-read to read timeline
	tlRespData, err := client.InvokeMethodWithContent(ctx, "dapr-timeline-read", "read", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-timeline-read:read err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	var tlResp timeline.ReadResp
	if err := json.Unmarshal(tlRespData, &tlResp); err != nil {
		logger.Printf("timelineHandler json.Unmarshal (tlRespData) err: %s", err.Error())
		return nil, err
	} else {
		servLat += time.Now().UnixMilli() - tlResp.SendUnixMilli
	}
	epoch := time.Now()
	// return directly if there is no post in timeline
	if len(tlResp.PostIds) == 0 {
		readTlReqLatHist.Observe(float64(servLat))
		// end-to-end latency metric
		e2eReadTlLatHist.Observe(float64(epoch.UnixMilli() - clientUnixMilli))
		emptyResp := post.ReadResp {
			Posts: make(map[string]post.Post),
		}
		emptyRespData, _ := json.Marshal(emptyResp)
		out = &common.Content{
			ContentType: "application/json",
			Data:        emptyRespData,
		}
		return
	}
	// invoke dapr-post to read actual posts
	postReq := post.ReadPostReq{
		PostIds: tlResp.PostIds, 
		SendUnixMilli: time.Now().UnixMilli(),
	}
	postReqData, err := json.Marshal(postReq)
	if err != nil {
		logger.Printf("timelineHanlder json.Marshal (postReq) err: %s", err.Error())
		return nil, err
	}
	content = &dapr.DataContent{
		ContentType: "application/json",
		Data:        postReqData,
	}
	// update service latency hist
	servLat += time.Now().UnixMilli() - epoch.UnixMilli()
	// invoke dapr-post
	postRespData, err := client.InvokeMethodWithContent(ctx, "dapr-post", "read", "get", content)
	if err != nil {
		logger.Printf("invoke dapr-post:read err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	var postResp post.ReadPostResp
	if err := json.Unmarshal(postRespData, &postResp); err != nil {
		logger.Printf("upvoteHandler json.Unmarshal (postRespData) err: %s", err.Error())
		return nil, err
	}
	// create response
	resp := post.ReadResp{
        Posts:  postResp.Posts,
    }
    respData, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("timelineHandler json.Marshal (respData) err: %s", err.Error())
		return nil, err
	}
	// update latency metric
	epoch = time.Now()
	servLat += epoch.UnixMilli() - postResp.SendUnixMilli
	readTlReqLatHist.Observe(float64(servLat))
	// end-to-end latency metric
	e2eReadTlLatHist.Observe(float64(epoch.UnixMilli() - clientUnixMilli))
	out = &common.Content{
		ContentType: "application/json",
		Data:        respData,
	}
	return
}


