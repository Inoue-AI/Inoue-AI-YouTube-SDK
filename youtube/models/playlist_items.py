"""Pydantic models for the YouTube PlaylistItems resource.

Reference: https://developers.google.com/youtube/v3/docs/playlistItems
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import (
    PageInfo,
    PrivacyStatus,
    ResourceId,
    ThumbnailDetails,
)

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class PlaylistItemPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    CONTENT_DETAILS = "contentDetails"
    STATUS = "status"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class PlaylistItemSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a PlaylistItem resource."""

    published_at: str | None = None
    channel_id: str | None = None
    title: str | None = None
    description: str | None = None
    thumbnails: ThumbnailDetails | None = None
    channel_title: str | None = None
    playlist_id: str | None = None
    position: int | None = None
    resource_id: ResourceId | None = None
    video_owner_channel_title: str | None = None
    video_owner_channel_id: str | None = None


class PlaylistItemContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a PlaylistItem resource."""

    video_id: str | None = None
    note: str | None = None
    video_published_at: str | None = None


class PlaylistItemStatus(YouTubeBaseModel):
    """The ``status`` part of a PlaylistItem resource."""

    privacy_status: PrivacyStatus | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class PlaylistItem(YouTubeBaseModel):
    """A YouTube PlaylistItem resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: PlaylistItemSnippet | None = None
    content_details: PlaylistItemContentDetails | None = None
    status: PlaylistItemStatus | None = None


class PlaylistItemListResponse(YouTubeBaseModel):
    """Response from ``playlistItems.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[PlaylistItem] = []
