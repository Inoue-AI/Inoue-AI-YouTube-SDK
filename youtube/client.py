"""Main entry point for the YouTube Python SDK.

Usage — async context manager (recommended)::

    async with YouTubeClient(access_token="OAUTH_TOKEN") as yt:
        response = await yt.videos.list(
            parts=[VideoPart.SNIPPET, VideoPart.STATISTICS],
            chart="mostPopular",
            max_results=10,
        )
        for video in response.items:
            print(video.snippet.title, video.statistics.view_count)

Usage — API key for public read-only access::

    async with YouTubeClient(api_key="YOUR_API_KEY") as yt:
        results = await yt.search.list(q="python tutorial", max_results=5)

Usage — manual lifecycle management::

    client = YouTubeClient(access_token="OAUTH_TOKEN")
    try:
        channel = await client.channels.list(parts=[ChannelPart.SNIPPET], mine=True)
    finally:
        await client.aclose()
"""

from __future__ import annotations

from types import TracebackType
from typing import Self

from youtube.apis.analytics import AnalyticsAPI
from youtube.apis.captions import CaptionsAPI
from youtube.apis.channel_banners import ChannelBannersAPI
from youtube.apis.channel_sections import ChannelSectionsAPI
from youtube.apis.channels import ChannelsAPI
from youtube.apis.comment_threads import CommentThreadsAPI
from youtube.apis.comments import CommentsAPI
from youtube.apis.playlist_items import PlaylistItemsAPI
from youtube.apis.playlists import PlaylistsAPI
from youtube.apis.search import SearchAPI
from youtube.apis.subscriptions import SubscriptionsAPI
from youtube.apis.thumbnails import ThumbnailsAPI
from youtube.apis.videos import VideosAPI
from youtube.apis.watermarks import WatermarksAPI
from youtube.config import DEFAULT_TIMEOUT
from youtube.exceptions import YouTubeConfigError
from youtube.http.session import YouTubeSession


class YouTubeClient:
    """Async YouTube API client.

    Provide either an OAuth 2.0 ``access_token`` (for full access) or a
    Google ``api_key`` (for read-only public data).  The SDK does **not**
    handle OAuth flows; token management must be done externally.

    Parameters
    ----------
    access_token:
        OAuth 2.0 access token (required for write operations).
    api_key:
        Google API key (sufficient for read-only public endpoints).
    timeout:
        Total HTTP request timeout in seconds (default: ``30.0``).

    Attributes
    ----------
    videos:           :class:`~youtube.apis.videos.VideosAPI`
    channels:         :class:`~youtube.apis.channels.ChannelsAPI`
    playlists:        :class:`~youtube.apis.playlists.PlaylistsAPI`
    playlist_items:   :class:`~youtube.apis.playlist_items.PlaylistItemsAPI`
    comments:         :class:`~youtube.apis.comments.CommentsAPI`
    comment_threads:  :class:`~youtube.apis.comment_threads.CommentThreadsAPI`
    subscriptions:    :class:`~youtube.apis.subscriptions.SubscriptionsAPI`
    thumbnails:       :class:`~youtube.apis.thumbnails.ThumbnailsAPI`
    search:           :class:`~youtube.apis.search.SearchAPI`
    captions:         :class:`~youtube.apis.captions.CaptionsAPI`
    channel_sections: :class:`~youtube.apis.channel_sections.ChannelSectionsAPI`
    channel_banners:  :class:`~youtube.apis.channel_banners.ChannelBannersAPI`
    watermarks:       :class:`~youtube.apis.watermarks.WatermarksAPI`
    analytics:        :class:`~youtube.apis.analytics.AnalyticsAPI`
    """

    def __init__(
        self,
        *,
        access_token: str | None = None,
        api_key: str | None = None,
        timeout: float = DEFAULT_TIMEOUT,
    ) -> None:
        if not access_token and not api_key:
            raise YouTubeConfigError(
                "At least one of access_token or api_key must be provided."
            )

        self._session = YouTubeSession(
            access_token=access_token,
            api_key=api_key,
            timeout=timeout,
        )

        # API namespaces
        self.videos = VideosAPI(self._session)
        self.channels = ChannelsAPI(self._session)
        self.playlists = PlaylistsAPI(self._session)
        self.playlist_items = PlaylistItemsAPI(self._session)
        self.comments = CommentsAPI(self._session)
        self.comment_threads = CommentThreadsAPI(self._session)
        self.subscriptions = SubscriptionsAPI(self._session)
        self.thumbnails = ThumbnailsAPI(self._session)
        self.search = SearchAPI(self._session)
        self.captions = CaptionsAPI(self._session)
        self.channel_sections = ChannelSectionsAPI(self._session)
        self.channel_banners = ChannelBannersAPI(self._session)
        self.watermarks = WatermarksAPI(self._session)
        self.analytics = AnalyticsAPI(self._session)

    # Context manager support -----------------------------------------------

    async def __aenter__(self) -> Self:
        return self

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> None:
        await self.aclose()

    # Lifecycle -------------------------------------------------------------

    async def aclose(self) -> None:
        """Close the underlying HTTP session and release all connections."""
        await self._session.aclose()
