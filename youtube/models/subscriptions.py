"""Pydantic models for the YouTube Subscriptions resource.

Reference: https://developers.google.com/youtube/v3/docs/subscriptions
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import PageInfo, ResourceId, ThumbnailDetails

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------


class SubscriptionPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    CONTENT_DETAILS = "contentDetails"
    SUBSCRIBER_SNIPPET = "subscriberSnippet"


class SubscriptionOrder(StrEnum):
    """Sort order for subscription lists."""

    ALPHABETICAL = "alphabetical"
    RELEVANCE = "relevance"
    UNREAD = "unread"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------


class SubscriptionSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Subscription resource."""

    published_at: str | None = None
    title: str | None = None
    description: str | None = None
    resource_id: ResourceId | None = None
    channel_id: str | None = None
    thumbnails: ThumbnailDetails | None = None


class SubscriptionContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a Subscription resource."""

    total_item_count: int | None = None
    new_item_count: int | None = None
    activity_type: str | None = None


class SubscriberSnippet(YouTubeBaseModel):
    """The ``subscriberSnippet`` part of a Subscription resource."""

    title: str | None = None
    description: str | None = None
    channel_id: str | None = None
    thumbnails: ThumbnailDetails | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------


class Subscription(YouTubeBaseModel):
    """A YouTube Subscription resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: SubscriptionSnippet | None = None
    content_details: SubscriptionContentDetails | None = None
    subscriber_snippet: SubscriberSnippet | None = None


class SubscriptionListResponse(YouTubeBaseModel):
    """Response from ``subscriptions.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[Subscription] = []
