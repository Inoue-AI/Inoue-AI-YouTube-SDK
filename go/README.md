# YouTube Go SDK

Typed, context-aware Go client for the YouTube Data API v3 (read paths only,
focused subset).

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

| Go method | YouTube endpoint |
|---|---|
| `GetChannel(ctx, params)` | `GET /channels` |
| `ListChannelVideos(ctx, params)` | `GET /playlistItems?playlistId=<uploads>` |
| `GetVideo(ctx, id, parts)` | `GET /videos` |

For unauthenticated access to public data, supply `APIKey` instead of
`AccessToken` in `ClientOptions`.

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
