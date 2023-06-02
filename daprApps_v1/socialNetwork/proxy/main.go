package main

import (
	"context"
	"log"
	"os"
	"encoding/json"
	"time"

	// dapr
	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/client"
	daprd "github.com/dapr/go-sdk/service/grpc"

	// socialnet
	"dapr-apps/socialnet/common/frontend"
	"dapr-apps/socialnet/common/post"
	"dapr-apps/socialnet/common/timeline"	
	"dapr-apps/socialnet/common/recommend"	
	"dapr-apps/socialnet/common/util"
)

var (
	logger         = log.New(os.Stdout, "", 0)
	serviceAddress = util.GetEnvVar("ADDRESS", ":5005")
)

func main() {
	s, err := daprd.NewService(serviceAddress)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
	// communicate with frontend
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
	// communicate with recmd
	if err := s.AddServiceInvocationHandler("recmd", recmdHandler); err != nil {
		log.Fatalf("error adding recmdHandler: %v", err)
	}

	// start the server to handle incoming events
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
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

func saveHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req frontend.SaveReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("saveHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("saveHandler dapr err: %s", err.Error())
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
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "save", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:save err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "save", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:save err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// delHanlder deletes the post from db
// all info including contents, meta, comments and upvotes are deleted
func delHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req frontend.DelReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("delHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("delHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("delHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "del", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:del err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "del", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:del err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// commentHandler updates the comments of a post
// The new comment is appended to the list
func commentHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req frontend.CommentReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("commentHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("commentHandler dapr err: %s", err.Error())
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
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "comment", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:comment err: %s", err.Error())
	// 		return _, err
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "comment", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:comment err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// upvoteHandler performs the upvote operation
// It adds the upvoter to the upvote list in the store
func upvoteHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req post.UpvoteReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("upvoteHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("upvoteHandler dapr err: %s", err.Error())
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
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "upvote", "post", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:upvote err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "upvote", "post", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:upvote err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

func imageHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req frontend.ImageReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("imageHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("imageHandler dapr err: %s", err.Error())
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
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "image", "get", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:image err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "image", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:image err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

// timelineHandler reads all fields of required posts and combine them into complete Post data structure
func timelineHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req timeline.ReadReq
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("timelineHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("timelineHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("timelineHanlder json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke frontend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "timeline", "get", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-socialnet-frontend:timeline err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }

	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-socialnet-frontend", "timeline", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-socialnet-frontend:timeline err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}

func recmdHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var req recommend.Req
	if err := json.Unmarshal(in.Data, &req); err != nil {
		logger.Printf("recmdHandler json.Unmarshal err: %s", err.Error())
		return nil, err
	}
	req.SendUnixMilli = time.Now().UnixMilli()
	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf("recmdHandler dapr err: %s", err.Error())
		return nil, err
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		logger.Printf("recmdHandler json.Marshal (req) err: %s", err.Error())
		return nil, err
	}
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        reqData,
	}
	// invoke recommend
	// go func() {
	// 	_, err := client.InvokeMethodWithContent(ctx, "dapr-recommend", "recmd", "get", content)	
	// 	if err != nil {
	// 		logger.Printf("invoke dapr-recommend:timeline err: %s", err.Error())
	// 	}
	// }()
	// out = &common.Content{
	// 	ContentType: "application/json",
	// 	Data:        dummyResp(),
	// }
	
	respdata, err := client.InvokeMethodWithContent(ctx, "dapr-recommend", "recmd", "get", content)	
	if err != nil {
		logger.Printf("invoke dapr-recommend:timeline err: %s", err.Error())
		return nil, err
	}
	out = &common.Content{
		ContentType: "application/json",
		Data:        respdata,
	}
	return
}