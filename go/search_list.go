package youtube

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// Search sort orders.
const (
	SearchOrderDate       = "date"
	SearchOrderRating     = "rating"
	SearchOrderRelevance  = "relevance"
	SearchOrderTitle      = "title"
	SearchOrderVideoCount = "videoCount"
	SearchOrderViewCount  = "viewCount"
)

// Safe-search filtering levels.
const (
	SafeSearchModerate = "moderate"
	SafeSearchNone     = "none"
	SafeSearchStrict   = "strict"
)

// Search result types.
const (
	SearchResultTypeChannel  = "channel"
	SearchResultTypePlaylist = "playlist"
	SearchResultTypeVideo    = "video"
)

// Live event type filters.
const (
	EventTypeCompleted = "completed"
	EventTypeLive      = "live"
	EventTypeUpcoming  = "upcoming"
)

// Caption availability filters.
const (
	VideoCaptionAny           = "any"
	VideoCaptionClosedCaption = "closedCaption"
	VideoCaptionNone          = "none"
)

// Video definition filters.
const (
	VideoDefinitionAny      = "any"
	VideoDefinitionHigh     = "high"
	VideoDefinitionStandard = "standard"
)

// Video dimension filters.
const (
	VideoDimensionAny   = "any"
	VideoDimensionTwoD  = "2d"
	VideoDimensionThree = "3d"
)

// Video duration filters.
const (
	VideoDurationAny    = "any"
	VideoDurationLong   = "long"
	VideoDurationMedium = "medium"
	VideoDurationShort  = "short"
)

// Video type filters.
const (
	VideoTypeAny     = "any"
	VideoTypeEpisode = "episode"
	VideoTypeMovie   = "movie"
)

// SearchResultSnippet is the snippet part of a SearchItem resource.
type SearchResultSnippet struct {
	PublishedAt          string            `json:"publishedAt,omitempty"`
	ChannelID            string            `json:"channelId,omitempty"`
	Title                string            `json:"title,omitempty"`
	Description          string            `json:"description,omitempty"`
	Thumbnails           *ThumbnailDetails `json:"thumbnails,omitempty"`
	ChannelTitle         string            `json:"channelTitle,omitempty"`
	LiveBroadcastContent string            `json:"liveBroadcastContent,omitempty"`
}

// SearchItem is a single search result. Unlike other resources, its ID is a
// ResourceID object (not a plain string). This is the full-parity equivalent of
// the Python SearchResult model.
type SearchItem struct {
	Kind    string               `json:"kind,omitempty"`
	Etag    string               `json:"etag,omitempty"`
	ID      *ResourceID          `json:"id,omitempty"`
	Snippet *SearchResultSnippet `json:"snippet,omitempty"`
}

// SearchListResponse is the response from ListSearch.
type SearchListResponse struct {
	Kind          string       `json:"kind,omitempty"`
	Etag          string       `json:"etag,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
	PrevPageToken string       `json:"prevPageToken,omitempty"`
	RegionCode    string       `json:"regionCode,omitempty"`
	PageInfo      *PageInfo    `json:"pageInfo,omitempty"`
	Items         []SearchItem `json:"items"`
}

// ListSearchParams configures ListSearch. This mirrors the full parameter
// surface of the Python search.list method. Each call costs 100 quota units.
type ListSearchParams struct {
	Q                 string
	Types             []string // e.g. ["video","channel"].
	ChannelID         string
	ForMine           *bool
	Order             string
	MaxResults        int // Defaults to 5.
	PageToken         string
	PublishedAfter    string
	PublishedBefore   string
	RegionCode        string
	RelevanceLanguage string
	SafeSearch        string
	EventType         string
	VideoCaption      string
	VideoCategoryID   string
	VideoDefinition   string
	VideoDimension    string
	VideoDuration     string
	VideoEmbeddable   *bool
	VideoLicense      string
	VideoType         string
}

// ListSearch performs a full-parity search.list query. Works with either an
// OAuth token or an API key. Mirrors search.list.
func (c *Client) ListSearch(ctx context.Context, p ListSearchParams) (*SearchListResponse, error) {
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.Q != "" {
		q.Set("q", p.Q)
	}
	if len(p.Types) > 0 {
		q.Set("type", strings.Join(p.Types, ","))
	}
	if p.ChannelID != "" {
		q.Set("channelId", p.ChannelID)
	}
	if p.ForMine != nil {
		q.Set("forMine", strconv.FormatBool(*p.ForMine))
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
	if p.RelevanceLanguage != "" {
		q.Set("relevanceLanguage", p.RelevanceLanguage)
	}
	if p.SafeSearch != "" {
		q.Set("safeSearch", p.SafeSearch)
	}
	if p.EventType != "" {
		q.Set("eventType", p.EventType)
	}
	if p.VideoCaption != "" {
		q.Set("videoCaption", p.VideoCaption)
	}
	if p.VideoCategoryID != "" {
		q.Set("videoCategoryId", p.VideoCategoryID)
	}
	if p.VideoDefinition != "" {
		q.Set("videoDefinition", p.VideoDefinition)
	}
	if p.VideoDimension != "" {
		q.Set("videoDimension", p.VideoDimension)
	}
	if p.VideoDuration != "" {
		q.Set("videoDuration", p.VideoDuration)
	}
	if p.VideoEmbeddable != nil {
		q.Set("videoEmbeddable", strconv.FormatBool(*p.VideoEmbeddable))
	}
	if p.VideoLicense != "" {
		q.Set("videoLicense", p.VideoLicense)
	}
	if p.VideoType != "" {
		q.Set("videoType", p.VideoType)
	}

	out := &SearchListResponse{}
	if err := c.doGet(ctx, "search", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterSearchResults walks search results, invoking fn for each. It pages through
// the result set up to maxPages API calls (each page costs 100 quota units) or
// until fn returns an error. This is the Go-idiomatic equivalent of the Python
// iter_results async generator.
func (c *Client) IterSearchResults(
	ctx context.Context,
	base ListSearchParams,
	maxPages int,
	fn func(SearchItem) error,
) error {
	if maxPages <= 0 {
		maxPages = 5
	}
	if base.MaxResults <= 0 {
		base.MaxResults = 25
	}
	pageToken := ""
	for i := 0; i < maxPages; i++ {
		base.PageToken = pageToken
		page, err := c.ListSearch(ctx, base)
		if err != nil {
			return err
		}
		for _, result := range page.Items {
			if err := fn(result); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
	return nil
}
