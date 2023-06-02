package util

import (
	"context"
	"encoding/json"
	"errors"
	// "fmt"
	"log"
	"os"
	"strings"
	"strconv"
	"time"

	// prometheus
	"github.com/prometheus/client_golang/prometheus"

	// dapr
	dapr "github.com/dapr/go-sdk/client"
	// daprd "github.com/dapr/go-sdk/service/grpc"
)

var dateLayout string = "2006-01-02"
func TimeToDate(t time.Time) string {
	return t.Format(dateLayout)
}
func DateToTime(d string) (time.Time, error) {
	return time.Parse(dateLayout, d)
}
// DatesBetween return all the dates between the given start & end date
// an error is generated is startDate or endDate does not conform to dateLayout
func DatesBetween(startDate string, endDate string) ([]string, error) {
	startT, err := DateToTime(startDate)
	if err != nil {
		return nil, err
	}
	endT, err := DateToTime(endDate)
	if err != nil {
		return nil, err
	}
	dates := make([]string, 0)
	for d := startT; !d.After(endT); d = d.AddDate(0, 0, 1) {
		dates = append(dates, TimeToDate(d))
	}
	return dates, nil
}
// DaysBetween return the number of days between two given dates
func DaysBetween(start time.Time, end time.Time) int {
	if start.After(end) {
        start, end = end, start
    }
    return int(end.Sub(start).Hours() / 24)
}

func GetEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}

func GetEnvVarInt(key string, fallbackValue int) int {
	if val, ok := os.LookupEnv(key); ok {
		s := strings.TrimSpace(val)
		v, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		} else {
			return v
		}
	}
	return fallbackValue
}

// IsValInSlice checks if a val is already in a slice (type is string)
func IsValInSlice(val string, list []string)(in bool, pos int) {
	in = false
	for i, v := range list {
		if val == v {
			pos = i
			in = true
			break
		}
	}
	return
}

// UpdateStoreSlice is a helper function for updating a slice to store
// the operation can be either add or remove (denoted by isAdd)
// It also trunncates the slice to maxLen (maxLen=0 indicates no constraint)
// It retries upon etag mismatch error and returns the error otherwise
func UpdateStoreSlice(ctx context.Context, storeName string, key string, val string, isAdd bool, maxLen int, logger *log.Logger) (
		succ bool, servLat float64, storeLat float64, err error) {
	servLat = 0.0
	storeLat = 0.0
	succ = false
	loop := 0
	// set up dapr client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("UpdateStoreSlice dapr client err: %s", err.Error())
		err = err
		return
	}
	// try until success or getting error other than etag mismatch
	for ; !succ; {
		loop += 1
		// quit if executed too many times
		if loop >= 100 {
			err = errors.New("UpdateStoreSlice loop exceeds 100 rounds, quitted")
			return
		}
		epoch := time.Now()
		// query store to get etag and up-to-date val
		item, errl := client.GetState(ctx, storeName, key)
		if errl != nil {
			logger.Printf("UpdateStoreSlice GetState err: %s", errl.Error())
			err = errl
			return
		}
		// update latency metric
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// get stale value
		etag := item.Etag
		var staleSlice []string
		if string(item.Value) != "" {
			if errl := json.Unmarshal(item.Value, &staleSlice); err != nil {
				logger.Printf("UpdateStoreSlice json unmarshal Value (key: %s), err: %s", 
					key, errl.Error())
				err = errl
				return
			}
		} else {
			staleSlice = make([]string, 0)
		}
		// compose new value
		var newSlice []string
		if isAdd {
			// add new element to the slice
			// first check repeats
			if repeat, _ := IsValInSlice(val, staleSlice); repeat {
				logger.Printf("UpdateStoreSlice (add) find repetitive val:%s in slice of key:%s", 
					val, key)
				// no update needed
				succ = true
				return 
			} else {
				newSlice = append(staleSlice, val)
			}
		} else {
			// remove element from the slice
			// first check if the value exists
			if exist, pos := IsValInSlice(val, staleSlice); !exist {
				logger.Printf("UpdateStoreSlice (del) find val:%s not existing in key:%s", 
					val, key)
				// no update needed
				succ = true
				return 
			} else {
				newSlice = append(staleSlice[:pos], staleSlice[pos+1:]...)
			}
		}
		// truncate the slice if it exceeds max length
		// maxLen=0 means no requiment on length
		if maxLen > 0 && len(newSlice) > maxLen {
			newSlice = newSlice[len(newSlice) - maxLen:]
		}
		// try update store with etag
		newVal, errl := json.Marshal(newSlice)
		if errl != nil {
			logger.Printf("UpdateStoreSlice json.Marshal (newVal) err:%s", errl.Error())
			err = errl
			return
		}
		// update latency metric
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
		// try perform store
		newItem := &dapr.SetStateItem{
			Etag: &dapr.ETag{
				Value: etag,
			},
			Key: key,
			// Metadata: map[string]string{
			// 	"created-on": time.Now().UTC().String(),
			// },
			Metadata: nil,
			Value: newVal,
			Options: &dapr.StateOptions{
				// Concurrency: dapr.StateConcurrencyLastWrite,
				Concurrency: dapr.StateConcurrencyFirstWrite,
				Consistency: dapr.StateConsistencyStrong,
			},
		}
		errl = client.SaveBulkState(ctx, storeName, newItem)
		storeLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		if errl == nil {
			succ = true
		} else if strings.Contains(errl.Error(), "etag mismatch") {
			// etag mismatch, keeping on trying
			succ = false
		} else {
			// other errors, return
			logger.Printf("UpdateStoreSlice SaveBulkState (newItem) err:%s", errl.Error())
			err = errl
			return
		}
	}
	return
}

// LatBuckets generate a common latency histogram buckets for prometheus histogram
func LatBuckets() []float64 {
	// 1-200 & 200 - 500
	buckets := append(
		prometheus.LinearBuckets(1.0, 1.0, 200),
		prometheus.LinearBuckets(202.0, 2.0, 150)...
	)
	// 505 - 1000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(505.0, 5.0, 100)...
	)
	// 1010 - 2000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(1010.0, 10.0, 100)...
	)
	// 2050 - 10000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(2050.0, 50.0, 160)...
	)
	// 10000 - 60000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(11000.0, 1000.0, 50)...
	)
	return buckets
}

// LatBucketsFFprobe generates a ffprobe workload latency histogram buckets for prometheus histogram
func LatBucketsFFprobe() []float64 {
	// 5-1000 & 1025-2500
	buckets := append(
		prometheus.LinearBuckets(5.0, 5.0, 200),
		prometheus.LinearBuckets(1025.0, 25.0, 60)...
	)
	// 2600-5000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(2600.0, 100.0, 25)...
	)
	// 10000 - 60000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(10000.0, 5000.0, 11)...
	)
	return buckets
}

// LatBucketsFFmpeg generates a ffmpeg-thumbnail workload latency histogram buckets for prometheus histogram
func LatBucketsFFmpegThumb() []float64 {
	// 10-2000 & 2050-4500
	buckets := append(
		prometheus.LinearBuckets(10.0, 10.0, 200),
		prometheus.LinearBuckets(2050.0, 50.0, 50)...
	)
	// 5000 - 60000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(5000.0, 5000.0, 11)...
	)
	return buckets
}

// LatBucketsFFmpeg generates a ffmpeg-scale workload latency histogram buckets for prometheus histogram
func LatBucketsFFmpegScale() []float64 {
	// 100-20000 & 20500-50000
	buckets := append(
		prometheus.LinearBuckets(100.0, 100.0, 200),
		prometheus.LinearBuckets(20500.0, 500.0, 60)...
	)
	// 55000 - 120000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(55000.0, 5000.0, 14)...
	)
	return buckets
}