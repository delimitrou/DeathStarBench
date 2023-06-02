/*
 Package objectdetect implements the data structures
 used to communicate with object-detect service
*/
package objectdetect

type Req struct {
	PostId string `json:"post_id"`
	// image ids
	Images []string `json:"images"`
	// timestamp that client sends this request
	ClientUnixMilli int64 `json:"client_unix_ms"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}