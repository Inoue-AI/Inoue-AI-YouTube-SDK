package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PlaylistItem parts accepted by the part parameter.
const (
	PlaylistItemPartID             = "id"
	PlaylistItemPartSnippet        = "snippet"
	PlaylistItemPartContentDetails = "contentDetails"
	PlaylistItemPartStatus         = "status"
)

// PlaylistItemSnippet is the snippet part of a PlaylistItem resource.
type PlaylistItemSnippet struct {
	PublishedAt            string            `json:"publishedAt,omitempty"`
	ChannelID              string            `json:"channelId,omitempty"`
	Title                  string            `json:"title,omitempty"`
	Description            string            `json:"description,omitempty"`
	Thumbnails             *ThumbnailDetails `json:"thumbnails,omitempty"`
	ChannelTitle           string            `json:"channelTitle,omitempty"`
	PlaylistID             string            `json:"playlistId,omitempty"`
	Position               *int              `json:"position,omitempty"`
	ResourceID             *ResourceID       `json:"resourceId,omitempty"`
	VideoOwnerChannelTitle string            `json:"videoOwnerChannelTitle,omitempty"`
	VideoOwnerChannelID    string            `json:"videoOwnerChannelId,omitempty"`
}

// PlaylistItemContentDetails is the contentDetails part of a PlaylistItem.
type PlaylistItemContentDetails struct {
	VideoID          string `json:"videoId,omitempty"`
	Note             string `json:"note,omitempty"`
	VideoPublishedAt string `json:"videoPublishedAt,omitempty"`
}

// PlaylistItemStatus is the status part of a PlaylistItem resource.
type PlaylistItemStatus struct {
	PrivacyStatus string `json:"privacyStatus,omitempty"`
}

// PlaylistItem is a YouTube PlaylistItem resource (request body and response).
type PlaylistItem struct {
	Kind           string                      `json:"kind,omitempty"`
	Etag           string                      `json:"etag,omitempty"`
	ID             string                      `json:"id,omitempty"`
	Snippet        *PlaylistItemSnippet        `json:"snippet,omitempty"`
	ContentDetails *PlaylistItemContentDetails `json:"contentDetails,omitempty"`
	Status         *PlaylistItemStatus         `json:"status,omitempty"`
}

// PlaylistItemListResponse is the response from ListPlaylistItems.
type PlaylistItemListResponse struct {
	Kind          string         `json:"kind,omitempty"`
	Etag          string         `json:"etag,omitempty"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
	PrevPageToken string         `json:"prevPageToken,omitempty"`
	PageInfo      *PageInfo      `json:"pageInfo,omitempty"`
	Items         []PlaylistItem `json:"items"`
}

// ListPlaylistItemsParams configures ListPlaylistItems.
type ListPlaylistItemsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// PlaylistID returns items from the given playlist.
	PlaylistID string
	// IDs restricts the result to specific playlist-item IDs.
	IDs []string
	// VideoID restricts the result to items referencing the given video.
	VideoID string
	// MaxResults is the page size. Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
}

// ListPlaylistItems returns items from a playlist. Mirrors playlistItems.list.
func (c *Client) ListPlaylistItems(ctx context.Context, p ListPlaylistItemsParams) (*PlaylistItemListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListPlaylistItems requires at least one part")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.PlaylistID != "" {
		q.Set("playlistId", p.PlaylistID)
	}
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.VideoID != "" {
		q.Set("videoId", p.VideoID)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}

	out := &PlaylistItemListResponse{}
	if err := c.doGet(ctx, "playlistItems", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterPlaylistItems walks every item in a playlist, invoking fn for each. It
// pages through the result set until exhausted or fn returns an error. This is
// the Go-idiomatic equivalent of the Python iter_playlist async generator.
func (c *Client) IterPlaylistItems(
	ctx context.Context,
	playlistID string,
	parts []string,
	maxResults int,
	fn func(PlaylistItem) error,
) error {
	if playlistID == "" {
		return errors.New("youtube: IterPlaylistItems requires a playlistID")
	}
	if maxResults <= 0 {
		maxResults = 50
	}
	pageToken := ""
	for {
		page, err := c.ListPlaylistItems(ctx, ListPlaylistItemsParams{
			Parts:      parts,
			PlaylistID: playlistID,
			MaxResults: maxResults,
			PageToken:  pageToken,
		})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if err := fn(item); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// InsertPlaylistItem adds a video to a playlist. The body must include
// snippet.playlistId and snippet.resourceId (with kind and videoId). When parts
// is empty, the parts are inferred from the populated fields of the body.
// Mirrors playlistItems.insert. Requires OAuth.
func (c *Client) InsertPlaylistItem(ctx context.Context, body PlaylistItem, parts []string) (*PlaylistItem, error) {
	if len(parts) == 0 {
		parts = inferPlaylistItemParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &PlaylistItem{}
	if err := c.doJSON(ctx, http.MethodPost, "playlistItems", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdatePlaylistItem updates a playlist item (e.g. change position or note).
// When parts is empty, the parts are inferred from the populated fields of the
// body. Mirrors playlistItems.update. Requires OAuth.
func (c *Client) UpdatePlaylistItem(ctx context.Context, body PlaylistItem, parts []string) (*PlaylistItem, error) {
	if len(parts) == 0 {
		parts = inferPlaylistItemParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &PlaylistItem{}
	if err := c.doJSON(ctx, http.MethodPut, "playlistItems", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeletePlaylistItem removes an item from a playlist by ID. Mirrors
// playlistItems.delete. Requires OAuth.
func (c *Client) DeletePlaylistItem(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeletePlaylistItem requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "playlistItems", q)
}

// inferPlaylistItemParts derives the resource parts from the populated fields of
// a PlaylistItem body, mirroring the Python _infer_parts helper.
func inferPlaylistItemParts(body PlaylistItem) []string {
	parts := make([]string, 0, 2)
	if body.Snippet != nil {
		parts = append(parts, PlaylistItemPartSnippet)
	}
	if body.ContentDetails != nil {
		parts = append(parts, PlaylistItemPartContentDetails)
	}
	if len(parts) == 0 {
		return []string{PlaylistItemPartSnippet}
	}
	return parts
}
