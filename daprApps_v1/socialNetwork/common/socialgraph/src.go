/*
 Package socialgraph implements the data structures
 used to communicate with socialgraph service
 */
package socialgraph

// reading follow & follower info
type GetReq struct {
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// users whose info needs fetched
	UserIds []string `json:"user_ids"`
}
type GetRecmdReq struct {
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// users whose info needs fetched
	UserIds []string `json:"user_ids"`
	// record latency if true
	Record bool `json:"record"`
	// latency from previous request
	Latency int64 `json:"latency"`
}
type GetFollowResp struct {
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// indexed by user id in the GetReq
	FollowIds map[string][]string `json:"follow_ids"`
}
type GetRecmdResp struct {
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// indexed by user id in the GetReq
	FollowIds map[string][]string `json:"follow_ids"`
	// latency of this request
	Latency int64 `json:"latency"`
}
type GetFollowerResp struct {
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
	// indexed by user id in the GetReq
	FollowerIds map[string][]string `json:"follower_ids"`
}
// update follow & follower info
type FollowReq struct {
	UserId string `json:"user_id"`
	FollowId string `json:"follow_id"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type UnfollowReq struct {
	UserId string `json:"user_id"`
	UnfollowId string `json:"unfollow_id"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type UpdateResp struct {
	SendUnixMilli int64 `json:"send_unix_ms"`
}