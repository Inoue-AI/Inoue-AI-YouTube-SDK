package youtube

// This file defines the shared resource types reused across multiple YouTube
// resources, mirroring youtube/models/common.py in the Python SDK. Field names
// and JSON tags match the YouTube Data API wire format exactly.

// PrivacyStatus is the visibility setting for videos, playlists, and other
// resources.
type PrivacyStatus = string

// Privacy status constants accepted by the YouTube Data API.
const (
	PrivacyPublic   PrivacyStatus = "public"
	PrivacyUnlisted PrivacyStatus = "unlisted"
	PrivacyPrivate  PrivacyStatus = "private"
)

// Thumbnail is a single thumbnail image rendition.
type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// ThumbnailDetails is the collection of thumbnail images at different
// resolutions. The Default field maps to the "default" key in the API response.
type ThumbnailDetails struct {
	Default  *Thumbnail `json:"default,omitempty"`
	Medium   *Thumbnail `json:"medium,omitempty"`
	High     *Thumbnail `json:"high,omitempty"`
	Standard *Thumbnail `json:"standard,omitempty"`
	Maxres   *Thumbnail `json:"maxres,omitempty"`
}

// PageInfo is the pagination metadata returned by every list endpoint.
type PageInfo struct {
	TotalResults   int `json:"totalResults,omitempty"`
	ResultsPerPage int `json:"resultsPerPage,omitempty"`
}

// ResourceID identifies a specific YouTube resource (video, channel, or
// playlist). It is used both in playlist items / subscriptions snippets and as
// the id field of a search result.
type ResourceID struct {
	Kind       string `json:"kind,omitempty"`
	VideoID    string `json:"videoId,omitempty"`
	ChannelID  string `json:"channelId,omitempty"`
	PlaylistID string `json:"playlistId,omitempty"`
}

// Localization is a localized title and description for a resource.
type Localization struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
