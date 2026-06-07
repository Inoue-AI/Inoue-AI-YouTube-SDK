// Package youtube provides a typed, context-aware Go client for the YouTube
// Data API v3 and the YouTube Analytics API v2, at full parity with the Inoue
// AI Python SDK.
//
// Operations are exposed as flat methods on *Client, named <Resource><Verb>:
//
//   - OAuth:           OAuthClient.RefreshAccessToken, ExchangeAuthorizationCode
//   - Channels:        ListChannels, GetChannel, UpdateChannel, ListChannelVideos
//   - Channel sections: ListChannelSections, InsertChannelSection,
//     UpdateChannelSection, DeleteChannelSection
//   - Channel banners: InsertChannelBanner, InsertChannelBannerFromFile
//   - Videos:          ListVideos, GetVideo, InsertVideo (resumable upload),
//     UpdateVideo, DeleteVideo, RateVideo, GetVideoRating, ReportVideoAbuse,
//     IterVideosByChart
//   - Playlists:       ListPlaylists, InsertPlaylist, UpdatePlaylist,
//     DeletePlaylist, IterMyPlaylists
//   - Playlist items:  ListPlaylistItems, InsertPlaylistItem,
//     UpdatePlaylistItem, DeletePlaylistItem, IterPlaylistItems
//   - Comments:        ListComments, InsertComment, UpdateComment,
//     DeleteComment, SetCommentModerationStatus, IterReplies
//   - Comment threads: ListCommentThreads, InsertCommentThread, IterVideoThreads
//   - Subscriptions:   ListSubscriptions, InsertSubscription,
//     DeleteSubscription, IterMySubscriptions
//   - Captions:        ListCaptions, InsertCaption (multipart),
//     InsertCaptionFromFile, UpdateCaption, DownloadCaption, DeleteCaption
//   - Thumbnails:      SetThumbnail (binary upload)
//   - Watermarks:      SetWatermark (binary upload), UnsetWatermark
//   - Search:          ListSearch (full parameter set), Search (convenience),
//     IterSearchResults
//   - Analytics:       AnalyticsQuery
//
// Cursor-paginated endpoints expose both a single-page List method and an Iter
// method that walks every page via a caller-supplied callback — the
// Go-idiomatic equivalent of the Python async generators (iter_mine, etc.).
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
