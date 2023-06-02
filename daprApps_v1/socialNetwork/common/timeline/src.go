/*
 Package timeline implements the data structures
 used to communicate with timeline-update & timeline-read service
*/
package timeline

// ReadReq & ReadResp supports reading timeline of one user at a time
type ReadReq struct {
	UserId string `json:"user_id"`
	// true if reading user's own timeline, false if reading user's home timeline
	UserTl bool `json:"user_tl"`
	// time of the earliest post that should be read
	EarlUnixMilli int64 `json:"earl_unix_milli"`
	// number of post to read from the timeline
	Posts int `json:"posts"`	
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type ReadResp struct {
	PostIds []string `json:"post_ids"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// UpdateReq supports add/del one post to/from multiple users
type UpdateReq struct {
	// UserId is the user who owns the post to be added/deleted,
	// rather than the user whose timeline needs updated
	// The actual users whose timeline needs updates should be read from socialgraph service
	UserId string `json:"user_id"`
	PostId string `json:"post_id"`
	// true of adding post to timeline, and false if deleting from timeline
	Add bool `json:"add"`
	// true if the post includes image (used in tracing to identify different paths)
	ImageIncluded bool `json:"image_included"`
	// timestamp that client sends this request
	ClientUnixMilli int64 `json:"client_unix_ms"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type UpdateResp struct {
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// UserTlKey generates the key for user timeline given a user
func UserTlKey(userId string) string {
	return userId + "-u"
}
// HomeTlKey generates the key for home timeline given a user
func HomeTlKey(userId string) string {
	return userId + "-h"
}