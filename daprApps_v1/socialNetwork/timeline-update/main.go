package main

import (
	"fmt"
	"context"
	"log"
	"os"
	"strings"
	"sort"
	"encoding/json"
	"time"
	"errors"
	"net/http"
	"math/rand"

	// dapr
	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/client"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// socialnet
	"dapr-apps/socialnet/common/timeline"
	"dapr-apps/socialnet/common/socialgraph"
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
	promAddress    = util.GetEnvVar("PROM_ADDRESS", ":8084") // address for prometheus service
	timelineStore  = util.GetEnvVar("TIMELINE_STORE", "timeline-store")
	pubSubName     = util.GetEnvVar("PUBSUB_NAME", "timeline-events")
	topicName      = util.GetEnvVar("TOPIC_NAME", "timeline")
)

// Subscription to tell the dapr what topic to subscribe.
// - PubsubName: is the name of the component configured in the metadata of pubsub.yaml.
// - Topic: is the name of the topic to subscribe.
// - Route: tell dapr where to request the API to publish the message to the subscriber when get a message from topic.
// - Match: (Optional) The CEL expression to match on the CloudEvent to select this route.
// - Priority: (Optional) The priority order of the route when Match is specificed.
//             If not specified, the matches are evaluated in the order in which they are added.
var updateSubsc = &common.Subscription{
	PubsubName: pubSubName,
	Topic:      topicName,
}

var (
	// max outstanding requests
	maxOutstand = util.GetEnvVarInt("MAX_OUTSTAND", 1000)
	// max attempts of each update operation
	maxTry = util.GetEnvVarInt("MAX_TRY", 100)
	// the worker pool size for update requests
	maxWorker = util.GetEnvVarInt("WORKER", 100)
	// max length of timeline, 0 means no limit
	maxTimelineLen = util.GetEnvVarInt("MAX_TIMELINE_LEN", 0)
)

// prometheus metric
var (
	reqCtr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tl_update_req",
			Help: "Number of timeline update requests received",
		},
	)
	reqLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tl_update_req_lat_hist",
		Help:    "Latency (ms) histogram of timeline update requests, excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	reqImgLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tl_update_img_req_lat_hist",
		Help:    "Latency (ms) histogram of timeline update requests (with image), excluding time waiting for kvs/db",
		Buckets: util.LatBuckets(), 
	})
	updateStoreLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "tl_store_update_lat_hist",
		Help:    "Latency (ms) histogram of updating (read then write) timeline store.",
		Buckets: util.LatBuckets(), 
	})
	e2eTlUpdateLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_tl_update_lat_hist",
		Help:    "End-to-end Latency (ms) histogram of timeline update.",
		Buckets: util.LatBuckets(), 
	})
	e2eTlUpdateImgLatHist = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "e2e_tl_update_img_lat_hist",
		Help:    "End-to-end Latency (ms) histogram of timeline update (with image).",
		Buckets: util.LatBuckets(), 
	})
)

// random generator
var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))

func setup_prometheus() {
	prometheus.MustRegister(reqCtr)
	prometheus.MustRegister(reqLatHist)
	prometheus.MustRegister(reqImgLatHist)
	prometheus.MustRegister(updateStoreLatHist)
	prometheus.MustRegister(e2eTlUpdateLatHist)
	prometheus.MustRegister(e2eTlUpdateImgLatHist)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	logger.Fatal(http.ListenAndServe(promAddress, nil))
}

// input fed to worker
type Job struct {
	Ctx context.Context
	Key string	// db key on which update is performed
	PostId string
	Add bool	// if true, add PostId to timeline, delete otherwise 
	UserTl bool  // if user tl update then true
	Epoch time.Time	// epoch that the job is fed into channel
	Res chan Result	// channel to which result is sent
}
// output of worker
type Result struct {
	Key string
	PostId string
	SumLat int64
	StoreLat int64
	ServLat int64
	Epoch time.Time	 // epoch that worker finishes processing the job
	Succ bool
	Err error
}
// global job queue
var queue chan Job = make(chan Job, maxOutstand)
// start all the workers
func startWorkers() {
	for i := 0; i < maxWorker; i+=1 {
		go worker(i, queue)
	}
}

// workers that performs update on each user's timeline
func worker(id int, jobs <-chan Job) {
	for j := range(jobs) {		
		res := Result {
			Key: j.Key,
			PostId: j.PostId,
			SumLat: 0,
			StoreLat: 0,
			ServLat: 0,
			Epoch: time.Now(),
			Succ: false,
			Err: nil,
		}
		ctx := j.Ctx
		// count queueing time as server latency
		res.ServLat += res.Epoch.UnixMilli() - j.Epoch.UnixMilli()
		epoch := res.Epoch
		// create the client
		client, err := dapr.NewClient()
		if err != nil {
			res.Epoch = time.Now()
			res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
			res.Err = err
			// signal completion and move to next job
			j.Res <- res
			continue
		}
		succ := false
		loop := 0
		for ; !succ; {
			// quit if executed too many times
			loop += 1
			if loop >= maxTry {
				res.Err = errors.New(fmt.Sprintf(
					"worker loop exceeds %d rounds for timeline:%s",
					maxTry, j.Key))
				break
			}
			// query store to get etag and up-to-date val
			item, err := client.GetState(ctx, timelineStore, j.Key)
			// update latency metric
			res.Epoch = time.Now()
			res.StoreLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
			epoch = res.Epoch
			// quit on err
			if err != nil {
				res.Err = err
				break
			}
			// get stale value
			etag := item.Etag
			var staleTl []string
			if string(item.Value) != "" {
				// latency metric updated later if err is nil
				if err := json.Unmarshal(item.Value, &staleTl); err != nil {
					res.Err = err
					res.Epoch = time.Now()
					res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
					break
				}
			} else {
				staleTl = make([]string, 0)
			}
			// // todo: for debug, remove later
			// logger.Printf("tl length = %d", len(staleTl))
			// // ---
			// compose new value
			var newTl []string
			if j.Add {
				// add new element to the slice
				// first check repeats
				if repeat, _ := util.IsValInSlice(j.PostId, staleTl); repeat {
					logger.Printf("worker (add) finds repetitive post:%s for timeline:%s", 
						j.PostId, j.Key)
					// no update needed
					succ = true
					// update latency metrics
					res.Epoch = time.Now()
					res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
					break
				} else {
					newTl = append(staleTl, j.PostId)
					if randGen.Intn(1000) <= 2 {
						// sort new timeline for home timeline
						sort.Slice(newTl, func(i, j int) bool {
							return util.PostIdTime(newTl[i]) < util.PostIdTime(newTl[j])
						})
					}
					if maxTimelineLen > 0 && len(newTl) > maxTimelineLen {
						start := len(newTl) - maxTimelineLen
						newTl = newTl[start:]
					}
				}
			} else {
				// remove element from the slice
				// first check if the value exists
				if exist, pos := util.IsValInSlice(j.PostId, staleTl); !exist {
					logger.Printf("worker (del) find post:%s non-existing for timeline:%s", 
						j.PostId, j.Key)
					// no update needed
					succ = true
					// update latency metrics
					res.Epoch = time.Now()
					res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
					break
				} else {
					newTl = append(staleTl[:pos], staleTl[pos+1:]...)
				}
			}

			// try update store with etag
			data, err := json.Marshal(newTl)
			if err != nil {
				res.Err = err
				// update latency metrics
				res.Epoch = time.Now()
				res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
				break
			}
			// update latency metrics
			res.Epoch = time.Now()
			res.ServLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
			epoch = res.Epoch
			// try perform store
			newItem := &dapr.SetStateItem{
				Etag: &dapr.ETag{
					Value: etag,
				},
				Key: j.Key,
				// Metadata: map[string]string{
				// 	"created-on": time.Now().UTC().String(),
				// },
				Metadata: nil,
				Value: data,
				Options: &dapr.StateOptions{
					// Concurrency: dapr.StateConcurrencyLastWrite,
					Concurrency: dapr.StateConcurrencyFirstWrite,
					Consistency: dapr.StateConsistencyStrong,
				},
			}
			err = client.SaveBulkState(ctx, timelineStore, newItem)
			res.Epoch = time.Now()
			res.StoreLat += res.Epoch.UnixMilli() - epoch.UnixMilli()
			epoch = res.Epoch
			if err == nil {
				succ = true
			} else if strings.Contains(err.Error(), "etag mismatch") {
				// etag mismatch, keeping on trying
				succ = false
			} else {
				// other errors, quit
				res.Err = err
				break
			}
		}
		// // todo: for debug, remove later
		// logger.Printf("update loop = %d", loop)
		// // ---
		// signal completion
		res.Succ = succ
		j.Res <- res
	}
}

func updateHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	reqCtr.Inc()
	// decode data as json
	data_bytes, ok := e.Data.([]byte)
	if !ok {
		logger.Printf("updateHanlder err: event.Data can not be converted to []byte: %s", 
			e.Data)
		return false, errors.New("event.Data can not be converted to []byte")
	}
	var req timeline.UpdateReq
	if err := json.Unmarshal(data_bytes, &req); err != nil {
		logger.Printf("updateHanlder json.Unmarshal (in.Data) err: %s", err.Error())
		return false, err
	}
	// check PostId is in correct format
	r, err := util.PostIdCheck(req.PostId)
	if !r {
		logger.Printf("updateHanlder PostId %s wrong format, err: %s", 
			req.PostId, err.Error())
		return false, errors.New(fmt.Sprintf("PostId:%s wrong format", req.PostId))
	}
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("updateHanlder dapr err: %s", err.Error())
		return false, err
	}
	// query socialgraph service to get user's follower graph
	userIds := make([]string, 1)
	userIds[0] = req.UserId
	graphReq := socialgraph.GetReq{
		SendUnixMilli: time.Now().UnixMilli(),
		UserIds: userIds,
	}
	graphReqData, err := json.Marshal(graphReq)
	if err != nil {
		logger.Printf("updateHanlder json.Marshal (graphReq) err: %s", err.Error())
		return false, err
	}
	// update service latency hist
	epoch := time.Now()
	servLat := epoch.UnixMilli() - req.SendUnixMilli
	redeliver := servLat >= util.RedeliverInterval()

	storeLat := int64(0)
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        graphReqData,
	}
	graphRespData, err := client.InvokeMethodWithContent(ctx, "dapr-social-graph", "getfollower", "get", content)	
	if err != nil {
		logger.Printf("dapr-social-graph:getfollower err: %s", err.Error())
		return false, err
	}
	// decode graphResp
	var graphResp socialgraph.GetFollowerResp
	if err := json.Unmarshal(graphRespData, &graphResp); err != nil {
		logger.Printf("updateHanlder json.Unmarshal (graphRespData) err: %s", err.Error())
		return false, err
	}
	flwers, ok := graphResp.FollowerIds[req.UserId]
	if !ok {
		logger.Printf("updateHanlder err: user %s not in graphResp", req.UserId)
		return false, errors.New(fmt.Sprintf("user %s not in graphResp", req.UserId))
	}
	// update service latency hist
	// Note: should not be updated after jobs are created, otherwise the compute+queue wait time is double counted
	servLat += time.Now().UnixMilli() - graphResp.SendUnixMilli
	// compose jobs
	jobs := make([]Job, len(flwers) + 1)
	results := make([]Result, len(flwers) + 1)
	// update user's user timeline
	jobs[0] = Job {
		Ctx: ctx,
		Key: timeline.UserTlKey(req.UserId),
		PostId: req.PostId,
		Add: req.Add,
		UserTl: true, // never truncate user's own timeline, so that the whole timline can be reconstructed
		Epoch: time.Now(),
		Res: make(chan Result),	
	}
	queue <- jobs[0]
	// update followers' home timeline
	for i, flwerId := range flwers {
		jobs[i + 1] = Job {
			Ctx: ctx,
			Key: timeline.HomeTlKey(flwerId),
			PostId: req.PostId,
			Add: req.Add,
			UserTl: false, 
			Epoch: time.Now(),
			Res: make(chan Result),	
		}
		queue <- jobs[i + 1]
	}
	// wait for results from workers
	for i, _ := range jobs {
		results[i] = <-jobs[i].Res
	}
	// todo: for debug, remove later ---
	workerLat := time.Now().UnixMilli() - epoch.UnixMilli()
	// ---
	// update latency metric
	epoch = time.Now()
	errStr := ""
	succ := true
	var workerServLat int64 = 0
	var workerStoreLat int64 = 0
	for _, r := range results {
		// include the time between worker finishes the job 
		// and handlder gets ALL resp serv latency, 
		// so that the average serv + store lat should equal
		// real time elapsed
		workerServLat += r.ServLat + epoch.UnixMilli() - r.Epoch.UnixMilli()
		workerStoreLat += r.StoreLat
		if !r.Succ {
			logger.Printf("worker err (timeline:%s, post:%s): %s", 
				r.Key, r.PostId, r.Err.Error())
			succ = false
			errStr += r.Err.Error() + "; "
		}
	}
	// compute latency metrics (by averaging)
	workerServLat = workerServLat / int64(len(jobs))
	workerStoreLat = workerStoreLat / int64(len(jobs))
	// update latency metrics
	servLat += workerServLat
	storeLat += workerStoreLat
	// todo: for debug, remove later ---
	logger.Printf("workerLat=%d, workerServLat=%d, workerStoreLat=%d, Lat-(Serv+Store)=%d, servLat=%d, e2eLat=%d",
		workerLat, workerServLat, workerStoreLat, 
		workerLat-workerServLat-workerStoreLat,
		servLat,
		time.Now().UnixMilli() - req.ClientUnixMilli,
	)
	// ---
	updateStoreLatHist.Observe(float64(storeLat))
	// update end-to-end latency metric
	if !redeliver {
		// ignore latency if this is a redeliver
		if req.ImageIncluded {
			reqImgLatHist.Observe(float64(servLat))
			e2eTlUpdateImgLatHist.Observe(float64(time.Now().UnixMilli() - req.ClientUnixMilli))
		} else {
			reqLatHist.Observe(float64(servLat))
			e2eTlUpdateLatHist.Observe(float64(time.Now().UnixMilli() - req.ClientUnixMilli))
		}
	}
	
	
	// return err if any update failed
	if succ {
		return false, nil
	} else {
		return false, errors.New(errStr)
	}
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
	// start worker pool
	startWorkers()
	if err := s.AddTopicEventHandler(updateSubsc, updateHandler); err != nil {
		log.Fatalf("error adding updateHandler: %v", err)
	}
	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
