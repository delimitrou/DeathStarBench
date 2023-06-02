/*
 Package datestore implements the data structures
 used to communicate with in dates service
 */
package dates

// get the list of videos of a date
type GetReq struct {
	Dates []string `json:"dates"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type GetResp struct {
	Videos map[string][]string `json:"videos"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// add a new video to a date
type UploadReq struct {
	Date string `json:"date"`
	VideoId string `json:"video_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type UploadResp struct {
	SendUnixMilli int64 `json:"send_unix_ms"`
}