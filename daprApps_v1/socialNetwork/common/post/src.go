/*
 Package post implements the data structures
 used to communicate with post service, and also the Post data structure
 */
package post

// data structures for Post
// Post comment
type Comment struct {
	CommentId string `json:"comment_id"`
	UserId string `json:"user_id"`
	// ReplyTo is the comment id this comment is replying to
	// set to "" if commenting on the original post
	ReplyTo string `json:"reply_to"`
	Text string `json:"text"`
}
// Post contents, comments & stats are stored separateley in db
type PostCont struct {
	// user that owns the post (repost changes ownership)
	UserId string `json:"user_id"`
	// text contents
	Text string `json:"text"`
	// image ids
	Images []string `json:"images"`
}
type PostMeta struct {
	// sentiment analysis results of text
	Sentiment string `json:"sentiment"`
	// image objects, indexed by image id
	Objects map[string]string `json:"objects"`
}
type PostComments struct {
	// contents of the post
	Comments []Comment `json:"comments"`
}

// Post complete structure (actual structure sent to user) 
type Post struct {
	PostId string `json:"post_id"`
	Content PostCont `json:"content"`
	Meta PostMeta `json:"meta"`
	Comments PostComments `json:"comments"`
	Upvotes []string `json:"upvotes"`
}

// data structures to communicate with PostService
// Update post store
type SavePostReq struct {
	PostId string `json:"post_id"`
	PostCont PostCont `json:"content"`
	// sender side timestamp in unix millisecond
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type DelPostReq struct {
	PostId string `json:"post_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// update metadata of a post (rewrite is safe)
type MetaReq struct {
	PostId string `json:"post_id"`
	Sentiment string `json:"sentiment"`
	Objects map[string]string `json:"objects"`
	SendUnixMilli int64 `json:"send_unix_ms"`
} 
// add comment to a post
type CommentReq struct {
	PostId string `json:"post_id"`
	Comm Comment `json:"comm"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// upvote a post
type UpvoteReq struct {
	PostId string `json:"post_id"`
	UserId string `json:"user_id"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// response of requests that update post store
type UpdatePostResp struct {
	SendUnixMilli int64 `json:"send_unix_ms"`
}

// Read post store
type ReadPostReq struct {
	PostIds []string `json:"post_ids"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
type ReadPostResp struct {
	Posts map[string]Post `json:"posts"`
	SendUnixMilli int64 `json:"send_unix_ms"`
}
// response to end users
type ReadResp struct {
	Posts map[string]Post `json:"posts"`
}