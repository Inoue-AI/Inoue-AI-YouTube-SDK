"""Tests for the Playlists API and models."""

from __future__ import annotations

import re

from aioresponses import aioresponses

from youtube.client import YouTubeClient
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.common import PrivacyStatus
from youtube.models.playlists import (
    Playlist,
    PlaylistListResponse,
    PlaylistPart,
    PlaylistSnippet,
    PlaylistStatus,
)

_BASE = YOUTUBE_API_BASE_URL
_PLAYLISTS_RE = re.compile(r"^https://www\.googleapis\.com/youtube/v3/playlists(\?.*)?$")

SAMPLE_PLAYLIST_LIST_RESPONSE = {
    "kind": "youtube#playlistListResponse",
    "etag": "abc123",
    "pageInfo": {"totalResults": 1, "resultsPerPage": 5},
    "items": [
        {
            "kind": "youtube#playlist",
            "etag": "def456",
            "id": "PLtest123",
            "snippet": {
                "publishedAt": "2024-01-01T00:00:00Z",
                "channelId": "UCtest",
                "title": "My Playlist",
                "description": "A test playlist",
                "channelTitle": "Test Channel",
            },
            "status": {
                "privacyStatus": "public",
            },
            "contentDetails": {
                "itemCount": 10,
            },
        }
    ],
}


class TestPlaylistModels:
    def test_playlist_from_api_response(self) -> None:
        playlist = Playlist.model_validate(SAMPLE_PLAYLIST_LIST_RESPONSE["items"][0])
        assert playlist.id == "PLtest123"
        assert playlist.snippet is not None
        assert playlist.snippet.title == "My Playlist"
        assert playlist.status is not None
        assert playlist.status.privacy_status == PrivacyStatus.PUBLIC
        assert playlist.content_details is not None
        assert playlist.content_details.item_count == 10

    def test_playlist_insert_serialization(self) -> None:
        playlist = Playlist(
            snippet=PlaylistSnippet(
                title="New Playlist",
                description="Created by SDK",
            ),
            status=PlaylistStatus(
                privacy_status=PrivacyStatus.PRIVATE,
            ),
        )
        dumped = playlist.model_dump(by_alias=True, exclude_none=True)
        assert dumped["snippet"]["title"] == "New Playlist"
        assert dumped["status"]["privacyStatus"] == "private"


class TestPlaylistsAPI:
    async def test_list_mine(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_PLAYLISTS_RE, payload=SAMPLE_PLAYLIST_LIST_RESPONSE)
                response = await client.playlists.list(
                    parts=[PlaylistPart.SNIPPET, PlaylistPart.STATUS],
                    mine=True,
                )
                assert isinstance(response, PlaylistListResponse)
                assert len(response.items) == 1

    async def test_insert(self) -> None:
        created_payload = {
            "kind": "youtube#playlist",
            "id": "PLnew123",
            "snippet": {"title": "New Playlist"},
            "status": {"privacyStatus": "private"},
        }
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.post(_PLAYLISTS_RE, payload=created_payload)
                body = Playlist(
                    snippet=PlaylistSnippet(title="New Playlist"),
                    status=PlaylistStatus(privacy_status=PrivacyStatus.PRIVATE),
                )
                result = await client.playlists.insert(body)
                assert result.id == "PLnew123"

    async def test_delete(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.delete(_PLAYLISTS_RE, status=204)
                await client.playlists.delete(id="PLtest123")
