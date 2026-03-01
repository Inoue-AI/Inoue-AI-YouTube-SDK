"""Comments API namespace.

Wraps the following YouTube endpoints:

* GET    /comments                    — list comments (replies)
* POST   /comments                    — post a reply
* PUT    /comments                    — update a comment
* DELETE /comments                    — delete a comment
* POST   /comments/setModerationStatus — moderate comments

Reference: https://developers.google.com/youtube/v3/docs/comments
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.comments import (
    Comment,
    CommentListResponse,
    CommentPart,
    ModerationStatus,
    TextFormat,
)

_BASE = YOUTUBE_API_BASE_URL


class CommentsAPI(BaseAPI):
    """Methods for managing YouTube comments."""

    async def list(
        self,
        *,
        parts: list[CommentPart],
        id: str | list[str] | None = None,
        parent_id: str | None = None,
        max_results: int = 20,
        page_token: str | None = None,
        text_format: TextFormat | None = None,
    ) -> CommentListResponse:
        """Return a list of comments.

        Use ``parent_id`` to retrieve replies to a specific comment.
        """
        params: dict[str, Any] = {
            "part": _join(parts),
            "maxResults": max_results,
        }
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if parent_id is not None:
            params["parentId"] = parent_id
        if page_token is not None:
            params["pageToken"] = page_token
        if text_format is not None:
            params["textFormat"] = text_format.value

        payload = await self._session.get(f"{_BASE}/comments", params=params)
        return CommentListResponse.model_validate(payload)

    async def iter_replies(
        self,
        parent_id: str,
        *,
        parts: list[CommentPart],
        max_results: int = 100,
    ) -> AsyncIterator[Comment]:
        """Async generator that pages through all replies to a comment."""
        page_token: str | None = None
        while True:
            page = await self.list(
                parts=parts,
                parent_id=parent_id,
                max_results=max_results,
                page_token=page_token,
            )
            for comment in page.items:
                yield comment
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    async def insert(
        self,
        body: Comment,
        *,
        parts: list[CommentPart] | None = None,
    ) -> Comment:
        """Post a reply to an existing comment.

        The ``body`` must include ``snippet.parent_id`` and
        ``snippet.text_original``.

        Note: use ``comment_threads.insert`` for top-level comments.
        """
        parts = parts or [CommentPart.SNIPPET]
        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.post(
            f"{_BASE}/comments",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Comment.model_validate(payload)

    async def update(
        self,
        body: Comment,
        *,
        parts: list[CommentPart] | None = None,
    ) -> Comment:
        """Update an existing comment's text.

        The ``body`` must include ``id`` and ``snippet.text_original``.
        """
        parts = parts or [CommentPart.SNIPPET]
        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.put(
            f"{_BASE}/comments",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Comment.model_validate(payload)

    async def delete(self, *, id: str) -> None:
        """Delete a comment."""
        await self._session.delete(f"{_BASE}/comments", params={"id": id})

    async def set_moderation_status(
        self,
        *,
        id: str | list[str],
        moderation_status: ModerationStatus,
        ban_author: bool = False,
    ) -> None:
        """Set the moderation status of one or more comments."""
        comment_ids = ",".join(id) if isinstance(id, list) else id
        params: dict[str, Any] = {
            "id": comment_ids,
            "moderationStatus": moderation_status.value,
        }
        if ban_author:
            params["banAuthor"] = "true"
        await self._session.post_no_content(
            f"{_BASE}/comments/setModerationStatus",
            params=params,
        )


def _join(parts: list[CommentPart]) -> str:
    return ",".join(p.value for p in parts)
