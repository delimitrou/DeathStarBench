/*
 Package trending implements the data structures
 used to communicate with trending service
 */
package trending

// get the trending of videos between start and end date
type Req struct {
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// Resp is directly user-facing, so timestamp is omitted
type Resp struct {
	// sorted list of videos according to trending
	Videos []string `json:"videos"`
	// // sender side timestamp in unix millisecond
	// SendUnixMilli int64 `json:"send_unix_ms"`
}