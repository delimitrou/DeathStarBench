/*
 Package frontend implements the data structure used to communicate with frontend service, 
 including creating posts and read images
 For request format for reading timelines, check timeline.ReadReq & post.ReadResp
 Registering new users is implemeted by direclty accessing user service
 Following other users is implmented by directly accessing socialgraph service
 Translation is implemented by directly accessing translate service
 Recommend is implemented by directly accessing recommend service
 */
package frontend

// save a post
type SaveReq struct {
	UserId string `json:"user_id"`
	// text contents
	Text string `json:"text"`
	// image data encoded as based64
	Images []string `json:"images"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// removes a post from both post store and timeline
type DelReq struct {
	UserId string `json:"user_id"`
	PostId string `json:"post_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// add comment to a post
type CommentReq struct {
	PostId string `json:"post_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
	// contents
	UserId string `json:"user_id"`
	// ReplyTo is the comment id this comment is replying to
	// set to "" if commenting on the original post
	ReplyTo string `json:"reply_to"`
	Text string `json:"text"`
}
// upvote uses post.UpvoteReq 
type UpdateResp struct {
	PostId string `json:"post_id"`
}

// read image (one at each time)
// response is the binary data of the image
type ImageReq struct {
	// image id
	Image string `json:"image"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}

