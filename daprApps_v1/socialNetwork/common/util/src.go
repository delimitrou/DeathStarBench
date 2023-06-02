package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// pubsub rediliver interval in ms
func RedeliverInterval() int64 {
	return int64(60000)
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

// FollowKey returns the store key for the following list of given user
func FollowKey(userId string) string {
	return userId + "-flw"
}

// FollowerKey returns the store key for the follower list of given user
func FollowerKey(userId string) string {
	return userId + "-flwer"
}

// PostId generates a post id given user id and timestamp 
func PostId(userId string, unixMilli int64) string {
	return fmt.Sprintf("%s*%d", userId, unixMilli)
}
// PostIdCheck returns true if a PostId is in correct format
// returns false and the err otherwise
func PostIdCheck(postId string) (bool, error) {
	s := strings.Split(postId, "*")
	_, err := strconv.ParseInt(s[len(s)-1], 10, 64)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
// PostTime gets the unix milli timestamp from a post id
func PostIdTime(postId string) int64 {
	s := strings.Split(postId, "*")
	t, err := strconv.ParseInt(s[len(s)-1], 10, 64)
	if err != nil {
		return int64(0)
	} else {
		return t
	}
}
// ImageId generates a image id given the post id that owns the image, 
// and the index of the image within the post
func ImageId(postId string, id int) string {
	return fmt.Sprintf("%s-img-%d", postId, id)
}
// CommentId generates a comment id given the user id that owns the image, 
// and the timestamp that the comment is received
func CommentId(userId string, unixMilli int64) string {
	return fmt.Sprintf("comment*%s*%d", userId, unixMilli)
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
	epoch := time.Now()
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
		servLat += float64(time.Now().UnixMilli() - epoch.UnixMilli())
		epoch = time.Now()
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
		epoch = time.Now()
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
	// 1-100 & 105-500
	buckets := append(
		prometheus.LinearBuckets(1.0, 1.0, 100),
		prometheus.LinearBuckets(105.0, 5.0, 80)...
	)
	// 510-1000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(510.0, 10.0, 50)...
	)
	// 1025 - 5000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(1025.0, 25.0, 160)...
	)
	// 5050 - 10000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(5050.0, 50.0, 100)...
	)
	// 10000 - 60000
	buckets = append(
		buckets,
		prometheus.LinearBuckets(11000.0, 500.0, 100)...
	)
	return buckets
}