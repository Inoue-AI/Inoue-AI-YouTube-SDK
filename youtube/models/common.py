"""Shared types reused across multiple YouTube API resources."""

from __future__ import annotations

from enum import StrEnum

from pydantic import Field

from youtube.models.base import YouTubeBaseModel

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class PrivacyStatus(StrEnum):
    """Visibility setting for videos, playlists, and other resources."""

    PUBLIC = "public"
    UNLISTED = "unlisted"
    PRIVATE = "private"


# ---------------------------------------------------------------------------
# Shared resource types
# ---------------------------------------------------------------------------

class Thumbnail(YouTubeBaseModel):
    """A single thumbnail image."""

    url: str | None = None
    width: int | None = None
    height: int | None = None


class ThumbnailDetails(YouTubeBaseModel):
    """Collection of thumbnail images at different resolutions.

    The ``default_thumbnail`` field maps to the ``"default"`` key in the API
    response (``default`` is a Python reserved word in some contexts).
    """

    default_thumbnail: Thumbnail | None = Field(None, alias="default")
    medium: Thumbnail | None = None
    high: Thumbnail | None = None
    standard: Thumbnail | None = None
    maxres: Thumbnail | None = None


class PageInfo(YouTubeBaseModel):
    """Pagination metadata returned by every list endpoint."""

    total_results: int | None = None
    results_per_page: int | None = None


class ResourceId(YouTubeBaseModel):
    """Identifies a specific YouTube resource (video, channel, or playlist)."""

    kind: str | None = None
    video_id: str | None = None
    channel_id: str | None = None
    playlist_id: str | None = None


class Localization(YouTubeBaseModel):
    """Localized title and description for a resource."""

    title: str | None = None
    description: str | None = None
