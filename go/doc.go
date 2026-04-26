// Package youtube provides a typed, context-aware Go client for the YouTube
// Data API v3 (read paths only).
//
// This SDK exposes only the methods the Inoue AI Go backend calls today:
// GetChannel, ListChannelVideos, and GetVideo. Future Phase 3 work will
// expand the surface area as needed.
//
//	client := youtube.New(youtube.ClientOptions{
//	    AccessToken: "OAUTH_TOKEN",
//	    Timeout:     30 * time.Second,
//	})
//	defer client.Close()
//
//	channel, err := client.GetChannel(ctx, youtube.GetChannelParams{Mine: true})
package youtube
