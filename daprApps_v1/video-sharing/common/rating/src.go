/*
 Package rating implements the data structures
 used to communicate with user-rating service
 */
package rating

// Rating of one movie
type Rating struct {
	Comment string `json:"comment"`
	Score float64 `json:"score"`
}
// UserRating is all the ratings uploaded by a user
type UserRating struct {
	Ratings map[string]Rating `json:"ratings"`
}

// get the rating of a movie
type GetReq struct {
	UserId string `json:"user_id"`
	VideoId string `json:"video_id"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type GetResp struct {
	// if the user already rates the movie
	Exist bool `json:"exist"`
	Comment string `json:"comment"`
	Score float64 `json:"score"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// rate a movie or change current rating of a movie
type RateReq struct {
	UserId string `json:"user_id"`
	VideoId string `json:"video_id"`
	Comment string `json:"comment"`
	Score float64 `json:"score"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type RateResp struct {
	Exist bool `json:"exist"`
	OriScore float64 `json:"ori_score"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}