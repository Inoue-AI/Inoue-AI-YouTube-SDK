"""Tests for the YouTubeClient initialization and lifecycle."""

from __future__ import annotations

import pytest

from youtube.client import YouTubeClient
from youtube.exceptions import YouTubeConfigError


class TestClientInit:
    """Tests for client initialization."""

    def test_requires_credentials(self) -> None:
        with pytest.raises(YouTubeConfigError, match="access_token or api_key"):
            YouTubeClient()

    def test_access_token_only(self) -> None:
        client = YouTubeClient(access_token="token123")
        assert client.videos is not None
        assert client.channels is not None
        assert client.playlists is not None

    def test_api_key_only(self) -> None:
        client = YouTubeClient(api_key="key123")
        assert client.search is not None

    def test_both_credentials(self) -> None:
        client = YouTubeClient(access_token="token", api_key="key")
        assert client.videos is not None

    def test_all_api_namespaces_exist(self) -> None:
        client = YouTubeClient(access_token="token")
        assert hasattr(client, "videos")
        assert hasattr(client, "channels")
        assert hasattr(client, "playlists")
        assert hasattr(client, "playlist_items")
        assert hasattr(client, "comments")
        assert hasattr(client, "comment_threads")
        assert hasattr(client, "subscriptions")
        assert hasattr(client, "thumbnails")
        assert hasattr(client, "search")
        assert hasattr(client, "captions")
        assert hasattr(client, "channel_sections")
        assert hasattr(client, "channel_banners")
        assert hasattr(client, "watermarks")
        assert hasattr(client, "analytics")

    async def test_context_manager(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            assert client.videos is not None
