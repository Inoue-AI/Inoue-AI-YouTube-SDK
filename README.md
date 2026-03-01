# YouTube Python SDK

Async Python SDK for the **YouTube Data API v3** and **YouTube Analytics API v2**.

Built with `aiohttp` and `pydantic` v2. Follows the same architecture as the [TikTok Python SDK](https://github.com/inoue-ai/tiktok-python-sdk).

## Features

- **14 API namespaces** — Videos, Channels, Playlists, PlaylistItems, Comments, CommentThreads, Subscriptions, Thumbnails, Search, Captions, ChannelSections, ChannelBanners, Watermarks, Analytics
- **Pydantic v2 models** for every request and response — full type safety, IDE autocomplete, and serialization
- **Async-first** — powered by `aiohttp` with lazy session creation
- **Resumable video uploads** — handles chunked uploads automatically
- **Auto-pagination** — async iterators for paging through large result sets
- **Dual auth** — OAuth 2.0 access tokens (full access) or API keys (public read-only)
- **Comprehensive error hierarchy** — typed exceptions mapped from YouTube error codes

## Installation

```bash
pip install youtube-python-sdk
```

Or install from source:

```bash
git clone https://github.com/inoue-ai/youtube-python-sdk.git
cd youtube-python-sdk
pip install -e ".[dev]"
```

## Quick Start

### Read public data with an API key

```python
import asyncio
from youtube import YouTubeClient, VideoPart

async def main():
    async with YouTubeClient(api_key="YOUR_API_KEY") as yt:
        response = await yt.videos.list(
            parts=[VideoPart.SNIPPET, VideoPart.STATISTICS],
            chart="mostPopular",
            max_results=10,
        )
        for video in response.items:
            print(f"{video.snippet.title} — {video.statistics.view_count} views")

asyncio.run(main())
```

### Manage your channel with OAuth

```python
import asyncio
from youtube import (
    YouTubeClient, ChannelPart, Video, VideoSnippet, VideoStatus,
    PrivacyStatus, VideoPart,
)

async def main():
    async with YouTubeClient(access_token="OAUTH_ACCESS_TOKEN") as yt:
        # Get your channel info
        channels = await yt.channels.list(
            parts=[ChannelPart.SNIPPET, ChannelPart.STATISTICS],
            mine=True,
        )
        channel = channels.items[0]
        print(f"Channel: {channel.snippet.title}")
        print(f"Subscribers: {channel.statistics.subscriber_count}")

        # Upload a video
        video = await yt.videos.insert(
            body=Video(
                snippet=VideoSnippet(
                    title="My New Video",
                    description="Uploaded via the YouTube Python SDK!",
                    tags=["sdk", "python", "youtube"],
                    category_id="22",
                ),
                status=VideoStatus(
                    privacy_status=PrivacyStatus.PRIVATE,
                ),
            ),
            file_path="/path/to/video.mp4",
        )
        print(f"Uploaded: {video.id}")

        # Set a custom thumbnail
        await yt.thumbnails.set_from_file(
            video_id=video.id,
            file_path="/path/to/thumbnail.jpg",
        )

asyncio.run(main())
```

### Search

```python
from youtube import YouTubeClient, SearchResultType, SearchOrder

async with YouTubeClient(api_key="KEY") as yt:
    results = await yt.search.list(
        q="python tutorial",
        type=SearchResultType.VIDEO,
        order=SearchOrder.VIEW_COUNT,
        max_results=25,
    )
    for item in results.items:
        print(f"{item.snippet.title} — {item.id.video_id}")
```

### Playlists

```python
from youtube import (
    YouTubeClient, Playlist, PlaylistSnippet, PlaylistStatus,
    PlaylistPart, PlaylistItem, PlaylistItemSnippet, PlaylistItemPart,
    PrivacyStatus, ResourceId,
)

async with YouTubeClient(access_token="TOKEN") as yt:
    # Create a playlist
    playlist = await yt.playlists.insert(
        body=Playlist(
            snippet=PlaylistSnippet(title="My SDK Playlist"),
            status=PlaylistStatus(privacy_status=PrivacyStatus.PUBLIC),
        ),
    )

    # Add a video to the playlist
    await yt.playlist_items.insert(
        body=PlaylistItem(
            snippet=PlaylistItemSnippet(
                playlist_id=playlist.id,
                resource_id=ResourceId(
                    kind="youtube#video",
                    video_id="dQw4w9WgXcQ",
                ),
            ),
        ),
    )
```

### Comments

```python
from youtube import (
    YouTubeClient, CommentThread, CommentThreadSnippet,
    CommentThreadPart, Comment, CommentSnippet, CommentPart,
)

async with YouTubeClient(access_token="TOKEN") as yt:
    # Post a top-level comment
    thread = await yt.comment_threads.insert(
        body=CommentThread(
            snippet=CommentThreadSnippet(
                video_id="VIDEO_ID",
                channel_id="CHANNEL_ID",
                top_level_comment=Comment(
                    snippet=CommentSnippet(text_original="Great video!"),
                ),
            ),
        ),
    )

    # List comments on a video
    async for thread in yt.comment_threads.iter_video_threads(
        "VIDEO_ID",
        parts=[CommentThreadPart.SNIPPET, CommentThreadPart.REPLIES],
    ):
        top = thread.snippet.top_level_comment.snippet
        print(f"{top.author_display_name}: {top.text_display}")
```

### Analytics

```python
from youtube import YouTubeClient

async with YouTubeClient(access_token="TOKEN") as yt:
    report = await yt.analytics.query(
        start_date="2025-01-01",
        end_date="2025-12-31",
        metrics=["views", "likes", "estimatedMinutesWatched"],
        dimensions=["day"],
    )
    for row in report.rows:
        print(row)  # [date, views, likes, minutes_watched]
```

### Captions

```python
from youtube import YouTubeClient, Caption, CaptionSnippet, CaptionPart, CaptionFormat

async with YouTubeClient(access_token="TOKEN") as yt:
    # List captions for a video
    captions = await yt.captions.list(
        parts=[CaptionPart.SNIPPET],
        video_id="VIDEO_ID",
    )

    # Download a caption track as SRT
    srt_data = await yt.captions.download(
        id=captions.items[0].id,
        tfmt=CaptionFormat.SRT,
    )

    # Upload a new caption track
    await yt.captions.insert_from_file(
        body=Caption(
            snippet=CaptionSnippet(
                video_id="VIDEO_ID",
                language="en",
                name="English",
            ),
        ),
        file_path="/path/to/captions.srt",
    )
```

## API Namespaces

| Namespace | Methods | Description |
|-----------|---------|-------------|
| `yt.videos` | list, insert, update, delete, rate, get_rating, report_abuse | Video management & upload |
| `yt.channels` | list, update | Channel info & branding |
| `yt.playlists` | list, insert, update, delete | Playlist CRUD |
| `yt.playlist_items` | list, insert, update, delete | Manage videos in playlists |
| `yt.comments` | list, insert, update, delete, set_moderation_status | Comment replies & moderation |
| `yt.comment_threads` | list, insert | Top-level comments |
| `yt.subscriptions` | list, insert, delete | Subscription management |
| `yt.thumbnails` | set, set_from_file | Custom thumbnail upload |
| `yt.search` | list | Search across YouTube |
| `yt.captions` | list, insert, insert_from_file, update, download, delete | Caption track management |
| `yt.channel_sections` | list, insert, update, delete | Channel page layout |
| `yt.channel_banners` | insert, insert_from_file | Banner image upload |
| `yt.watermarks` | set, unset | Channel watermark |
| `yt.analytics` | query | YouTube Analytics reports |

## Authentication

The SDK does **not** handle OAuth flows. You must provide a pre-obtained token:

- **`access_token`** — OAuth 2.0 token for full read/write access
- **`api_key`** — Google API key for read-only public data

```python
# OAuth (full access)
client = YouTubeClient(access_token="ya29....")

# API key (public read-only)
client = YouTubeClient(api_key="AIzaSy...")

# Both (OAuth takes priority, API key added to all requests)
client = YouTubeClient(access_token="ya29....", api_key="AIzaSy...")
```

## Development

```bash
# Install dev dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Lint & format
ruff check .
ruff format .

# Type check
mypy youtube/
```

## License

MIT License — see [LICENSE](LICENSE) for details.
