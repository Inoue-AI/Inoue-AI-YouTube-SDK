"""CommentThreads API namespace.

Wraps the following YouTube endpoints:

* GET  /commentThreads — list comment threads
* POST /commentThreads — post a top-level comment

Reference: https://developers.google.com/youtube/v3/docs/commentThreads
"""

from __future__ import annotations

from collections.abc import AsyncIterator, Sequence
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.apis.params import join_ids, join_parts
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.comment_threads import (
    CommentThread,
    CommentThreadListResponse,
    CommentThreadOrder,
    CommentThreadPart,
)
from youtube.models.comments import ModerationStatus, TextFormat

_BASE = YOUTUBE_API_BASE_URL


class CommentThreadsAPI(BaseAPI):
    """Methods for managing YouTube comment threads."""

    async def list(
        self,
        *,
        parts: Sequence[CommentThreadPart],
        video_id: str | None = None,
        channel_id: str | None = None,
        id: str | Sequence[str] | None = None,
        all_threads_related_to_channel_id: str | None = None,
        max_results: int = 20,
        page_token: str | None = None,
        moderation_status: ModerationStatus | None = None,
        order: CommentThreadOrder | None = None,
        search_terms: str | None = None,
        text_format: TextFormat | None = None,
    ) -> CommentThreadListResponse:
        """Return a list of comment threads.

        Exactly one of ``video_id``, ``channel_id``, ``id``, or
        ``all_threads_related_to_channel_id`` must be provided.
        """
        params: dict[str, Any] = {
            "part": join_parts(parts),
            "maxResults": max_results,
        }
        if video_id is not None:
            params["videoId"] = video_id
        if channel_id is not None:
            params["channelId"] = channel_id
        if id is not None:
            params["id"] = join_ids(id)
        if all_threads_related_to_channel_id is not None:
            params["allThreadsRelatedToChannelId"] = all_threads_related_to_channel_id
        if page_token is not None:
            params["pageToken"] = page_token
        if moderation_status is not None:
            params["moderationStatus"] = moderation_status.value
        if order is not None:
            params["order"] = order.value
        if search_terms is not None:
            params["searchTerms"] = search_terms
        if text_format is not None:
            params["textFormat"] = text_format.value

        payload = await self._session.get(f"{_BASE}/commentThreads", params=params)
        return CommentThreadListResponse.model_validate(payload)

    async def iter_video_threads(
        self,
        video_id: str,
        *,
        parts: Sequence[CommentThreadPart],
        max_results: int = 100,
        order: CommentThreadOrder | None = None,
    ) -> AsyncIterator[CommentThread]:
        """Async generator that pages through all comment threads on a video."""
        page_token: str | None = None
        while True:
            page = await self.list(
                parts=parts,
                video_id=video_id,
                max_results=max_results,
                page_token=page_token,
                order=order,
            )
            for thread in page.items:
                yield thread
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    async def insert(
        self,
        body: CommentThread,
        *,
        parts: Sequence[CommentThreadPart] | None = None,
    ) -> CommentThread:
        """Post a new top-level comment on a video.

        The ``body`` must include ``snippet.channel_id``,
        ``snippet.video_id``, and
        ``snippet.top_level_comment.snippet.text_original``.
        """
        parts = parts or [CommentThreadPart.SNIPPET]
        params: dict[str, Any] = {"part": join_parts(parts)}
        payload = await self._session.post(
            f"{_BASE}/commentThreads",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return CommentThread.model_validate(payload)
