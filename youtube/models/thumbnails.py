"""Pydantic models for the YouTube Thumbnails resource.

Reference: https://developers.google.com/youtube/v3/docs/thumbnails/set
"""

from __future__ import annotations

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import ThumbnailDetails


class ThumbnailSetResponse(YouTubeBaseModel):
    """Response from ``thumbnails.set``."""

    kind: str | None = None
    etag: str | None = None
    items: list[ThumbnailDetails] = []
