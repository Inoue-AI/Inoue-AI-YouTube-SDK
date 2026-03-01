"""Pydantic models for the YouTube Channels resource.

Reference: https://developers.google.com/youtube/v3/docs/channels
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

class ChannelPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    CONTENT_DETAILS = "contentDetails"
    STATISTICS = "statistics"
    TOPIC_DETAILS = "topicDetails"
    STATUS = "status"
    BRANDING_SETTINGS = "brandingSettings"
    LOCALIZATIONS = "localizations"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class ChannelSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Channel resource."""

    title: str | None = None
    description: str | None = None
    custom_url: str | None = None
    published_at: str | None = None
    thumbnails: ThumbnailDetails | None = None
    default_language: str | None = None
    localized: Localization | None = None
    country: str | None = None


class ChannelRelatedPlaylists(YouTubeBaseModel):
    """Auto-generated playlists linked to the channel."""

    likes: str | None = None
    uploads: str | None = None


class ChannelContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a Channel resource."""

    related_playlists: ChannelRelatedPlaylists | None = None


class ChannelStatistics(YouTubeBaseModel):
    """The ``statistics`` part of a Channel resource."""

    view_count: str | None = None
    subscriber_count: str | None = None
    hidden_subscriber_count: bool | None = None
    video_count: str | None = None


class ChannelTopicDetails(YouTubeBaseModel):
    """The ``topicDetails`` part of a Channel resource."""

    topic_ids: list[str] | None = None
    topic_categories: list[str] | None = None


class ChannelStatus(YouTubeBaseModel):
    """The ``status`` part of a Channel resource."""

    privacy_status: PrivacyStatus | None = None
    is_linked: bool | None = None
    long_uploads_status: str | None = None
    made_for_kids: bool | None = None
    self_declared_made_for_kids: bool | None = None


class ChannelBrandingChannel(YouTubeBaseModel):
    """Branding settings for the channel itself."""

    title: str | None = None
    description: str | None = None
    keywords: str | None = None
    tracking_analytics_account_id: str | None = None
    unsubscribed_trailer: str | None = None
    default_language: str | None = None
    country: str | None = None


class ChannelBrandingImage(YouTubeBaseModel):
    """Branding settings for channel images."""

    banner_external_url: str | None = None


class ChannelBrandingSettings(YouTubeBaseModel):
    """The ``brandingSettings`` part of a Channel resource."""

    channel: ChannelBrandingChannel | None = None
    image: ChannelBrandingImage | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class Channel(YouTubeBaseModel):
    """A YouTube Channel resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: ChannelSnippet | None = None
    content_details: ChannelContentDetails | None = None
    statistics: ChannelStatistics | None = None
    topic_details: ChannelTopicDetails | None = None
    status: ChannelStatus | None = None
    branding_settings: ChannelBrandingSettings | None = None
    localizations: dict[str, Localization] | None = None


class ChannelListResponse(YouTubeBaseModel):
    """Response from ``channels.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[Channel] = []
