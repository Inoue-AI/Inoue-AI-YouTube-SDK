"""Pydantic models for the YouTube CommentThreads resource.

Reference: https://developers.google.com/youtube/v3/docs/commentThreads
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.comments import Comment
from youtube.models.common import PageInfo

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------


class CommentThreadPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    REPLIES = "replies"


class CommentThreadOrder(StrEnum):
    """Sort order for comment threads."""

    TIME = "time"
    RELEVANCE = "relevance"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------


class CommentThreadSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a CommentThread resource."""

    channel_id: str | None = None
    video_id: str | None = None
    top_level_comment: Comment | None = None
    can_reply: bool | None = None
    total_reply_count: int | None = None
    is_public: bool | None = None


class CommentThreadReplies(YouTubeBaseModel):
    """The ``replies`` part of a CommentThread resource.

    Note: this may not include *all* replies.  Use ``comments.list`` with
    ``parent_id`` for the full reply set.
    """

    comments: list[Comment] = []


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------


class CommentThread(YouTubeBaseModel):
    """A YouTube CommentThread resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: CommentThreadSnippet | None = None
    replies: CommentThreadReplies | None = None


class CommentThreadListResponse(YouTubeBaseModel):
    """Response from ``commentThreads.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[CommentThread] = []
