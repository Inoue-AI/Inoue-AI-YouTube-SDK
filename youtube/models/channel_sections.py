"""Pydantic models for the YouTube ChannelSections resource.

Reference: https://developers.google.com/youtube/v3/docs/channelSections
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import Localization

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------


class ChannelSectionPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    CONTENT_DETAILS = "contentDetails"


class ChannelSectionType(StrEnum):
    """Type of content in a channel section."""

    SINGLE_PLAYLIST = "singlePlaylist"
    MULTIPLE_PLAYLISTS = "multiplePlaylists"
    ALL_PLAYLISTS = "allPlaylists"
    RECENT_UPLOADS = "recentUploads"
    POPULAR_UPLOADS = "popularUploads"
    MULTIPLE_CHANNELS = "multipleChannels"
    SUBSCRIPTIONS = "subscriptions"
    LIVE_EVENTS = "liveEvents"
    UPCOMING_EVENTS = "upcomingEvents"
    COMPLETED_EVENTS = "completedEvents"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------


class ChannelSectionSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a ChannelSection resource."""

    type: ChannelSectionType | None = None
    channel_id: str | None = None
    title: str | None = None
    position: int | None = None
    default_language: str | None = None
    localized: Localization | None = None


class ChannelSectionContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a ChannelSection resource."""

    playlists: list[str] | None = None
    channels: list[str] | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------


class ChannelSection(YouTubeBaseModel):
    """A YouTube ChannelSection resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: ChannelSectionSnippet | None = None
    content_details: ChannelSectionContentDetails | None = None


class ChannelSectionListResponse(YouTubeBaseModel):
    """Response from ``channelSections.list``."""

    kind: str | None = None
    etag: str | None = None
    items: list[ChannelSection] = []
