package youtube

import (
	"context"
	"errors"
	"net/url"
	"os"
)

// ChannelBannerResource is the response from InsertChannelBanner. Its URL must
// be used in a subsequent UpdateChannel call to set
// brandingSettings.image.bannerExternalUrl.
type ChannelBannerResource struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`
	URL  string `json:"url,omitempty"`
}

// InsertChannelBannerParams configures InsertChannelBanner.
type InsertChannelBannerParams struct {
	// ImageData is the raw image bytes (JPEG or PNG, max 6 MB, 16:9 aspect
	// ratio, recommended 2560x1440 px). Required.
	ImageData []byte
	// ContentType is the image MIME type. Defaults to "image/jpeg".
	ContentType string
}

// InsertChannelBanner uploads a channel banner image via a direct binary media
// upload. Mirrors channelBanners.insert. Requires OAuth.
func (c *Client) InsertChannelBanner(ctx context.Context, p InsertChannelBannerParams) (*ChannelBannerResource, error) {
	if err := c.requireOAuth(); err != nil {
		return nil, err
	}
	if len(p.ImageData) == 0 {
		return nil, errors.New("youtube: InsertChannelBanner requires ImageData")
	}
	contentType := p.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}
	q := url.Values{}
	q.Set("uploadType", "media")

	out := &ChannelBannerResource{}
	if err := c.uploadMedia(ctx, "channelBanners/insert", q, p.ImageData, contentType, out); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertChannelBannerFromFile uploads a channel banner from a local file.
// Mirrors channelBanners.insert_from_file. Requires OAuth.
func (c *Client) InsertChannelBannerFromFile(ctx context.Context, filePath, contentType string) (*ChannelBannerResource, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("youtube: InsertChannelBannerFromFile: read file: " + err.Error())
	}
	return c.InsertChannelBanner(ctx, InsertChannelBannerParams{
		ImageData:   data,
		ContentType: contentType,
	})
}
