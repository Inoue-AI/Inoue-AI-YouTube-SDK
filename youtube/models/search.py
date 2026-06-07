"""Pydantic models for the YouTube Search resource.

Reference: https://developers.google.com/youtube/v3/docs/search/list
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import PageInfo, ResourceId, ThumbnailDetails

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------


class SearchPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    SNIPPET = "snippet"


class SearchOrder(StrEnum):
    """Sort order for search results."""

    DATE = "date"
    RATING = "rating"
    RELEVANCE = "relevance"
    TITLE = "title"
    VIDEO_COUNT = "videoCount"
    VIEW_COUNT = "viewCount"


class SafeSearch(StrEnum):
    """Safe-search filtering level."""

    MODERATE = "moderate"
    NONE = "none"
    STRICT = "strict"


class SearchResultType(StrEnum):
    """Type of resource in search results."""

    CHANNEL = "channel"
    PLAYLIST = "playlist"
    VIDEO = "video"


class EventType(StrEnum):
    """Live event type filter."""

    COMPLETED = "completed"
    LIVE = "live"
    UPCOMING = "upcoming"


class VideoCaption(StrEnum):
    """Caption availability filter."""

    ANY = "any"
    CLOSED_CAPTION = "closedCaption"
    NONE = "none"


class VideoDefinition(StrEnum):
    """Video definition filter."""

    ANY = "any"
    HIGH = "high"
    STANDARD = "standard"


class VideoDimension(StrEnum):
    """Video dimension filter."""

    ANY = "any"
    TWO_D = "2d"
    THREE_D = "3d"


class VideoDuration(StrEnum):
    """Video duration filter."""

    ANY = "any"
    LONG = "long"
    MEDIUM = "medium"
    SHORT = "short"


class VideoType(StrEnum):
    """Video type filter."""

    ANY = "any"
    EPISODE = "episode"
    MOVIE = "movie"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------


class SearchResultSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a SearchResult resource."""

    published_at: str | None = None
    channel_id: str | None = None
    title: str | None = None
    description: str | None = None
    thumbnails: ThumbnailDetails | None = None
    channel_title: str | None = None
    live_broadcast_content: str | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------


class SearchResult(YouTubeBaseModel):
    """A single search result.

    Note: unlike other resources, the ``id`` field is a :class:`ResourceId`
    object (not a plain string).
    """

    kind: str | None = None
    etag: str | None = None
    id: ResourceId | None = None
    snippet: SearchResultSnippet | None = None


class SearchListResponse(YouTubeBaseModel):
    """Response from ``search.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    region_code: str | None = None
    page_info: PageInfo | None = None
    items: list[SearchResult] = []
