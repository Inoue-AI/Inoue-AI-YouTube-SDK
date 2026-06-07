# YouTube Go SDK

Typed, context-aware Go client for the YouTube Data API v3 and the YouTube
Analytics API v2, at **full parity** with the Python SDK: OAuth token refresh,
every read, the complete write/manage surface (resumable video upload, all CRUD
endpoints, captions multipart upload, banner/thumbnail/watermark binary uploads,
comment moderation, ratings, abuse reports), and cursor-walking iterators.

Operations are flat methods on `*Client`, named `<Resource><Verb>` (e.g.
`ListPlaylists`, `InsertCaption`, `RateVideo`).

## Install

```bash
go get github.com/Inoue-AI/Inoue-AI-YouTube-SDK/go@latest
```

## Quickstart

```go
package main

import (
	"context"
	"log"
	"time"

	youtube "github.com/Inoue-AI/Inoue-AI-YouTube-SDK/go"
)

func main() {
	client := youtube.New(youtube.ClientOptions{
		AccessToken: "OAUTH_TOKEN",
		Timeout:     30 * time.Second,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	channel, err := client.GetChannel(ctx, youtube.GetChannelParams{Mine: true})
	if err != nil {
		log.Fatalf("get channel: %v", err)
	}
	log.Printf("channel %s subscribers=%s", channel.ID, channel.Statistics.SubscriberCount)
}
```

## Methods

### OAuth (`*OAuthClient`)

| Go method | OAuth grant |
|---|---|
| `RefreshAccessToken(ctx, refreshToken)` | `grant_type=refresh_token` |
| `ExchangeAuthorizationCode(ctx, code, redirectURI, codeVerifier)` | `grant_type=authorization_code` (+ optional PKCE) |

```go
oauth, _ := youtube.NewOAuthClient(youtube.OAuthOptions{
    ClientID:     os.Getenv("YT_OAUTH_CLIENT_ID"),
    ClientSecret: os.Getenv("YT_OAUTH_CLIENT_SECRET"),
})
defer oauth.Close()
tok, err := oauth.RefreshAccessToken(ctx, os.Getenv("YT_REFRESH_TOKEN"))
// tok.AccessToken -> youtube.New(youtube.ClientOptions{AccessToken: tok.AccessToken})
```

### Reads (`*Client`, OAuth or API key)

| Go method | YouTube endpoint |
|---|---|
| `ListChannels(ctx, params)` / `GetChannel(ctx, params)` | `GET /channels` |
| `ListChannelVideos(ctx, params)` | `GET /playlistItems?playlistId=<uploads>` |
| `ListChannelSections(ctx, params)` | `GET /channelSections` |
| `ListVideos(ctx, params)` / `GetVideo(ctx, id, parts)` | `GET /videos` |
| `ListPlaylists(ctx, params)` | `GET /playlists` |
| `ListPlaylistItems(ctx, params)` | `GET /playlistItems` |
| `ListComments(ctx, params)` | `GET /comments` |
| `ListCommentThreads(ctx, params)` | `GET /commentThreads` |
| `ListSubscriptions(ctx, params)` | `GET /subscriptions` |
| `ListCaptions(ctx, params)` | `GET /captions` |
| `DownloadCaption(ctx, params)` | `GET /captions/{id}` (binary) |
| `GetVideoRating(ctx, ids)` | `GET /videos/getRating` |
| `ListSearch(ctx, params)` / `Search(ctx, params)` | `GET /search` |
| `AnalyticsQuery(ctx, params)` | `GET /reports` (Analytics API; OAuth only) |

Iterators (`IterVideosByChart`, `IterMyPlaylists`, `IterPlaylistItems`,
`IterReplies`, `IterVideoThreads`, `IterMySubscriptions`, `IterSearchResults`)
walk every page via a caller-supplied callback.

### Write / manage (`*Client`, OAuth only)

| Go method | YouTube endpoint | Protocol |
|---|---|---|
| `InsertVideo(ctx, params)` | `POST /upload/videos` | resumable, chunked |
| `UpdateVideo` / `DeleteVideo` | `PUT` / `DELETE /videos` | JSON / query |
| `RateVideo` / `ReportVideoAbuse` | `POST /videos/rate`, `/videos/reportAbuse` | query / JSON |
| `UpdateChannel(ctx, body, parts)` | `PUT /channels` | JSON |
| `Insert/Update/DeleteChannelSection` | `POST/PUT/DELETE /channelSections` | JSON / query |
| `InsertChannelBanner(ctx, params)` | `POST /upload/channelBanners/insert` | binary media |
| `Insert/Update/DeletePlaylist` | `POST/PUT/DELETE /playlists` | JSON / query |
| `Insert/Update/DeletePlaylistItem` | `POST/PUT/DELETE /playlistItems` | JSON / query |
| `InsertComment` / `UpdateComment` / `DeleteComment` | `POST/PUT/DELETE /comments` | JSON / query |
| `SetCommentModerationStatus` | `POST /comments/setModerationStatus` | query |
| `InsertCommentThread` | `POST /commentThreads` | JSON |
| `InsertSubscription` / `DeleteSubscription` | `POST/DELETE /subscriptions` | JSON / query |
| `InsertCaption` / `UpdateCaption` | `POST/PUT /upload/captions` | multipart / JSON |
| `DeleteCaption` | `DELETE /captions` | query |
| `SetThumbnail(ctx, params)` | `POST /upload/thumbnails/set` | binary media |
| `SetWatermark` / `UnsetWatermark` | `POST /upload/watermarks/set`, `/watermarks/unset` | binary / query |

For unauthenticated access to public data, supply `APIKey` instead of
`AccessToken` in `ClientOptions`. Write, upload, and Analytics endpoints
reject API-key-only auth and require an `AccessToken`.

### Resumable upload

`InsertVideo` implements the two-phase resumable protocol: a session-init
`POST` (advertising `X-Upload-Content-Length`), followed by chunked `PUT`s that
honor the `308 Resume Incomplete` response and stream from any `io.Reader`.
`MediaSize` must be supplied so the total length is known up front; `ChunkSize`
defaults to 8 MiB.

## Operating principles

- Every method takes `context.Context` first.
- Each `*Client` owns one `*http.Client` with explicit `Timeout`,
  `MaxIdleConnsPerHost`, and `IdleConnTimeout`. `http.DefaultClient` is never
  used.
- `defer client.Close()` releases idle connections.
- Errors surface as `*youtube.Error` with `StatusCode`, `Status`, `Message`,
  and `Reasons` slice (mapped from `error.errors[*].reason`).

## Repository layout

The Go SDK lives in the `go/` subdirectory. The Python SDK remains under
`youtube/` and is unchanged. See the top-level [README](../README.md) for the
multi-language overview.
