"""Tests for the Videos API and models."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses
from pydantic import ValidationError

from youtube.client import YouTubeClient
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.common import PrivacyStatus
from youtube.models.videos import (
    Rating,
    UploadStatus,
    Video,
    VideoGetRatingResponse,
    VideoLicense,
    VideoListResponse,
    VideoPart,
    VideoSnippet,
    VideoStatus,
)

_BASE = YOUTUBE_API_BASE_URL
_VIDEOS_RE = re.compile(r"^https://www\.googleapis\.com/youtube/v3/videos(\?.*)?$")
_RATE_RE = re.compile(r"^https://www\.googleapis\.com/youtube/v3/videos/rate(\?.*)?$")
_GET_RATING_RE = re.compile(r"^https://www\.googleapis\.com/youtube/v3/videos/getRating(\?.*)?$")

# ---------------------------------------------------------------------------
# Sample payloads
# ---------------------------------------------------------------------------

SAMPLE_VIDEO_LIST_RESPONSE = {
    "kind": "youtube#videoListResponse",
    "etag": "abc123",
    "pageInfo": {"totalResults": 1, "resultsPerPage": 5},
    "items": [
        {
            "kind": "youtube#video",
            "etag": "def456",
            "id": "dQw4w9WgXcQ",
            "snippet": {
                "publishedAt": "2009-10-25T06:57:33Z",
                "channelId": "UCuAXFkgsw1L7xaCfnd5JJOw",
                "title": "Rick Astley - Never Gonna Give You Up",
                "description": "The official video...",
                "channelTitle": "Rick Astley",
                "tags": ["rick astley", "never gonna give you up"],
                "categoryId": "10",
                "liveBroadcastContent": "none",
            },
            "statistics": {
                "viewCount": "1500000000",
                "likeCount": "15000000",
                "favoriteCount": "0",
                "commentCount": "3000000",
            },
        }
    ],
}

SAMPLE_GET_RATING_RESPONSE = {
    "kind": "youtube#videoGetRatingResponse",
    "etag": "ghi789",
    "items": [
        {"videoId": "dQw4w9WgXcQ", "rating": "like"},
    ],
}


# ---------------------------------------------------------------------------
# Model tests
# ---------------------------------------------------------------------------


class TestVideoModels:
    """Tests for Pydantic video models."""

    def test_video_from_api_response(self) -> None:
        video = Video.model_validate(SAMPLE_VIDEO_LIST_RESPONSE["items"][0])
        assert video.id == "dQw4w9WgXcQ"
        assert video.snippet is not None
        assert video.snippet.title == "Rick Astley - Never Gonna Give You Up"
        assert video.snippet.channel_id == "UCuAXFkgsw1L7xaCfnd5JJOw"
        assert video.snippet.tags == ["rick astley", "never gonna give you up"]
        assert video.snippet.category_id == "10"

    def test_video_statistics_as_strings(self) -> None:
        video = Video.model_validate(SAMPLE_VIDEO_LIST_RESPONSE["items"][0])
        assert video.statistics is not None
        assert video.statistics.view_count == "1500000000"
        assert video.statistics.like_count == "15000000"

    def test_video_list_response(self) -> None:
        response = VideoListResponse.model_validate(SAMPLE_VIDEO_LIST_RESPONSE)
        assert response.kind == "youtube#videoListResponse"
        assert len(response.items) == 1
        assert response.page_info is not None
        assert response.page_info.total_results == 1
        assert response.next_page_token is None

    def test_video_model_serialization(self) -> None:
        """Video models should round-trip through camelCase serialization."""
        video = Video(
            snippet=VideoSnippet(
                title="Test Video",
                description="A test",
                category_id="22",
            ),
            status=VideoStatus(
                privacy_status=PrivacyStatus.PRIVATE,
            ),
        )
        dumped = video.model_dump(by_alias=True, exclude_none=True)
        assert dumped["snippet"]["title"] == "Test Video"
        assert dumped["snippet"]["categoryId"] == "22"
        assert dumped["status"]["privacyStatus"] == "private"

    def test_video_get_rating_response(self) -> None:
        response = VideoGetRatingResponse.model_validate(SAMPLE_GET_RATING_RESPONSE)
        assert len(response.items) == 1
        assert response.items[0].video_id == "dQw4w9WgXcQ"
        assert response.items[0].rating == Rating.LIKE

    def test_upload_status_enum(self) -> None:
        assert UploadStatus.PROCESSED == "processed"
        assert UploadStatus.FAILED == "failed"

    def test_video_license_enum(self) -> None:
        assert VideoLicense.YOUTUBE == "youtube"
        assert VideoLicense.CREATIVE_COMMON == "creativeCommon"

    def test_video_part_enum_values(self) -> None:
        assert VideoPart.SNIPPET == "snippet"
        assert VideoPart.STATISTICS == "statistics"
        assert VideoPart.CONTENT_DETAILS == "contentDetails"
        assert VideoPart.LIVE_STREAMING_DETAILS == "liveStreamingDetails"

    def test_ignores_unknown_fields(self) -> None:
        """Forward compatibility: unknown fields should be silently ignored."""
        data = {
            "kind": "youtube#video",
            "id": "test123",
            "brandNewField": "should not break",
        }
        video = Video.model_validate(data)
        assert video.id == "test123"

    def test_frozen_model(self) -> None:
        video = Video(id="test")
        with pytest.raises(ValidationError):
            video.id = "changed"


# ---------------------------------------------------------------------------
# API tests
# ---------------------------------------------------------------------------


class TestVideosAPI:
    """Tests for the VideosAPI methods with mocked HTTP."""

    async def test_list_by_chart(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_VIDEOS_RE, payload=SAMPLE_VIDEO_LIST_RESPONSE)
                response = await client.videos.list(
                    parts=[VideoPart.SNIPPET, VideoPart.STATISTICS],
                    chart="mostPopular",
                    max_results=1,
                )
                assert isinstance(response, VideoListResponse)
                assert len(response.items) == 1
                assert response.items[0].id == "dQw4w9WgXcQ"

    async def test_list_by_id(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_VIDEOS_RE, payload=SAMPLE_VIDEO_LIST_RESPONSE)
                response = await client.videos.list(
                    parts=[VideoPart.SNIPPET],
                    id=["dQw4w9WgXcQ"],
                )
                assert len(response.items) == 1

    async def test_delete(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.delete(_VIDEOS_RE, status=204)
                await client.videos.delete(id="dQw4w9WgXcQ")

    async def test_rate(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.post(_RATE_RE, status=204)
                await client.videos.rate(id="dQw4w9WgXcQ", rating=Rating.LIKE)

    async def test_get_rating(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_GET_RATING_RE, payload=SAMPLE_GET_RATING_RESPONSE)
                response = await client.videos.get_rating(id="dQw4w9WgXcQ")
                assert isinstance(response, VideoGetRatingResponse)
                assert response.items[0].rating == Rating.LIKE

    async def test_update(self) -> None:
        updated_video_payload = {
            "kind": "youtube#video",
            "id": "dQw4w9WgXcQ",
            "snippet": {
                "title": "Updated Title",
                "categoryId": "22",
            },
        }
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.put(_VIDEOS_RE, payload=updated_video_payload)
                body = Video(
                    id="dQw4w9WgXcQ",
                    snippet=VideoSnippet(title="Updated Title", category_id="22"),
                )
                result = await client.videos.update(body)
                assert result.snippet is not None
                assert result.snippet.title == "Updated Title"
