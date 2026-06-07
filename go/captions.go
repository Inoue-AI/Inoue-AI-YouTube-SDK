package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Caption parts accepted by the part parameter (captions.list / insert).
const (
	CaptionPartID      = "id"
	CaptionPartSnippet = "snippet"
)

// Caption track kinds.
const (
	TrackKindASR      = "ASR"
	TrackKindForced   = "forced"
	TrackKindStandard = "standard"
)

// Caption download formats accepted by the tfmt parameter.
const (
	CaptionFormatSBV  = "sbv"
	CaptionFormatSCC  = "scc"
	CaptionFormatSRT  = "srt"
	CaptionFormatTTML = "ttml"
	CaptionFormatVTT  = "vtt"
)

// CaptionSnippet is the snippet part of a Caption resource. It is used both as
// a request body (insert/update) and a response projection.
type CaptionSnippet struct {
	VideoID        string `json:"videoId,omitempty"`
	LastUpdated    string `json:"lastUpdated,omitempty"`
	TrackKind      string `json:"trackKind,omitempty"`
	Language       string `json:"language,omitempty"`
	Name           string `json:"name,omitempty"`
	AudioTrackType string `json:"audioTrackType,omitempty"`
	IsCC           *bool  `json:"isCC,omitempty"`
	IsDraft        *bool  `json:"isDraft,omitempty"`
	IsAutoSynced   *bool  `json:"isAutoSynced,omitempty"`
	Status         string `json:"status,omitempty"`
	FailureReason  string `json:"failureReason,omitempty"`
}

// Caption is a YouTube Caption resource (request body and response).
type Caption struct {
	Kind    string          `json:"kind,omitempty"`
	Etag    string          `json:"etag,omitempty"`
	ID      string          `json:"id,omitempty"`
	Snippet *CaptionSnippet `json:"snippet,omitempty"`
}

// CaptionListResponse is the response from ListCaptions.
type CaptionListResponse struct {
	Kind  string    `json:"kind,omitempty"`
	Etag  string    `json:"etag,omitempty"`
	Items []Caption `json:"items"`
}

// ListCaptionsParams configures ListCaptions.
type ListCaptionsParams struct {
	// Parts is the list of resource parts to include (e.g. ["snippet"]).
	// Required.
	Parts []string
	// VideoID is the video whose caption tracks are returned. Required.
	VideoID string
	// IDs optionally restricts the result to specific caption track IDs.
	IDs []string
}

// ListCaptions returns the caption tracks for a video. Mirrors
// captions.list.
func (c *Client) ListCaptions(ctx context.Context, p ListCaptionsParams) (*CaptionListResponse, error) {
	if len(p.Parts) == 0 {
		return nil, errors.New("youtube: ListCaptions requires at least one part")
	}
	if p.VideoID == "" {
		return nil, errors.New("youtube: ListCaptions requires VideoID")
	}
	q := url.Values{}
	q.Set("part", strings.Join(p.Parts, ","))
	q.Set("videoId", p.VideoID)
	if len(p.IDs) > 0 {
		q.Set("id", strings.Join(p.IDs, ","))
	}

	out := &CaptionListResponse{}
	if err := c.doGet(ctx, "captions", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertCaptionParams configures InsertCaption.
type InsertCaptionParams struct {
	// Body is the caption metadata. snippet.videoId, snippet.language, and
	// snippet.name are required.
	Body Caption
	// CaptionData is the raw caption file bytes. Required.
	CaptionData []byte
	// Parts overrides the resource parts. Defaults to ["snippet"].
	Parts []string
	// ContentType is the MIME type of the caption file. Defaults to
	// "application/octet-stream".
	ContentType string
}

// InsertCaption uploads a new caption track via a multipart upload (JSON
// metadata + binary file). Mirrors captions.insert. Requires OAuth.
func (c *Client) InsertCaption(ctx context.Context, p InsertCaptionParams) (*Caption, error) {
	if len(p.CaptionData) == 0 {
		return nil, errors.New("youtube: InsertCaption requires CaptionData")
	}
	parts := p.Parts
	if len(parts) == 0 {
		parts = []string{CaptionPartSnippet}
	}
	contentType := p.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Caption{}
	if err := c.uploadMultipart(ctx, "captions", q, p.Body, p.CaptionData, contentType, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertCaptionFromFile uploads a new caption track from a local file. Mirrors
// captions.insert_from_file. Requires OAuth.
func (c *Client) InsertCaptionFromFile(
	ctx context.Context,
	body Caption,
	filePath string,
	parts []string,
	contentType string,
) (*Caption, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("youtube: InsertCaptionFromFile: read file: " + err.Error())
	}
	return c.InsertCaption(ctx, InsertCaptionParams{
		Body:        body,
		CaptionData: data,
		Parts:       parts,
		ContentType: contentType,
	})
}

// UpdateCaptionParams configures UpdateCaption.
type UpdateCaptionParams struct {
	// Body is the caption metadata. id is required.
	Body Caption
	// CaptionData, when non-nil, replaces the caption file (multipart upload).
	// When nil, only the metadata is updated (JSON PUT).
	CaptionData []byte
	// Parts overrides the resource parts. Defaults to ["snippet"].
	Parts []string
	// ContentType is the MIME type used when CaptionData is provided. Defaults
	// to "application/octet-stream".
	ContentType string
}

// UpdateCaption updates an existing caption track. If CaptionData is provided,
// the caption file is replaced via a multipart upload; otherwise only metadata
// is updated via a JSON PUT. Mirrors captions.update. Requires OAuth.
func (c *Client) UpdateCaption(ctx context.Context, p UpdateCaptionParams) (*Caption, error) {
	parts := p.Parts
	if len(parts) == 0 {
		parts = []string{CaptionPartSnippet}
	}
	q := url.Values{}
	q.Set("part", strings.Join(parts, ","))

	out := &Caption{}
	if p.CaptionData != nil {
		contentType := p.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if err := c.uploadMultipart(ctx, "captions", q, p.Body, p.CaptionData, contentType, out); err != nil {
			return nil, err
		}
		return out, nil
	}
	if err := c.doJSON(ctx, http.MethodPut, "captions", q, p.Body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DownloadCaptionParams configures DownloadCaption.
type DownloadCaptionParams struct {
	// ID is the caption track ID. Required.
	ID string
	// Tfmt is the output format (e.g. "srt", "vtt"). If empty, the original
	// format is returned.
	Tfmt string
	// Tlang is a BCP-47 language code to translate the captions into.
	Tlang string
}

// DownloadCaption downloads a caption track as raw bytes. Mirrors
// captions.download.
func (c *Client) DownloadCaption(ctx context.Context, p DownloadCaptionParams) ([]byte, error) {
	if p.ID == "" {
		return nil, errors.New("youtube: DownloadCaption requires an ID")
	}
	q := url.Values{}
	if p.Tfmt != "" {
		q.Set("tfmt", p.Tfmt)
	}
	if p.Tlang != "" {
		q.Set("tlang", p.Tlang)
	}
	return c.doGetBinary(ctx, "captions/"+p.ID, q)
}

// DeleteCaption deletes a caption track by ID. Mirrors captions.delete.
// Requires OAuth.
func (c *Client) DeleteCaption(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("youtube: DeleteCaption requires an ID")
	}
	q := url.Values{}
	q.Set("id", id)
	return c.doDelete(ctx, "captions", q)
}
