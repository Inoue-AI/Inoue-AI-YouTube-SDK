package youtube

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// Channel mirrors the subset of YouTube channel resource fields the Inoue
// backend consumes. Fields are populated only if the corresponding "part"
// was requested.
type Channel struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet *struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		CustomURL   string `json:"customUrl,omitempty"`
		PublishedAt string `json:"publishedAt,omitempty"`
		Country     string `json:"country,omitempty"`
		Thumbnails  map[string]struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnails,omitempty"`
	} `json:"snippet,omitempty"`
	Statistics *struct {
		ViewCount             string `json:"viewCount,omitempty"`
		SubscriberCount       string `json:"subscriberCount,omitempty"`
		HiddenSubscriberCount bool   `json:"hiddenSubscriberCount,omitempty"`
		VideoCount            string `json:"videoCount,omitempty"`
	} `json:"statistics,omitempty"`
	ContentDetails *struct {
		RelatedPlaylists struct {
			Likes   string `json:"likes,omitempty"`
			Uploads string `json:"uploads,omitempty"`
		} `json:"relatedPlaylists,omitempty"`
	} `json:"contentDetails,omitempty"`
}

// channelListResponse is the Google API list wrapper.
type channelListResponse struct {
	Kind          string    `json:"kind"`
	Etag          string    `json:"etag"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
	PrevPageToken string    `json:"prevPageToken,omitempty"`
	PageInfo      pageInfo  `json:"pageInfo"`
	Items         []Channel `json:"items"`
}

type pageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}

// GetChannelParams configures GetChannel. Exactly one of ID, ForHandle,
// ForUsername, or Mine must be set.
type GetChannelParams struct {
	Parts       []string // Defaults to ["snippet","statistics","contentDetails"].
	ID          string
	ForHandle   string
	ForUsername string
	Mine        bool
}

// GetChannel returns the first matching channel resource. When the upstream
// returns no matches it returns a *Error with StatusCode=404 so callers can
// branch on AsError(...).IsNotFound().
func (c *Client) GetChannel(ctx context.Context, p GetChannelParams) (*Channel, error) {
	parts := p.Parts
	if len(parts) == 0 {
		parts = []string{"snippet", "statistics", "contentDetails"}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	switch {
	case p.ID != "":
		q.Set("id", p.ID)
	case p.ForHandle != "":
		q.Set("forHandle", p.ForHandle)
	case p.ForUsername != "":
		q.Set("forUsername", p.ForUsername)
	case p.Mine:
		q.Set("mine", "true")
	default:
		return nil, errors.New("youtube: GetChannel requires one of ID, ForHandle, ForUsername, or Mine")
	}

	resp := &channelListResponse{}
	if err := c.doGet(ctx, "channels", q, resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &Error{StatusCode: 404, Status: "NOT_FOUND", Message: "no channel matched"}
	}
	return &resp.Items[0], nil
}

// ListChannelVideosParams configures ListChannelVideos.
type ListChannelVideosParams struct {
	// UploadsPlaylistID is the channel's uploads playlist ID (typically
	// channel.contentDetails.relatedPlaylists.uploads). Required.
	UploadsPlaylistID string
	// PageToken is an optional pagination token (nextPageToken from a
	// previous response).
	PageToken string
	// MaxResults is the page size (1-50). Defaults to 50.
	MaxResults int
}

// ChannelVideo is the playlistItems projection of a video on the channel's
// uploads playlist.
type ChannelVideo struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt  string `json:"publishedAt"`
		ChannelID    string `json:"channelId"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		ResourceID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"resourceId"`
		Thumbnails map[string]struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnails"`
	} `json:"snippet"`
}

// ChannelVideosPage is the cursor-paginated wrapper for ListChannelVideos.
type ChannelVideosPage struct {
	NextPageToken string         `json:"nextPageToken,omitempty"`
	PrevPageToken string         `json:"prevPageToken,omitempty"`
	PageInfo      pageInfo       `json:"pageInfo"`
	Items         []ChannelVideo `json:"items"`
}

// ListChannelVideos fetches one page of videos uploaded to the given uploads
// playlist. Use Channel.ContentDetails.RelatedPlaylists.Uploads to discover
// the playlist ID for a channel.
func (c *Client) ListChannelVideos(ctx context.Context, p ListChannelVideosParams) (*ChannelVideosPage, error) {
	if p.UploadsPlaylistID == "" {
		return nil, errors.New("youtube: ListChannelVideos requires UploadsPlaylistID")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}
	if maxResults > 50 {
		maxResults = 50
	}

	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("playlistId", p.UploadsPlaylistID)
	q.Set("maxResults", strconv.Itoa(maxResults))
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}

	out := &ChannelVideosPage{}
	if err := c.doGet(ctx, "playlistItems", q, out); err != nil {
		return nil, err
	}
	return out, nil
}
