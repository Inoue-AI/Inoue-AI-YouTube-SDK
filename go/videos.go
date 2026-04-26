package youtube

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Video mirrors the subset of YouTube video resource fields the Inoue backend
// consumes. Fields are populated only if the corresponding "part" was requested.
type Video struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet *struct {
		PublishedAt          string   `json:"publishedAt,omitempty"`
		ChannelID            string   `json:"channelId,omitempty"`
		Title                string   `json:"title,omitempty"`
		Description          string   `json:"description,omitempty"`
		Tags                 []string `json:"tags,omitempty"`
		CategoryID           string   `json:"categoryId,omitempty"`
		LiveBroadcastContent string   `json:"liveBroadcastContent,omitempty"`
		DefaultAudioLanguage string   `json:"defaultAudioLanguage,omitempty"`
		Thumbnails           map[string]struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"thumbnails,omitempty"`
	} `json:"snippet,omitempty"`
	ContentDetails *struct {
		Duration   string `json:"duration,omitempty"`
		Dimension  string `json:"dimension,omitempty"`
		Definition string `json:"definition,omitempty"`
		Caption    string `json:"caption,omitempty"`
	} `json:"contentDetails,omitempty"`
	Statistics *struct {
		ViewCount    string `json:"viewCount,omitempty"`
		LikeCount    string `json:"likeCount,omitempty"`
		CommentCount string `json:"commentCount,omitempty"`
	} `json:"statistics,omitempty"`
	Status *struct {
		PrivacyStatus       string `json:"privacyStatus,omitempty"`
		UploadStatus        string `json:"uploadStatus,omitempty"`
		License             string `json:"license,omitempty"`
		Embeddable          bool   `json:"embeddable,omitempty"`
		PublicStatsViewable bool   `json:"publicStatsViewable,omitempty"`
		MadeForKids         bool   `json:"madeForKids,omitempty"`
	} `json:"status,omitempty"`
}

// videoListResponse is the Google API list wrapper.
type videoListResponse struct {
	Kind          string   `json:"kind"`
	Etag          string   `json:"etag"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
	PageInfo      pageInfo `json:"pageInfo"`
	Items         []Video  `json:"items"`
}

// GetVideo fetches a single video by ID. When parts is empty, a sensible
// default set is used.
func (c *Client) GetVideo(ctx context.Context, videoID string, parts []string) (*Video, error) {
	if videoID == "" {
		return nil, errors.New("youtube: GetVideo requires a non-empty video ID")
	}
	if len(parts) == 0 {
		parts = []string{"snippet", "statistics", "contentDetails", "status"}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))
	q.Set("id", videoID)

	resp := &videoListResponse{}
	if err := c.doGet(ctx, "videos", q, resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, &Error{StatusCode: 404, Status: "NOT_FOUND", Message: "video not found"}
	}
	return &resp.Items[0], nil
}
