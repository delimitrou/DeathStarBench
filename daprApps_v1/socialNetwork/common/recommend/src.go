/*
 Package recommend implements the data structures
 used to communicate with recommend service
 */
package recommend

type Req struct {
	UserId string `json:"user_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type Resp struct {
	UserIds []string `json:"user_ids"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}