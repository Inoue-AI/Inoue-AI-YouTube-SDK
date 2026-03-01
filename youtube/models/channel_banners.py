"""Pydantic models for the YouTube ChannelBanners resource.

Reference: https://developers.google.com/youtube/v3/docs/channelBanners/insert
"""

from __future__ import annotations

from youtube.models.base import YouTubeBaseModel


class ChannelBannerResource(YouTubeBaseModel):
    """Response from ``channelBanners.insert``.

    The ``url`` field must be used in a subsequent ``channels.update`` call
    to set ``brandingSettings.image.bannerExternalUrl``.
    """

    kind: str | None = None
    etag: str | None = None
    url: str | None = None
