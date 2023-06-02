/*
 Package translate implements the data structures
 used to communicate with translate service
*/
package user

type RegisterReq struct {
	UserId string `json:"user_id"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type RegisterResp struct {
	// UserId string `json:"user_id"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}