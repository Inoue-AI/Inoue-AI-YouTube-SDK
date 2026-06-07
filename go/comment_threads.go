package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CommentThread parts accepted by the part parameter.
const (
	CommentThreadPartID      = "id"
	CommentThreadPartSnippet = "snippet"
	CommentThreadPartReplies = "replies"
)

// Comment thread sort orders.
const (
	CommentThreadOrderTime      = "time"
	CommentThreadOrderRelevance = "relevance"
)

// CommentThreadSnippet is the snippet part of a CommentThread resource.
type CommentThreadSnippet struct {
	ChannelID       string   `json:"channelId,omitempty"`
	VideoID         string   `json:"videoId,omitempty"`
	TopLevelComment *Comment `json:"topLevelComment,omitempty"`
	CanReply        *bool    `json:"canReply,omitempty"`
	TotalReplyCount *int     `json:"totalReplyCount,omitempty"`
	IsPublic        *bool    `json:"isPublic,omitempty"`
}

// CommentThreadReplies is the replies part of a CommentThread resource. It may
// not include all replies; use ListComments with ParentID for the full set.
type CommentThreadReplies struct {
	Comments []Comment `json:"comments"`
}

// CommentThread is a YouTube CommentThread resource (request body and response).
type CommentThread struct {
	Kind    string                `json:"kind,omitempty"`
	Etag    string                `json:"etag,omitempty"`
	ID      string                `json:"id,omitempty"`
	Snippet *CommentThreadSnippet `json:"snippet,omitempty"`
	Replies *CommentThreadReplies `json:"replies,omitempty"`
}

// CommentThreadListResponse is the response from ListCommentThreads.
type CommentThreadListResponse struct {
	Kind          string          `json:"kind,omitempty"`
	Etag          string          `json:"etag,omitempty"`
	NextPageToken string          `json:"nextPageToken,omitempty"`
	PageInfo      *PageInfo       `json:"pageInfo,omitempty"`
	Items         []CommentThread `json:"items"`
}

// ListCommentThreadsParams configures ListCommentThreads. Exactly one of
// VideoID, ChannelID, IDs, or AllThreadsRelatedToChannelID should be supplied.
type ListCommentThreadsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// VideoID returns threads associated with a video.
	VideoID string
	// ChannelID returns threads associated with a channel.
	ChannelID string
	// IDs restricts the result to specific thread IDs.
	IDs []string
	// AllThreadsRelatedToChannelID returns all threads associated with a
	// channel (including threads about the channel's videos).
	AllThreadsRelatedToChannelID string
	// MaxResults is the page size. Defaults to 20.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// ModerationStatus filters by moderation status.
	ModerationStatus string
	// Order selects time or relevance ordering.
	Order string
	// SearchTerms restricts the result to threads matching the terms.
	SearchTerms string
	// TextFormat selects html or plainText for the returned comment text.
	TextFormat string
}

// ListCommentThreads returns comment threads matching the given filter. Mirrors
// commentThreads.list.
func (c *Client) ListCommentThreads(ctx context.Context, p ListCommentThreadsParams) (*CommentThreadListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListCommentThreads requires at least one part")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 20
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.VideoID != "" {
		q.Set("videoId", p.VideoID)
	}
	if p.ChannelID != "" {
		q.Set("channelId", p.ChannelID)
	}
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.AllThreadsRelatedToChannelID != "" {
		q.Set("allThreadsRelatedToChannelId", p.AllThreadsRelatedToChannelID)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.ModerationStatus != "" {
		q.Set("moderationStatus", p.ModerationStatus)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	if p.SearchTerms != "" {
		q.Set("searchTerms", p.SearchTerms)
	}
	if p.TextFormat != "" {
		q.Set("textFormat", p.TextFormat)
	}

	out := &CommentThreadListResponse{}
	if err := c.doGet(ctx, "commentThreads", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterVideoThreads walks every comment thread on a video, invoking fn for each.
// It pages through the result set until exhausted or fn returns an error. This
// is the Go-idiomatic equivalent of the Python iter_video_threads async
// generator.
func (c *Client) IterVideoThreads(
	ctx context.Context,
	videoID string,
	parts []string,
	maxResults int,
	order string,
	fn func(CommentThread) error,
) error {
	if videoID == "" {
		return errors.New("youtube: IterVideoThreads requires a videoID")
	}
	if maxResults <= 0 {
		maxResults = 100
	}
	pageToken := ""
	for {
		page, err := c.ListCommentThreads(ctx, ListCommentThreadsParams{
			Parts:      parts,
			VideoID:    videoID,
			MaxResults: maxResults,
			PageToken:  pageToken,
			Order:      order,
		})
		if err != nil {
			return err
		}
		for _, thread := range page.Items {
			if err := fn(thread); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// InsertCommentThread posts a new top-level comment on a video. The body must
// include snippet.channelId, snippet.videoId, and
// snippet.topLevelComment.snippet.textOriginal. When parts is empty it defaults
// to ["snippet"]. Mirrors commentThreads.insert. Requires OAuth.
func (c *Client) InsertCommentThread(ctx context.Context, body CommentThread, parts []string) (*CommentThread, error) {
	if len(parts) == 0 {
		parts = []string{CommentThreadPartSnippet}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &CommentThread{}
	if err := c.doJSON(ctx, http.MethodPost, "commentThreads", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}
