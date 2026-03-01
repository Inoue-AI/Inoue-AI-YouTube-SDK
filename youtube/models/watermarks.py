"""Pydantic models for the YouTube Watermarks resource.

Reference: https://developers.google.com/youtube/v3/docs/watermarks
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class WatermarkTimingType(StrEnum):
    """When the watermark appears relative to the video."""

    OFFSET_FROM_START = "offsetFromStart"
    OFFSET_FROM_END = "offsetFromEnd"


class WatermarkCornerPosition(StrEnum):
    """Corner position for the watermark overlay."""

    TOP_LEFT = "topLeft"
    TOP_RIGHT = "topRight"
    BOTTOM_LEFT = "bottomLeft"
    BOTTOM_RIGHT = "bottomRight"


# ---------------------------------------------------------------------------
# Resource parts
# ---------------------------------------------------------------------------

class WatermarkTiming(YouTubeBaseModel):
    """Timing configuration for a watermark."""

    type: WatermarkTimingType | None = None
    offset_ms: str | None = None
    duration_ms: str | None = None


class WatermarkPosition(YouTubeBaseModel):
    """Position configuration for a watermark."""

    type: str | None = None
    corner_position: WatermarkCornerPosition | None = None


class Watermark(YouTubeBaseModel):
    """A YouTube Watermark resource (request body for ``watermarks.set``)."""

    timing: WatermarkTiming | None = None
    position: WatermarkPosition | None = None
    image_url: str | None = None
    image_bytes: str | None = None
    target_channel_id: str | None = None
