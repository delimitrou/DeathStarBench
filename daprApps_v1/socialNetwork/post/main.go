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

	// socialnet
	"dapr-apps/socialnet/common/post"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	postStore      = util.GetEnvVar("POST_STORE", "post-store")
)
// the maximum attempts to save to post-store. Quit the operation if exceeded
var maxTry = 100

// prometheus metric
var (
	readCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "post_read_req",
			Help: "Number of post read requests received.",
		},
	)
	updateCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "post_update_req",
			Help: "Number of post update requests received.",
		},
	)
	saveReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_save_req_lat_hist",
		Help:    "Latency (ms) histogram of post save requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	updateReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_update_req_lat_hist",
		Help:    "Latency (ms) histogram of post update requests (meta, comment, upvote & del), excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	readReqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_read_req_lat_hist",
		Help:    "Latency (ms) histogram of post read requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	readStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_store_read_lat_hist",
		Help:    "Latency (ms) histogram of reading post store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	writeStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_store_write_lat_hist",
		Help:    "Latency (ms) histogram of updating (write) post store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "post_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) post store (kvs/db).",
		Buckets: util.LatBuckets(), 
	})
)

func setup_prometheus() {
	prometheus.MustRegister(readCtr)
	prometheus.MustRegister(updateCtr)
	// service latency
	prometheus.MustRegister(saveReqLatHist)
	prometheus.MustRegister(updateReqLatHist)
	prometheus.MustRegister(readReqLatHist)
	// state store latency
	prometheus.MustRegister(readStoreLatHist)
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
	if err := s.AddServiceInvocationHandler("save", saveHandler); err != nil {
		log.Fatalf("error adding saveHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("del", delHandler); err != nil {
		log.Fatalf("error adding delHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("meta", metaHandler); err != nil {
		log.Fatalf("error adding metaHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("comment", commentHandler); err != nil {
		log.Fatalf("error adding commentHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("upvote", upvoteHandler); err != nil {
		log.Fatalf("error adding upvoteHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("read", readHandler); err != nil {
		log.Fatalf("error adding readHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
// metaKey returns the content store key given a post id
func contKey(postId string) string {
	return postId + "-ct"
} 
// metaKey returns the metadata store key given a post id
func metaKey(postId string) string {
	return postId + "-me"
}
// commKey returns the comment store key given a post id
func commKey(postId string) string {
	return postId + "-cm"
}
// upvoteKey returns the comment store key given a post id
func upvoteKey(postId string) string {
	return postId + "-up"
}

// saveHanlder saves the post content into db
// metaData, comment and upvotes are not touched
func saveHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req post.SavePostReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("saveHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("saveHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	contkey := contKey(req.PostId)
	contjson, err := json.Marshal(req.PostCont)
	if err != nil {
		logger.Printf("saveHandlere json.Marshal (contjson) err:%s", err.Error())
		return nil, err
	}
	// update service latency hist
	epoch := time.Now()
	saveReqLatHist.Observe(float64(epoch.UnixMilli() - req.SendUnixMilli))
	// update post store
	if err := client.SaveState(ctx, postStore, contkey, contjson); err != nil {
		logger.Printf("saveHandler saveState (contjson) err: %s", err.Error())
		return nil, err
	}
	// update store latency hist
	writeStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	// create response
	resp := post.UpdatePostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// delHanlder deletes the post from db
// all info including contents, meta, comments and upvotes are deleted
func delHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req post.DelPostReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("delHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("delHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	contkey := contKey(req.PostId)
	metakey := metaKey(req.PostId)
	commkey := commKey(req.PostId)
	upkey := upvoteKey(req.PostId)
	keys := []string{contkey, metakey, commkey, upkey}
	// update service latency hist
	epoch := time.Now()
	updateReqLatHist.Observe(float64(epoch.UnixMilli() - req.SendUnixMilli))
	// del from post store
	if err := client.DeleteBulkState(ctx, postStore, keys); err != nil {
		logger.Printf("delHandler saveState (contjson) err: %s", err.Error())
		return nil, err
	}
	// update store latency hist
	updateStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	// create response
	resp := post.UpdatePostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// metaHandler updates the metadata of a post
// It can either be updating Sentiment or Objects
// The fields not updated should be set to default val in the inv request
func metaHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req post.MetaReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("metaHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// latency metrics
	epoch := time.Now()
	var servLat float64 = float64(epoch.UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("metaHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	metakey := metaKey(req.PostId)
	// loop to update store
	var succ = false
	var loop = 0
	for ; !succ; {
		loop += 1
		// quit if loop exceeds maxTry
		if loop > maxTry {
			err = errors.New(fmt.Sprintf("metaHandler update key: %s loop exceeds %d rounds, quitted",
				metakey, maxTry))
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, postStore, metakey)
		if errl != nil {
			logger.Printf("metaHandler GetState (key: %s) err: %s", metakey, errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var meta post.PostMeta
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &meta); err != nil {
				logger.Printf("metaHandler json.Unmarshal Value (key: %s), err: %s", 
					metakey, errl.Error())
				err = errl
				return nil, err
			}
		} else {
			meta = post.PostMeta{
				Sentiment: "",
				Objects: nil,
			}
		}
		// fill in new data
		if req.Sentiment != "" {
			meta.Sentiment = req.Sentiment
		}
		if req.Objects != nil {
			meta.Objects = req.Objects
		}
		// try update store with etag
		metajson, errl := json.Marshal(meta)
		if errl != nil {
			logger.Printf("metaHandler json.Marshal (metajson) err:%s", errl.Error())
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
			Key: metakey,
			// Metadata: map[string]string{
			// 	"created-on": time.Now().UTC().String(),
			// },
			Metadata: nil,
			Value: metajson,
			Options: &dapr.StateOptions{
				// Concurrency: dapr.StateConcurrencyLastWrite,
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		}
		errl = client.SaveBulkState(ctx, postStore, newItem)
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		if errl == nil {
			succ = true
		} else if strings.Contains(errl.Error(), "etag mismatch") {
			// etag mismatch, keeping on trying
			succ = false
		} else {
			// other errors, return
			logger.Printf("metaHandler SaveBulkState (metajson) key:%s, err:%s", 
				metakey, errl.Error())
			err = errl
			return nil, err
		}

	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := post.UpdatePostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// commentHandler updates the comments of a post
// The new comment is appended to the list
func commentHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req post.CommentReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("commentHandler json.Unmarshal (in.Data) err: %s", err.Error())
		return nil, err
	}
	// latency metrics
	epoch := time.Now()
	var servLat float64 = float64(epoch.UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("commentHandler dapr err: %s", err.Error())
		return nil, err
	}
	// generate data
	commkey := commKey(req.PostId)
	// loop to update store
	var succ = false
	var loop = 0
	for ; !succ; {
		loop += 1
		// quit if loop exceeds maxTry
		if loop > maxTry {
			err = errors.New(fmt.Sprintf("commentHandler update key:%s loop exceeds %d rounds, quitted",
				commkey, maxTry))
			return nil, err
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, postStore, commkey)
		if errl != nil {
			logger.Printf("commentHandler GetState (key: %s) err: %s", commkey, errl.Error())
			err = errl
			return nil, err
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var comm post.PostComments
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &comm); err != nil {
				logger.Printf("commentHandler json.Unmarshal Value (key: %s), err: %s", 
					commkey, errl.Error())
				err = errl
				return nil, err
			}
		} else {
			comm = post.PostComments{
				Comments: make([]post.Comment, 0),
			}
		}
		// fill in new data
		comm.Comments = append(comm.Comments, req.Comm)
		// try update store with etag
		commjson, errl := json.Marshal(comm)
		if errl != nil {
			logger.Printf("commentHandler json.Marshal (commjson) err:%s", errl.Error())
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
			Key: commkey,
			// Metadata: map[string]string{
			// 	"created-on": time.Now().UTC().String(),
			// },
			Metadata: nil,
			Value: commjson,
			Options: &dapr.StateOptions{
				// Concurrency: dapr.StateConcurrencyLastWrite,
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		}
		errl = client.SaveBulkState(ctx, postStore, newItem)
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		if errl == nil {
			succ = true
		} else if strings.Contains(errl.Error(), "etag mismatch") {
			// etag mismatch, keeping on trying
			succ = false
		} else {
			// other errors, return
			logger.Printf("commentHandler SaveBulkState (commjson) key:%s, err:%s", 
				commkey, errl.Error())
			err = errl
			return nil, err
		}

	}
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	// create response
	resp := post.UpdatePostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// upvoteHandler performs the upvote operation
// It adds the upvoter to the upvote list in the store
func upvoteHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	updateCtr.Inc()
	var req post.UpvoteReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("upvoteHanlder json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	// update service latency hist
	var servLat float64 = float64(time.Now().UnixMilli() - req.SendUnixMilli)
	var storeLat float64 = 0.0
	// update follow list of user
	upkey := upvoteKey(req.PostId)
	succ, servl, storel, err := util.UpdateStoreSlice(ctx, postStore, upkey, req.UserId, true, 0, logger)
	servLat += servl
	storeLat += storel
	// update store latency hist
	updateReqLatHist.Observe(servLat)
	updateStoreLatHist.Observe(storeLat)
	if !succ {
		return nil, err
	}
	// create response
	resp := post.UpdatePostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
    }
    respdata, _ := json.Marshal(resp)
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// readHandler reads all fields of required posts and combine them into complete Post data structure
func readHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	readCtr.Inc()
	var req post.ReadPostReq
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
	// query store to get all fields of required posts
	keys := make([]string, 0)
	for _, pid  := range req.PostIds {
		contkey := contKey(pid)
		metakey := metaKey(pid)
		commkey := commKey(pid)
		upkey := upvoteKey(pid)
		keys = append(keys, contkey, metakey, commkey, upkey)
	}
	// update service latency hist
	epoch := time.Now()
	servLat := float64(epoch.UnixMilli() - req.SendUnixMilli)
	items, err := client.GetBulkState(ctx, postStore, keys, nil, int32(len(keys)))
	// update store latency hist
	readStoreLatHist.Observe(float64(time.Now().UnixMilli() - epoch.UnixMilli()))
	epoch = time.Now()
	if err != nil {
		logger.Printf("readHandler getBulkState err: %s", err.Error())
		return nil, err
	}
	if len(items) != len(keys) {
		logger.Printf("readHandler getBulkState len(items) != len(keys)")
		return nil, errors.New(fmt.Sprintf("store %s: getBulkState len(items) != len(keys)", postStore))
	}
	// combine fields into complete posts
	posts := make(map[string]post.Post)
	itMap := make(map[string]*dapr.BulkStateItem)
	for _, it := range(items) {
		itMap[it.Key] = it
	}
	for _, pid := range req.PostIds {
		contkey := contKey(pid)
		metakey := metaKey(pid)
		commkey := commKey(pid)
		upkey := upvoteKey(pid)
		// empty post
		p := post.Post {
			PostId: pid,
			Content: post.PostCont{
				UserId: "",
				Text: "Missing post",
				Images: nil,
			},
			Meta: post.PostMeta{
				Sentiment: "",
				Objects: nil,
			},
			Comments: post.PostComments{
				Comments: nil,
			},
			Upvotes: nil,
		}
		// unmarshal content
		if string(itMap[contkey].Value) != "" {
			var cont post.PostCont
			if err = json.Unmarshal(itMap[contkey].Value, &cont); err != nil {
				logger.Printf("readHandler json.Unmarshal (content Value) for key: %s, err: %s", 
					contKey(pid), err.Error())
				return nil, err
			} else {
				p.Content = cont
			}
		} else {
			// missing contents, return an empty post and continue
			posts[pid] = p
			continue
		}
		// unmarshal meta
		if string(itMap[metakey].Value) != "" {
			var meta post.PostMeta
			if err = json.Unmarshal(itMap[metakey].Value, &meta); err != nil {
				logger.Printf("readHandler json.Unmarshal (meta Value) for key: %s, err: %s", 
					metaKey(pid), err.Error())
				return nil, err
			} else {
				p.Meta = meta
			}
		}
		// unmarshal comments
		if string(itMap[commkey].Value) != "" {
			var comm post.PostComments
			if err = json.Unmarshal(itMap[commkey].Value, &comm); err != nil {
				logger.Printf("readHandler json.Unmarshal (comment Value) for key: %s, err: %s", 
					commKey(pid), err.Error())
				return nil, err
			} else {
				p.Comments = comm
			}
		}
		// unmarshal upvotes
		if string(itMap[upkey].Value) != "" {
			var up []string
			if err = json.Unmarshal(itMap[upkey].Value, &up); err != nil {
				logger.Printf("readHandler json.Unmarshal (upvoes Value) for key: %s, err: %s", 
					upvoteKey(pid), err.Error())
				return nil, err
			} else {
				p.Upvotes = up
			}
		}
		// add this post to the map
		posts[pid] = p
	}
	// create response
	resp := post.ReadPostResp{
        SendUnixMilli:  time.Now().UnixMilli(),
        Posts: posts,
    }
    respdata, err := json.Marshal(resp)
	// update service latency hist
	servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
	readReqLatHist.Observe(servLat)
	if err != nil {
		logger.Printf("readHandler json.Marshal (resp) err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}