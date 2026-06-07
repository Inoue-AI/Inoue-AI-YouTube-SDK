package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Comment parts accepted by the part parameter.
const (
	CommentPartID      = "id"
	CommentPartSnippet = "snippet"
)

// Moderation status values for SetCommentModerationStatus.
const (
	ModerationStatusHeldForReview = "heldForReview"
	ModerationStatusPublished     = "published"
	ModerationStatusRejected      = "rejected"
)

// Text format options for comment listing.
const (
	TextFormatHTML      = "html"
	TextFormatPlainText = "plainText"
)

// AuthorChannelID wraps the author's channel ID inside a comment snippet.
type AuthorChannelID struct {
	Value string `json:"value,omitempty"`
}

// CommentSnippet is the snippet part of a Comment resource.
type CommentSnippet struct {
	AuthorDisplayName     string           `json:"authorDisplayName,omitempty"`
	AuthorProfileImageURL string           `json:"authorProfileImageUrl,omitempty"`
	AuthorChannelURL      string           `json:"authorChannelUrl,omitempty"`
	AuthorChannelID       *AuthorChannelID `json:"authorChannelId,omitempty"`
	ChannelID             string           `json:"channelId,omitempty"`
	VideoID               string           `json:"videoId,omitempty"`
	TextDisplay           string           `json:"textDisplay,omitempty"`
	TextOriginal          string           `json:"textOriginal,omitempty"`
	ParentID              string           `json:"parentId,omitempty"`
	CanRate               *bool            `json:"canRate,omitempty"`
	ViewerRating          string           `json:"viewerRating,omitempty"`
	LikeCount             *int             `json:"likeCount,omitempty"`
	ModerationStatus      string           `json:"moderationStatus,omitempty"`
	PublishedAt           string           `json:"publishedAt,omitempty"`
	UpdatedAt             string           `json:"updatedAt,omitempty"`
}

// Comment is a YouTube Comment resource (request body and response).
type Comment struct {
	Kind    string          `json:"kind,omitempty"`
	Etag    string          `json:"etag,omitempty"`
	ID      string          `json:"id,omitempty"`
	Snippet *CommentSnippet `json:"snippet,omitempty"`
}

// CommentListResponse is the response from ListComments.
type CommentListResponse struct {
	Kind          string    `json:"kind,omitempty"`
	Etag          string    `json:"etag,omitempty"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
	PageInfo      *PageInfo `json:"pageInfo,omitempty"`
	Items         []Comment `json:"items"`
}

// ListCommentsParams configures ListComments.
type ListCommentsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// IDs restricts the result to specific comment IDs.
	IDs []string
	// ParentID retrieves replies to a specific comment.
	ParentID string
	// MaxResults is the page size. Defaults to 20.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// TextFormat selects html or plainText for the returned comment text.
	TextFormat string
}

// ListComments returns comments matching the given filter. Use ParentID to
// retrieve replies to a specific comment. Mirrors comments.list.
func (c *Client) ListComments(ctx context.Context, p ListCommentsParams) (*CommentListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListComments requires at least one part")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 20
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("maxResults", strconv.Itoa(maxResults))
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.ParentID != "" {
		q.Set("parentId", p.ParentID)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.TextFormat != "" {
		q.Set("textFormat", p.TextFormat)
	}

	out := &CommentListResponse{}
	if err := c.doGet(ctx, "comments", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterReplies walks every reply to a comment, invoking fn for each. It pages
// through the result set until exhausted or fn returns an error. This is the
// Go-idiomatic equivalent of the Python iter_replies async generator.
func (c *Client) IterReplies(
	ctx context.Context,
	parentID string,
	parts []string,
	maxResults int,
	fn func(Comment) error,
) error {
	if parentID == "" {
		return errors.New("youtube: IterReplies requires a parentID")
	}
	if maxResults <= 0 {
		maxResults = 100
	}
	pageToken := ""
	for {
		page, err := c.ListComments(ctx, ListCommentsParams{
			Parts:      parts,
			ParentID:   parentID,
			MaxResults: maxResults,
			PageToken:  pageToken,
		})
		if err != nil {
			return err
		}
		for _, comment := range page.Items {
			if err := fn(comment); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// InsertComment posts a reply to an existing comment. The body must include
// snippet.parentId and snippet.textOriginal. Use InsertCommentThread for
// top-level comments. When parts is empty it defaults to ["snippet"]. Mirrors
// comments.insert. Requires OAuth.
func (c *Client) InsertComment(ctx context.Context, body Comment, parts []string) (*Comment, error) {
	if len(parts) == 0 {
		parts = []string{CommentPartSnippet}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Comment{}
	if err := c.doJSON(ctx, http.MethodPost, "comments", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateComment updates an existing comment's text. The body must include id
// and snippet.textOriginal. When parts is empty it defaults to ["snippet"].
// Mirrors comments.update. Requires OAuth.
func (c *Client) UpdateComment(ctx context.Context, body Comment, parts []string) (*Comment, error) {
	if len(parts) == 0 {
		parts = []string{CommentPartSnippet}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Comment{}
	if err := c.doJSON(ctx, http.MethodPut, "comments", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteComment deletes a comment by ID. Mirrors comments.delete. Requires
// OAuth.
func (c *Client) DeleteComment(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeleteComment requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "comments", q)
}

// SetCommentModerationStatusParams configures SetCommentModerationStatus.
type SetCommentModerationStatusParams struct {
	// IDs is one or more comment IDs to moderate. Required.
	IDs []string
	// ModerationStatus is the new moderation status. Required.
	ModerationStatus string
	// BanAuthor, when true, also bans the author from commenting.
	BanAuthor bool
}

// SetCommentModerationStatus sets the moderation status of one or more
// comments. Mirrors comments.set_moderation_status. Requires OAuth.
func (c *Client) SetCommentModerationStatus(ctx context.Context, p SetCommentModerationStatusParams) error {
	if len(p.IDs) == 0 {
		return errors.New("youtube: SetCommentModerationStatus requires at least one ID")
	}
	if p.ModerationStatus == "" {
		return errors.New("youtube: SetCommentModerationStatus requires a ModerationStatus")
	}
	q := url.Values{}
	q.Set("id", strings.Join(p.IDs, ","))
	q.Set("moderationStatus", p.ModerationStatus)
	if p.BanAuthor {
		q.Set("banAuthor", "true")
	}
	return c.doPostNoContent(ctx, "comments/setModerationStatus", q)
}
