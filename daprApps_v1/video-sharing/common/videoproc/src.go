/*
 Package trending implements the data structures
 used to communicate with video-scale & video-thumbnail service
*/
package videoproc

// scale the target video to certain resolution
type ScaleReq struct {
	// unique id of the video
	VideoId string `json:"video_id"`
	// key of the actual video data in video-store
	DataId string `json:"data_id"`
	Width int `json:"width"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// client side timestamp in unix millisecond
	ClientUnixMilli int64 `json:"client_unix_ms"`
}
// generate thumnbail image of a certain video
type ThumbnailReq struct {
	// unique id of the video
	VideoId string `json:"video_id"`
	// key of the actual video data in video-store
	DataId string `json:"data_id"`
	Duration float64 `json:"duration"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// client side timestamp in unix millisecond
	ClientUnixMilli int64 `json:"client_unix_ms"`
}
