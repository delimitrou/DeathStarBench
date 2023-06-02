/*
 Package translate implements the data structures
 used to communicate with sentiment service
*/
package sentiment

type Req struct {
	PostId string `json:"post_id"`
	Text string `json:"text"`
	// true if the post includes image (used in tracing to identify different paths)
	ImageIncluded bool `json:"image_included"`
	// timestamp that client sends this request
	ClientUnixMilli int64 `json:"client_unix_ms"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}