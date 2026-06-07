package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Channel parts accepted by the part parameter.
const (
	ChannelPartID               = "id"
	ChannelPartSnippet          = "snippet"
	ChannelPartContentDetails   = "contentDetails"
	ChannelPartStatistics       = "statistics"
	ChannelPartTopicDetails     = "topicDetails"
	ChannelPartStatus           = "status"
	ChannelPartBrandingSettings = "brandingSettings"
	ChannelPartLocalizations    = "localizations"
)

// ChannelSnippet is the snippet part of a Channel resource.
type ChannelSnippet struct {
	Title           string            `json:"title,omitempty"`
	Description     string            `json:"description,omitempty"`
	CustomURL       string            `json:"customUrl,omitempty"`
	PublishedAt     string            `json:"publishedAt,omitempty"`
	Thumbnails      *ThumbnailDetails `json:"thumbnails,omitempty"`
	DefaultLanguage string            `json:"defaultLanguage,omitempty"`
	Localized       *Localization     `json:"localized,omitempty"`
	Country         string            `json:"country,omitempty"`
}

// ChannelRelatedPlaylists holds the auto-generated playlists linked to a channel.
type ChannelRelatedPlaylists struct {
	Likes   string `json:"likes,omitempty"`
	Uploads string `json:"uploads,omitempty"`
}

// ChannelContentDetails is the contentDetails part of a Channel resource.
type ChannelContentDetails struct {
	RelatedPlaylists *ChannelRelatedPlaylists `json:"relatedPlaylists,omitempty"`
}

// ChannelStatistics is the statistics part of a Channel resource.
type ChannelStatistics struct {
	ViewCount             string `json:"viewCount,omitempty"`
	SubscriberCount       string `json:"subscriberCount,omitempty"`
	HiddenSubscriberCount bool   `json:"hiddenSubscriberCount,omitempty"`
	VideoCount            string `json:"videoCount,omitempty"`
}

// ChannelTopicDetails is the topicDetails part of a Channel resource.
type ChannelTopicDetails struct {
	TopicIDs        []string `json:"topicIds,omitempty"`
	TopicCategories []string `json:"topicCategories,omitempty"`
}

// ChannelStatus is the status part of a Channel resource.
type ChannelStatus struct {
	PrivacyStatus           string `json:"privacyStatus,omitempty"`
	IsLinked                *bool  `json:"isLinked,omitempty"`
	LongUploadsStatus       string `json:"longUploadsStatus,omitempty"`
	MadeForKids             *bool  `json:"madeForKids,omitempty"`
	SelfDeclaredMadeForKids *bool  `json:"selfDeclaredMadeForKids,omitempty"`
}

// ChannelBrandingChannel holds branding settings for the channel itself.
type ChannelBrandingChannel struct {
	Title                      string `json:"title,omitempty"`
	Description                string `json:"description,omitempty"`
	Keywords                   string `json:"keywords,omitempty"`
	TrackingAnalyticsAccountID string `json:"trackingAnalyticsAccountId,omitempty"`
	UnsubscribedTrailer        string `json:"unsubscribedTrailer,omitempty"`
	DefaultLanguage            string `json:"defaultLanguage,omitempty"`
	Country                    string `json:"country,omitempty"`
}

// ChannelBrandingImage holds branding settings for channel images.
type ChannelBrandingImage struct {
	BannerExternalURL string `json:"bannerExternalUrl,omitempty"`
}

// ChannelBrandingSettings is the brandingSettings part of a Channel resource.
type ChannelBrandingSettings struct {
	Channel *ChannelBrandingChannel `json:"channel,omitempty"`
	Image   *ChannelBrandingImage   `json:"image,omitempty"`
}

// Channel is a YouTube Channel resource. It is used both as a response model
// and as the request body for UpdateChannel. Fields are populated only if the
// corresponding part was requested.
type Channel struct {
	Kind             string                   `json:"kind,omitempty"`
	Etag             string                   `json:"etag,omitempty"`
	ID               string                   `json:"id,omitempty"`
	Snippet          *ChannelSnippet          `json:"snippet,omitempty"`
	ContentDetails   *ChannelContentDetails   `json:"contentDetails,omitempty"`
	Statistics       *ChannelStatistics       `json:"statistics,omitempty"`
	TopicDetails     *ChannelTopicDetails     `json:"topicDetails,omitempty"`
	Status           *ChannelStatus           `json:"status,omitempty"`
	BrandingSettings *ChannelBrandingSettings `json:"brandingSettings,omitempty"`
	Localizations    map[string]Localization  `json:"localizations,omitempty"`
}

// ChannelListResponse is the Google API list wrapper for channels.
type ChannelListResponse struct {
	Kind          string    `json:"kind,omitempty"`
	Etag          string    `json:"etag,omitempty"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
	PrevPageToken string    `json:"prevPageToken,omitempty"`
	PageInfo      *PageInfo `json:"pageInfo,omitempty"`
	Items         []Channel `json:"items"`
}

// channelListResponse is the internal list wrapper retained for GetChannel.
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

// ListChannelsParams configures ListChannels. Exactly one of ID, ForHandle,
// ForUsername, Mine, or ManagedByMe should be supplied.
type ListChannelsParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// IDs restricts the result to specific channel IDs.
	IDs []string
	// ForHandle looks up a channel by its handle (e.g. "@handle").
	ForHandle string
	// ForUsername looks up a channel by its legacy username.
	ForUsername string
	// Mine, when non-nil, restricts the result to the authenticated user's
	// channel.
	Mine *bool
	// ManagedByMe, when non-nil, restricts to channels managed by the
	// authenticated content owner.
	ManagedByMe *bool
	// MaxResults is the page size. Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// HL is the language for localized metadata.
	HL string
}

// ListChannels returns channels matching the given filter. Mirrors
// channels.list.
func (c *Client) ListChannels(ctx context.Context, p ListChannelsParams) (*ChannelListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListChannels requires at least one part")
	}
	maxResults := p.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("maxResults", strconv.Itoa(maxResults))
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}
	if p.ForHandle != "" {
		q.Set("forHandle", p.ForHandle)
	}
	if p.ForUsername != "" {
		q.Set("forUsername", p.ForUsername)
	}
	if p.Mine != nil {
		q.Set("mine", strconv.FormatBool(*p.Mine))
	}
	if p.ManagedByMe != nil {
		q.Set("managedByMe", strconv.FormatBool(*p.ManagedByMe))
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.HL != "" {
		q.Set("hl", p.HL)
	}

	out := &ChannelListResponse{}
	if err := c.doGet(ctx, "channels", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateChannel updates a channel's metadata. The body must contain ID. Only
// brandingSettings, localizations, and status.selfDeclaredMadeForKids are
// writable. Properties not included in the body are deleted. When parts is
// empty, the parts are inferred from the populated fields of the body. Mirrors
// channels.update. Requires OAuth.
func (c *Client) UpdateChannel(ctx context.Context, body Channel, parts []string) (*Channel, error) {
	if body.ID == "" {
		return nil, errors.New("youtube: UpdateChannel requires body.ID")
	}
	if len(parts) == 0 {
		parts = inferChannelParts(body)
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Channel{}
	if err := c.doJSON(ctx, http.MethodPut, "channels", q, body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// inferChannelParts derives the resource parts from the populated fields of a
// Channel body, mirroring the Python _infer_parts helper.
func inferChannelParts(body Channel) []string {
	parts := make([]string, 0, 3)
	if body.BrandingSettings != nil {
		parts = append(parts, ChannelPartBrandingSettings)
	}
	if body.Localizations != nil {
		parts = append(parts, ChannelPartLocalizations)
	}
	if body.Status != nil {
		parts = append(parts, ChannelPartStatus)
	}
	if len(parts) == 0 {
		return []string{ChannelPartBrandingSettings}
	}
	return parts
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
		parts = []string{ChannelPartSnippet, ChannelPartStatistics, ChannelPartContentDetails}
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
		PublishedAt string `json:"publishedAt"`
		ChannelID   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ResourceID  struct {
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
