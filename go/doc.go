// Package youtube provides a typed, context-aware Go client for the YouTube
// Data API v3 and the YouTube Analytics API v2, at parity with the core
// operations of the Inoue AI Python SDK.
//
// Core operations:
//
//   - OAuth:      OAuthClient.RefreshAccessToken, ExchangeAuthorizationCode
//   - Reads:      GetChannel, ListChannelVideos, GetVideo, Search,
//     AnalyticsQuery
//   - Publish:    InsertVideo (resumable upload), UpdateVideo, DeleteVideo,
//     SetThumbnail
//
// The client is safe for concurrent use. It never falls back to
// http.DefaultClient: each client owns an *http.Client with an explicit
// timeout and bounded idle connections. Callers must invoke Close (or use
// defer) so idle TCP connections are released.
//
//	client := youtube.New(youtube.ClientOptions{
//	    AccessToken: "OAUTH_TOKEN",
//	    Timeout:     30 * time.Second,
//	})
//	defer client.Close()
//
//	channel, err := client.GetChannel(ctx, youtube.GetChannelParams{Mine: true})
//
// Refreshing an expired access token before constructing the client:
//
//	oauth, _ := youtube.NewOAuthClient(youtube.OAuthOptions{
//	    ClientID:     os.Getenv("YT_OAUTH_CLIENT_ID"),
//	    ClientSecret: os.Getenv("YT_OAUTH_CLIENT_SECRET"),
//	})
//	defer oauth.Close()
//	tok, err := oauth.RefreshAccessToken(ctx, os.Getenv("YT_REFRESH_TOKEN"))
package youtube
