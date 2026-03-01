"""Pydantic models for the YouTube Playlists resource.

Reference: https://developers.google.com/youtube/v3/docs/playlists
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import (
    Localization,
    PageInfo,
    PrivacyStatus,
    ThumbnailDetails,
)

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class PlaylistPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    STATUS = "status"
    CONTENT_DETAILS = "contentDetails"
    PLAYER = "player"
    LOCALIZATIONS = "localizations"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class PlaylistSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Playlist resource."""

    published_at: str | None = None
    channel_id: str | None = None
    title: str | None = None
    description: str | None = None
    thumbnails: ThumbnailDetails | None = None
    channel_title: str | None = None
    default_language: str | None = None
    localized: Localization | None = None


class PlaylistStatus(YouTubeBaseModel):
    """The ``status`` part of a Playlist resource."""

    privacy_status: PrivacyStatus | None = None


class PlaylistContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a Playlist resource."""

    item_count: int | None = None


class PlaylistPlayer(YouTubeBaseModel):
    """The ``player`` part of a Playlist resource."""

    embed_html: str | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class Playlist(YouTubeBaseModel):
    """A YouTube Playlist resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: PlaylistSnippet | None = None
    status: PlaylistStatus | None = None
    content_details: PlaylistContentDetails | None = None
    player: PlaylistPlayer | None = None
    localizations: dict[str, Localization] | None = None


class PlaylistListResponse(YouTubeBaseModel):
    """Response from ``playlists.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[Playlist] = []
