/*
 Package translate implements the data structures
 used to communicate with translate service
*/
package translate

type TranslReq struct {
	Text string `json:"text"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}

type TranslResp struct {
	Transl string `json:"translation"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}