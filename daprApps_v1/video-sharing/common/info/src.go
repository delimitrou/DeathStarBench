/*
 Package info implements the data structures
 used to communicate with video-info service, video Meta & Stats structure
 and related helper functions
 */
package info
// import (
// 	"fmt"
// )

// data structures for video MetaData
type Meta struct {
	// user that uploads the video
	UserId string `json:"user_id"`
	// available resolution of the video
	Resolutions []string `json:"resolutions"`
	// duration of the video
	Duration float64 `json:"duration"`
	// text description of the video
	Description string `json:"description"`
	// date that this video is uploaded
	Date string `json:"date"`
}
type Rating struct {
	// number of ratings
	Num int64 `json:"num"`
	// mean score of all ratings
	Score float64 `json:"score"`
	// mean squared score of all ratings
	ScoreSq float64 `json:"score_sq"`
}
// data structre for video info
type Info struct {
	VideoMeta Meta  `json:"meta"`
	// number of views
	Views int64 `json:"views"`
	Rate Rating  `json:"rating"`
}

// data structures to communicate with video-info
// upload a video
type UploadReq struct {
	VideoId string `json:"video_id"`
	// user that uploads the video
	UserId string `json:"user_id"`
	// available resolutions of the video
	Resolutions []string `json:"resolutions"`
	// duration of the video
	Duration float64 `json:"duration"`
	// date uploaded
	Date string `json:"date"`
	// text description of the video
	Description string `json:"description"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// rate a video
type RateReq struct {
	VideoId string `json:"video_id"`
	// if this request is changing an old rate
	Change bool `json:"change"`
	// rating score
	Score float64 `json:"score"`
	// original score for rate changing request
	OriScore float64 `json:"ori_score"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// view a video
type ViewReq struct {
	VideoId string `json:"video_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// genenral response
type Resp struct {
	VideoId string `json:"video_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}

// read info (both meta and stats) of a video
type InfoReq struct {
	VideoIds []string `json:"video_ids"`
	Upstream string `json:"upstream"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// resp of info request
type InfoResp struct {
	VideoInfo map[string]Info `json:"info"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
