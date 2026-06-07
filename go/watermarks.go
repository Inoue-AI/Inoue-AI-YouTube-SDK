package youtube

import (
	"context"
	"errors"
	"net/url"
)

// Watermark timing type constants.
const (
	WatermarkTimingOffsetFromStart = "offsetFromStart"
	WatermarkTimingOffsetFromEnd   = "offsetFromEnd"
)

// Watermark corner position constants.
const (
	WatermarkCornerTopLeft     = "topLeft"
	WatermarkCornerTopRight    = "topRight"
	WatermarkCornerBottomLeft  = "bottomLeft"
	WatermarkCornerBottomRight = "bottomRight"
)

// WatermarkTiming is the timing configuration for a watermark.
type WatermarkTiming struct {
	Type       string `json:"type,omitempty"`
	OffsetMs   string `json:"offsetMs,omitempty"`
	DurationMs string `json:"durationMs,omitempty"`
}

// WatermarkPosition is the position configuration for a watermark.
type WatermarkPosition struct {
	Type           string `json:"type,omitempty"`
	CornerPosition string `json:"cornerPosition,omitempty"`
}

// Watermark is a YouTube Watermark resource (configuration for SetWatermark).
type Watermark struct {
	Timing          *WatermarkTiming   `json:"timing,omitempty"`
	Position        *WatermarkPosition `json:"position,omitempty"`
	ImageURL        string             `json:"imageUrl,omitempty"`
	ImageBytes      string             `json:"imageBytes,omitempty"`
	TargetChannelID string             `json:"targetChannelId,omitempty"`
}

// SetWatermarkParams configures SetWatermark.
type SetWatermarkParams struct {
	// ChannelID is the channel to set the watermark on. Required.
	ChannelID string
	// Watermark is the watermark configuration (timing, position). Mirrors the
	// Python SDK, which transmits the image via the binary media upload; the
	// configuration is accepted for parity.
	Watermark Watermark
	// ImageData is the raw image bytes (JPEG or PNG, max 10 MB). Required.
	ImageData []byte
	// ContentType is the image MIME type. Defaults to "image/png".
	ContentType string
}

// SetWatermark uploads and sets a watermark for a channel via a binary media
// upload. Mirrors watermarks.set. Requires OAuth.
func (c *Client) SetWatermark(ctx context.Context, p SetWatermarkParams) error {
	if err := c.requireOAuth(); err != nil {
		return err
	}
	if p.ChannelID == "" {
		return errors.New("youtube: SetWatermark requires a ChannelID")
	}
	if len(p.ImageData) == 0 {
		return errors.New("youtube: SetWatermark requires ImageData")
	}
	contentType := p.ContentType
	if contentType == "" {
		contentType = "image/png"
	}
	q := url.Values{}
	q.Set("channelId", p.ChannelID)
	q.Set("uploadType", "media")

	return c.uploadMedia(ctx, "watermarks/set", q, p.ImageData, contentType, nil)
}

// UnsetWatermark removes the watermark from a channel. Mirrors
// watermarks.unset. Requires OAuth.
func (c *Client) UnsetWatermark(ctx context.Context, channelID string) error {
	if channelID == "" {
		return errors.New("youtube: UnsetWatermark requires a ChannelID")
	}
	q := url.Values{}
	q.Set("channelId", channelID)
	return c.doPostNoContent(ctx, "watermarks/unset", q)
}
