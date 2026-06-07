# YouTube Go SDK

Typed, context-aware Go client for the YouTube Data API v3 and the YouTube
Analytics API v2, at parity with the **core operations** of the Python SDK:
OAuth token refresh, the main reads, and the publish/manage surface
(resumable video upload, update, delete, thumbnail set).

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

### Reads (`*Client`)

| Go method | YouTube endpoint | Auth |
|---|---|---|
| `GetChannel(ctx, params)` | `GET /channels` | OAuth or API key |
| `ListChannelVideos(ctx, params)` | `GET /playlistItems?playlistId=<uploads>` | OAuth or API key |
| `GetVideo(ctx, id, parts)` | `GET /videos` | OAuth or API key |
| `Search(ctx, params)` | `GET /search` | OAuth or API key |
| `AnalyticsQuery(ctx, params)` | `GET /reports` (Analytics API) | OAuth only |

### Publish / manage (`*Client`, OAuth only)

| Go method | YouTube endpoint |
|---|---|
| `InsertVideo(ctx, params)` | `POST /upload/videos` (resumable, chunked) |
| `UpdateVideo(ctx, metadata, parts)` | `PUT /videos` |
| `DeleteVideo(ctx, videoID)` | `DELETE /videos` |
| `SetThumbnail(ctx, params)` | `POST /upload/thumbnails/set` (direct media) |

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
