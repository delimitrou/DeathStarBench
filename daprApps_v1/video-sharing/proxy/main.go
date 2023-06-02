package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	// dapr
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// dapr-video-sharing
	"dapr-apps/video-sharing/common/trending"
	"dapr-apps/video-sharing/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
)

func setupConn() (*grpc.ClientConn) {
	// Testing 40 MB data exchange
	maxRequestBodySize := 32
	var opts []grpc.CallOption
	opts = append(opts, grpc.MaxCallRecvMsgSize((maxRequestBodySize)*1024*1024))
	conn, err := grpc.Dial(net.JoinHostPort("127.0.0.1",
		GetEnvValue("DAPR_GRPC_PORT", "50001")),
		grpc.WithDefaultCallOptions(opts...), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

var customConn = setupConn()

func main() {
	s, err := daprd.NewService(serviceAddress)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
	// communicate with frontend
	if err := s.AddServiceInvocationHandler("upload", uploadHandler); err != nil {
		log.Fatalf("error adding uploadHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("info", infoHandler); err != nil {
		log.Fatalf("error adding infoHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("video", videoHandler); err != nil {
		log.Fatalf("error adding videoHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("rate", rateHandler); err != nil {
		log.Fatalf("error adding rateHandler: %v", err)
	}
	if err := s.AddServiceInvocationHandler("getRate", getRateHandler); err != nil {
		log.Fatalf("error adding getRateHandler: %v", err)
	}
	// communicate with recmd
	if err := s.AddServiceInvocationHandler("trending", trendngHandler); err != nil {
		log.Fatalf("error adding trendngHandler: %v", err)
	}
	

	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func GetEnvValue(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}


func dummyResp()([]byte) {
	type DummyResp struct {
		send_unix_ms int64
	}
	resp := DummyResp{
		time.Now().UnixMilli(),
	}
	d, _ := json.Marshal(resp)
	return d
}

type UploadReq struct {
	User string `json:"user"`
	Video string `json:"video_b64"`
	Desc string `json:"description"`
	Date string `json:"date,omitempty"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type InfoReq struct {
	Videos []string `json:"videos"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type VideoReq struct {
	Video string `json:"video"`
	Res string `json:"resolution"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type RateReq struct {
	User string `json:"user"`
	Video string `json:"video"`
	Comment string `json:"comment"`
	Score float64 `json:"score"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type GetRateReq struct {
	Video string `json:"Video"`
	User string `json:"user"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

func trendngHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req trending.Req
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("trendngHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("trendngHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("saveHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "save", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:save err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-trending", "get", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:save err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// uploadHandler uploads a new video to db
func uploadHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req UploadReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("uploadHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("uploadHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("uploadHandler json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "upload", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:upload err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "upload", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:upload err: %s", err.Error())
		return nil, err
	}
	// logger.Println(string(respdata))
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// infoHandler gets metadata of a video
func infoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req InfoReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("infoHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("infoHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("commentHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "info", "get", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:info err: %s", err.Error())
	// 		return _, err
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "info", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:info err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// videoHandler fetches video
func videoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req VideoReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("videoHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client

	// Instantiate DAPR client with custom-grpc-client gRPC connection
	client := dapr.NewClientWithConnection(customConn)
	// defer client.Close()
	// client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("videoHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("upvoteHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "video", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:get_video err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "video", "get", content)
	// _, err = client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "video", "get", content)		
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:get_video err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
		// Data:        dummyResp(),
	}
	return
}

func rateHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req RateReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("rateHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("rateHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("imageHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "rate_video", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:rate_video err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "rate", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:rate err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

func getRateHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req GetRateReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("getRateHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client := dapr.NewClientWithConnection(customConn)
	if err != nil {
		logger.Printf("getRateHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("getRateHandler json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "get_rate", "get", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-video-frontend:timeline err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }
	
	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-video-frontend", "get_rate", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-video-frontend:get_rate err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}