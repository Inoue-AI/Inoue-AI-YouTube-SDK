package youtube

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// SearchResult is one item returned by Search. The ID subfields are populated
// according to the result kind.
type SearchResult struct {
	Kind string `json:"kind"`
	Etag string `json:"etag"`
	ID   struct {
		Kind       string `json:"kind"`
		VideoID    string `json:"videoId,omitempty"`
		ChannelID  string `json:"channelId,omitempty"`
		PlaylistID string `json:"playlistId,omitempty"`
	} `json:"id"`
	Snippet struct {
		PublishedAt          string `json:"publishedAt"`
		ChannelID            string `json:"channelId"`
		Title                string `json:"title"`
		Description          string `json:"description"`
		ChannelTitle         string `json:"channelTitle"`
		LiveBroadcastContent string `json:"liveBroadcastContent"`
		Thumbnails           map[string]struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnails"`
	} `json:"snippet"`
}

// SearchPage is the cursor-paginated wrapper for Search.
type SearchPage struct {
	Kind          string         `json:"kind"`
	Etag          string         `json:"etag"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
	PrevPageToken string         `json:"prevPageToken,omitempty"`
	RegionCode    string         `json:"regionCode,omitempty"`
	PageInfo      pageInfo       `json:"pageInfo"`
	Items         []SearchResult `json:"items"`
}

// SearchParams configures Search. Each call costs 100 quota units.
type SearchParams struct {
	// Query is the free-text search term.
	Query string
	// Types restricts results (e.g. ["video"], ["channel","playlist"]). When
	// empty, YouTube returns all kinds.
	Types []string
	// ChannelID restricts results to a single channel.
	ChannelID string
	// Order is the result ordering (date, rating, relevance, title,
	// videoCount, viewCount). Empty uses the API default (relevance).
	Order string
	// MaxResults is the page size (0-50). Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor from a previous response.
	PageToken string
	// PublishedAfter / PublishedBefore are RFC 3339 timestamps.
	PublishedAfter  string
	PublishedBefore string
	// RegionCode is an ISO 3166-1 alpha-2 country code.
	RegionCode string
}

// Search performs a search.list query. Works with either an OAuth token or an
// API key.
func (c *Client) Search(ctx context.Context, p SearchParams) (*SearchPage, error) {
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	if maxResults > 50 {
		maxResults = 50
	}

	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.Query != "" {
		q.Set("q", p.Query)
	}
	if len(p.Types) > 0 {
		q.Set("type", strings.Join(p.Types, ","))
	}
	if p.ChannelID != "" {
		q.Set("channelId", p.ChannelID)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.PublishedAfter != "" {
		q.Set("publishedAfter", p.PublishedAfter)
	}
	if p.PublishedBefore != "" {
		q.Set("publishedBefore", p.PublishedBefore)
	}
	if p.RegionCode != "" {
		q.Set("regionCode", p.RegionCode)
	}

	out := &SearchPage{}
	if err := c.doGet(ctx, "search", q, out); err != nil {
		return nil, err
	}
	return out, nil
}
