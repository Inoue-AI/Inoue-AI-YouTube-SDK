"""Pydantic models for the YouTube Comments resource.

Reference: https://developers.google.com/youtube/v3/docs/comments
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import PageInfo

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class CommentPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"


class ModerationStatus(StrEnum):
    """Moderation status values for ``comments.setModerationStatus``."""

    HELD_FOR_REVIEW = "heldForReview"
    PUBLISHED = "published"
    REJECTED = "rejected"


class TextFormat(StrEnum):
    """Text format options for comment listing."""

    HTML = "html"
    PLAIN_TEXT = "plainText"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class AuthorChannelId(YouTubeBaseModel):
    """Wrapper for the author's channel ID inside a comment snippet."""

    value: str | None = None


class CommentSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Comment resource."""

    author_display_name: str | None = None
    author_profile_image_url: str | None = None
    author_channel_url: str | None = None
    author_channel_id: AuthorChannelId | None = None
    channel_id: str | None = None
    video_id: str | None = None
    text_display: str | None = None
    text_original: str | None = None
    parent_id: str | None = None
    can_rate: bool | None = None
    viewer_rating: str | None = None
    like_count: int | None = None
    moderation_status: ModerationStatus | None = None
    published_at: str | None = None
    updated_at: str | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class Comment(YouTubeBaseModel):
    """A YouTube Comment resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: CommentSnippet | None = None


class CommentListResponse(YouTubeBaseModel):
    """Response from ``comments.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[Comment] = []
