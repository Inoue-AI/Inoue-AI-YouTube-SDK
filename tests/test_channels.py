"""Tests for the Channels API and models."""

from __future__ import annotations

import re

from aioresponses import aioresponses

from youtube.client import YouTubeClient
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.channels import (
    Channel,
    ChannelBrandingChannel,
    ChannelBrandingSettings,
    ChannelListResponse,
    ChannelPart,
)

_BASE = YOUTUBE_API_BASE_URL
_CHANNELS_RE = re.compile(r"^https://www\.googleapis\.com/youtube/v3/channels(\?.*)?$")

SAMPLE_CHANNEL_LIST_RESPONSE = {
    "kind": "youtube#channelListResponse",
    "etag": "abc123",
    "pageInfo": {"totalResults": 1, "resultsPerPage": 5},
    "items": [
        {
            "kind": "youtube#channel",
            "etag": "def456",
            "id": "UCuAXFkgsw1L7xaCfnd5JJOw",
            "snippet": {
                "title": "Rick Astley",
                "description": "Official channel",
                "customUrl": "@rickastley",
                "publishedAt": "2006-09-12T00:00:00Z",
                "country": "GB",
            },
            "statistics": {
                "viewCount": "2000000000",
                "subscriberCount": "5000000",
                "hiddenSubscriberCount": False,
                "videoCount": "100",
            },
            "contentDetails": {
                "relatedPlaylists": {
                    "likes": "LL",
                    "uploads": "UUuAXFkgsw1L7xaCfnd5JJOw",
                }
            },
        }
    ],
}


class TestChannelModels:
    def test_channel_from_api_response(self) -> None:
        channel = Channel.model_validate(SAMPLE_CHANNEL_LIST_RESPONSE["items"][0])
        assert channel.id == "UCuAXFkgsw1L7xaCfnd5JJOw"
        assert channel.snippet is not None
        assert channel.snippet.title == "Rick Astley"
        assert channel.snippet.custom_url == "@rickastley"
        assert channel.snippet.country == "GB"

    def test_channel_statistics(self) -> None:
        channel = Channel.model_validate(SAMPLE_CHANNEL_LIST_RESPONSE["items"][0])
        assert channel.statistics is not None
        assert channel.statistics.subscriber_count == "5000000"
        assert channel.statistics.hidden_subscriber_count is False

    def test_channel_content_details(self) -> None:
        channel = Channel.model_validate(SAMPLE_CHANNEL_LIST_RESPONSE["items"][0])
        assert channel.content_details is not None
        assert channel.content_details.related_playlists is not None
        assert channel.content_details.related_playlists.uploads == "UUuAXFkgsw1L7xaCfnd5JJOw"

    def test_channel_list_response(self) -> None:
        response = ChannelListResponse.model_validate(SAMPLE_CHANNEL_LIST_RESPONSE)
        assert response.kind == "youtube#channelListResponse"
        assert len(response.items) == 1

    def test_channel_update_serialization(self) -> None:
        channel = Channel(
            id="UCuAXFkgsw1L7xaCfnd5JJOw",
            branding_settings=ChannelBrandingSettings(
                channel=ChannelBrandingChannel(
                    description="Updated description",
                    country="US",
                ),
            ),
        )
        dumped = channel.model_dump(by_alias=True, exclude_none=True)
        assert dumped["id"] == "UCuAXFkgsw1L7xaCfnd5JJOw"
        assert dumped["brandingSettings"]["channel"]["description"] == "Updated description"
        assert dumped["brandingSettings"]["channel"]["country"] == "US"


class TestChannelsAPI:
    async def test_list_mine(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_CHANNELS_RE, payload=SAMPLE_CHANNEL_LIST_RESPONSE)
                response = await client.channels.list(
                    parts=[ChannelPart.SNIPPET, ChannelPart.STATISTICS],
                    mine=True,
                )
                assert isinstance(response, ChannelListResponse)
                assert len(response.items) == 1
                assert response.items[0].snippet.title == "Rick Astley"

    async def test_list_by_handle(self) -> None:
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.get(_CHANNELS_RE, payload=SAMPLE_CHANNEL_LIST_RESPONSE)
                response = await client.channels.list(
                    parts=[ChannelPart.SNIPPET],
                    for_handle="@rickastley",
                )
                assert len(response.items) == 1

    async def test_update(self) -> None:
        updated_payload = {
            "kind": "youtube#channel",
            "id": "UCtest",
            "brandingSettings": {
                "channel": {"description": "New desc"},
            },
        }
        async with YouTubeClient(access_token="token") as client:
            with aioresponses() as mocked:
                mocked.put(_CHANNELS_RE, payload=updated_payload)
                body = Channel(
                    id="UCtest",
                    branding_settings=ChannelBrandingSettings(
                        channel=ChannelBrandingChannel(description="New desc"),
                    ),
                )
                result = await client.channels.update(body)
                assert result.branding_settings.channel.description == "New desc"
