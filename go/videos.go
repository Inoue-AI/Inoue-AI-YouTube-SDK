package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Video parts accepted by the part parameter.
const (
	VideoPartID                   = "id"
	VideoPartSnippet              = "snippet"
	VideoPartContentDetails       = "contentDetails"
	VideoPartStatus               = "status"
	VideoPartStatistics           = "statistics"
	VideoPartPlayer               = "player"
	VideoPartTopicDetails         = "topicDetails"
	VideoPartRecordingDetails     = "recordingDetails"
	VideoPartFileDetails          = "fileDetails"
	VideoPartProcessingDetails    = "processingDetails"
	VideoPartSuggestions          = "suggestions"
	VideoPartLiveStreamingDetails = "liveStreamingDetails"
	VideoPartLocalizations        = "localizations"
)

// Upload status values for a video (status.uploadStatus).
const (
	UploadStatusDeleted   = "deleted"
	UploadStatusFailed    = "failed"
	UploadStatusProcessed = "processed"
	UploadStatusRejected  = "rejected"
	UploadStatusUploaded  = "uploaded"
)

// Video license values (status.license).
const (
	VideoLicenseCreativeCommon = "creativeCommon"
	VideoLicenseYouTube        = "youtube"
)

// Viewer rating values for RateVideo / GetVideoRating.
const (
	RatingLike    = "like"
	RatingDislike = "dislike"
	RatingNone    = "none"
)

// VideoSnippet is the snippet part of a Video resource.
type VideoSnippet struct {
	PublishedAt          string            `json:"publishedAt,omitempty"`
	ChannelID            string            `json:"channelId,omitempty"`
	Title                string            `json:"title,omitempty"`
	Description          string            `json:"description,omitempty"`
	Thumbnails           *ThumbnailDetails `json:"thumbnails,omitempty"`
	ChannelTitle         string            `json:"channelTitle,omitempty"`
	Tags                 []string          `json:"tags,omitempty"`
	CategoryID           string            `json:"categoryId,omitempty"`
	LiveBroadcastContent string            `json:"liveBroadcastContent,omitempty"`
	DefaultLanguage      string            `json:"defaultLanguage,omitempty"`
	Localized            *Localization     `json:"localized,omitempty"`
	DefaultAudioLanguage string            `json:"defaultAudioLanguage,omitempty"`
}

// VideoContentDetails is the contentDetails part of a Video resource.
type VideoContentDetails struct {
	Duration        string `json:"duration,omitempty"`
	Dimension       string `json:"dimension,omitempty"`
	Definition      string `json:"definition,omitempty"`
	Caption         string `json:"caption,omitempty"`
	LicensedContent *bool  `json:"licensedContent,omitempty"`
	Projection      string `json:"projection,omitempty"`
}

// VideoStatus is the status part of a Video resource.
type VideoStatus struct {
	UploadStatus            string `json:"uploadStatus,omitempty"`
	FailureReason           string `json:"failureReason,omitempty"`
	RejectionReason         string `json:"rejectionReason,omitempty"`
	PrivacyStatus           string `json:"privacyStatus,omitempty"`
	PublishAt               string `json:"publishAt,omitempty"`
	License                 string `json:"license,omitempty"`
	Embeddable              *bool  `json:"embeddable,omitempty"`
	PublicStatsViewable     *bool  `json:"publicStatsViewable,omitempty"`
	MadeForKids             *bool  `json:"madeForKids,omitempty"`
	SelfDeclaredMadeForKids *bool  `json:"selfDeclaredMadeForKids,omitempty"`
}

// VideoStatistics is the statistics part of a Video resource. YouTube returns
// all counts as strings to handle large numbers.
type VideoStatistics struct {
	ViewCount     string `json:"viewCount,omitempty"`
	LikeCount     string `json:"likeCount,omitempty"`
	DislikeCount  string `json:"dislikeCount,omitempty"`
	FavoriteCount string `json:"favoriteCount,omitempty"`
	CommentCount  string `json:"commentCount,omitempty"`
}

// VideoPlayer is the player part of a Video resource.
type VideoPlayer struct {
	EmbedHTML   string `json:"embedHtml,omitempty"`
	EmbedHeight *int   `json:"embedHeight,omitempty"`
	EmbedWidth  *int   `json:"embedWidth,omitempty"`
}

// VideoTopicDetails is the topicDetails part of a Video resource.
type VideoTopicDetails struct {
	TopicIDs         []string `json:"topicIds,omitempty"`
	RelevantTopicIDs []string `json:"relevantTopicIds,omitempty"`
	TopicCategories  []string `json:"topicCategories,omitempty"`
}

// VideoRecordingDetails is the recordingDetails part of a Video resource.
type VideoRecordingDetails struct {
	RecordingDate string `json:"recordingDate,omitempty"`
}

// VideoLiveStreamingDetails is the liveStreamingDetails part of a Video.
type VideoLiveStreamingDetails struct {
	ActualStartTime    string `json:"actualStartTime,omitempty"`
	ActualEndTime      string `json:"actualEndTime,omitempty"`
	ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	ScheduledEndTime   string `json:"scheduledEndTime,omitempty"`
	ConcurrentViewers  string `json:"concurrentViewers,omitempty"`
	ActiveLiveChatID   string `json:"activeLiveChatId,omitempty"`
}

// Video is a YouTube Video resource. It is the response model for GetVideo /
// ListVideos and the decoded result of InsertVideo / UpdateVideo. Fields are
// populated only if the corresponding part was requested. (Video insert/update
// request bodies use the dedicated VideoResource type in upload.go.)
type Video struct {
	Kind                 string                     `json:"kind,omitempty"`
	Etag                 string                     `json:"etag,omitempty"`
	ID                   string                     `json:"id,omitempty"`
	Snippet              *VideoSnippet              `json:"snippet,omitempty"`
	ContentDetails       *VideoContentDetails       `json:"contentDetails,omitempty"`
	Status               *VideoStatus               `json:"status,omitempty"`
	Statistics           *VideoStatistics           `json:"statistics,omitempty"`
	Player               *VideoPlayer               `json:"player,omitempty"`
	TopicDetails         *VideoTopicDetails         `json:"topicDetails,omitempty"`
	RecordingDetails     *VideoRecordingDetails     `json:"recordingDetails,omitempty"`
	LiveStreamingDetails *VideoLiveStreamingDetails `json:"liveStreamingDetails,omitempty"`
	Localizations        map[string]Localization    `json:"localizations,omitempty"`
}

// VideoListResponse is the response from ListVideos.
type VideoListResponse struct {
	Kind          string    `json:"kind,omitempty"`
	Etag          string    `json:"etag,omitempty"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
	PrevPageToken string    `json:"prevPageToken,omitempty"`
	PageInfo      *PageInfo `json:"pageInfo,omitempty"`
	Items         []Video   `json:"items"`
}

// videoListResponse is the internal list wrapper retained for GetVideo.
type videoListResponse struct {
	Kind          string   `json:"kind"`
	Etag          string   `json:"etag"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
	PageInfo      pageInfo `json:"pageInfo"`
	Items         []Video  `json:"items"`
}

// VideoRating is a single item from the GetVideoRating response.
type VideoRating struct {
	VideoID string `json:"videoId,omitempty"`
	Rating  string `json:"rating,omitempty"`
}

// VideoGetRatingResponse is the response from GetVideoRating.
type VideoGetRatingResponse struct {
	Kind  string        `json:"kind,omitempty"`
	Etag  string        `json:"etag,omitempty"`
	Items []VideoRating `json:"items"`
}

// ListVideosParams configures ListVideos. Exactly one of IDs, Chart, or
// MyRating should be supplied.
type ListVideosParams struct {
	// Parts is the list of resource parts to include. Required.
	Parts []string
	// IDs returns specific videos by ID.
	IDs []string
	// Chart returns chart videos (e.g. "mostPopular").
	Chart string
	// MyRating returns videos the authenticated user rated ("like"/"dislike").
	MyRating string
	// MaxResults is the page size. Defaults to 5.
	MaxResults int
	// PageToken is the pagination cursor.
	PageToken string
	// RegionCode restricts chart results to a country.
	RegionCode string
	// VideoCategoryID restricts chart results to a category.
	VideoCategoryID string
	// HL is the language for localized metadata.
	HL string
}

// ListVideos returns videos matching the given filter. Mirrors videos.list.
func (c *Client) ListVideos(ctx context.Context, p ListVideosParams) (*VideoListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListVideos requires at least one part")
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
	if p.Chart != "" {
		q.Set("chart", p.Chart)
	}
	if p.MyRating != "" {
		q.Set("myRating", p.MyRating)
	}
	if p.PageToken != "" {
		q.Set("pageToken", p.PageToken)
	}
	if p.RegionCode != "" {
		q.Set("regionCode", p.RegionCode)
	}
	if p.VideoCategoryID != "" {
		q.Set("videoCategoryId", p.VideoCategoryID)
	}
	if p.HL != "" {
		q.Set("hl", p.HL)
	}

	out := &VideoListResponse{}
	if err := c.doGet(ctx, "videos", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterVideosByChart walks chart results, invoking fn for each video. It pages
// through the result set until exhausted or fn returns an error. This is the
// Go-idiomatic equivalent of the Python iter_by_chart async generator.
func (c *Client) IterVideosByChart(
	ctx context.Context,
	parts []string,
	chart string,
	maxResults int,
	regionCode string,
	videoCategoryID string,
	fn func(Video) error,
) error {
	if chart == "" {
		chart = "mostPopular"
	}
	if maxResults <= 0 {
		maxResults = 5
	}
	pageToken := ""
	for {
		page, err := c.ListVideos(ctx, ListVideosParams{
			Parts:           parts,
			Chart:           chart,
			MaxResults:      maxResults,
			PageToken:       pageToken,
			RegionCode:      regionCode,
			VideoCategoryID: videoCategoryID,
		})
		if err != nil {
			return err
		}
		for _, video := range page.Items {
			if err := fn(video); err != nil {
				return err
			}
		}
		if page.NextPageToken == "" {
			return nil
		}
		pageToken = page.NextPageToken
	}
}

// GetVideo fetches a single video by ID. When parts is empty, a sensible
// default set is used. When the upstream returns no matches it returns a *Error
// with StatusCode=404.
func (c *Client) GetVideo(ctx context.Context, videoID string, parts []string) (*Video, error) {
	if videoID == "" {
		return nil, errors.New("youtube: GetVideo requires a non-empty video ID")
	}
	if len(parts) == 0 {
		parts = []string{VideoPartSnippet, VideoPartStatistics, VideoPartContentDetails, VideoPartStatus}
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

// RateVideo likes, dislikes, or removes the authenticated user's rating on a
// video. Rating must be one of RatingLike, RatingDislike, or RatingNone.
// Mirrors videos.rate. Requires OAuth.
func (c *Client) RateVideo(ctx context.Context, videoID, rating string) error {
	if videoID == "" {
		return errors.New("youtube: RateVideo requires a video ID")
	}
	if rating == "" {
		return errors.New("youtube: RateVideo requires a rating")
	}
	q := url.Values{}
	q.Set("id", videoID)
	q.Set("rating", rating)
	return c.doPostNoContent(ctx, "videos/rate", q)
}

// GetVideoRating retrieves the authenticated user's rating for one or more
// videos. Mirrors videos.get_rating. Requires OAuth.
func (c *Client) GetVideoRating(ctx context.Context, videoIDs []string) (*VideoGetRatingResponse, error) {
	if len(videoIDs) == 0 {
		return nil, errors.New("youtube: GetVideoRating requires at least one video ID")
	}
	q := url.Values{}
	q.Set("id", strings.Join(videoIDs, ","))

	out := &VideoGetRatingResponse{}
	if err := c.doGet(ctx, "videos/getRating", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReportVideoAbuseParams configures ReportVideoAbuse.
type ReportVideoAbuseParams struct {
	// VideoID is the video to report. Required.
	VideoID string
	// ReasonID is the abuse reason ID. Required.
	ReasonID string
	// SecondaryReasonID is an optional secondary abuse reason ID.
	SecondaryReasonID string
	// Comments is optional additional information.
	Comments string
	// Language is the BCP-47 language of the report text.
	Language string
}

// reportAbuseBody is the JSON request body for videos.reportAbuse.
type reportAbuseBody struct {
	VideoID           string `json:"videoId"`
	ReasonID          string `json:"reasonId"`
	SecondaryReasonID string `json:"secondaryReasonId,omitempty"`
	Comments          string `json:"comments,omitempty"`
	Language          string `json:"language,omitempty"`
}

// ReportVideoAbuse reports a video for abusive content. Mirrors
// videos.report_abuse. Requires OAuth.
func (c *Client) ReportVideoAbuse(ctx context.Context, p ReportVideoAbuseParams) error {
	if p.VideoID == "" {
		return errors.New("youtube: ReportVideoAbuse requires a VideoID")
	}
	if p.ReasonID == "" {
		return errors.New("youtube: ReportVideoAbuse requires a ReasonID")
	}
	body := reportAbuseBody{
		VideoID:           p.VideoID,
		ReasonID:          p.ReasonID,
		SecondaryReasonID: p.SecondaryReasonID,
		Comments:          p.Comments,
		Language:          p.Language,
	}
	return c.doJSON(ctx, http.MethodPost, "videos/reportAbuse", nil, body, nil)
}
