"""PlaylistItems API namespace.

Wraps the following YouTube endpoints:

* GET    /playlistItems — list items in a playlist
* POST   /playlistItems — add a video to a playlist
* PUT    /playlistItems — update a playlist item
* DELETE /playlistItems — remove an item from a playlist

Reference: https://developers.google.com/youtube/v3/docs/playlistItems
"""

from __future__ import annotations

from collections.abc import AsyncIterator, Sequence
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.apis.params import join_ids, join_parts
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.playlist_items import (
    PlaylistItem,
    PlaylistItemListResponse,
    PlaylistItemPart,
)

_BASE = YOUTUBE_API_BASE_URL


class PlaylistItemsAPI(BaseAPI):
    """Methods for managing items within YouTube playlists."""

    async def list(
        self,
        *,
        parts: Sequence[PlaylistItemPart],
        playlist_id: str | None = None,
        id: str | Sequence[str] | None = None,
        video_id: str | None = None,
        max_results: int = 5,
        page_token: str | None = None,
    ) -> PlaylistItemListResponse:
        """Return items from a playlist."""
        params: dict[str, Any] = {
            "part": join_parts(parts),
            "maxResults": max_results,
        }
        if playlist_id is not None:
            params["playlistId"] = playlist_id
        if id is not None:
            params["id"] = join_ids(id)
        if video_id is not None:
            params["videoId"] = video_id
        if page_token is not None:
            params["pageToken"] = page_token

        payload = await self._session.get(f"{_BASE}/playlistItems", params=params)
        return PlaylistItemListResponse.model_validate(payload)

    async def iter_playlist(
        self,
        playlist_id: str,
        *,
        parts: Sequence[PlaylistItemPart],
        max_results: int = 50,
    ) -> AsyncIterator[PlaylistItem]:
        """Async generator that pages through all items in a playlist."""
        page_token: str | None = None
        while True:
            page = await self.list(
                parts=parts,
                playlist_id=playlist_id,
                max_results=max_results,
                page_token=page_token,
            )
            for item in page.items:
                yield item
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    async def insert(
        self,
        body: PlaylistItem,
        *,
        parts: Sequence[PlaylistItemPart] | None = None,
    ) -> PlaylistItem:
        """Add a video to a playlist.

        The ``body`` must include ``snippet.playlist_id`` and
        ``snippet.resource_id`` (with ``kind`` and ``video_id``).
        """
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": join_parts(parts)}
        payload = await self._session.post(
            f"{_BASE}/playlistItems",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return PlaylistItem.model_validate(payload)

    async def update(
        self,
        body: PlaylistItem,
        *,
        parts: Sequence[PlaylistItemPart] | None = None,
    ) -> PlaylistItem:
        """Update a playlist item (e.g. change position or note)."""
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": join_parts(parts)}
        payload = await self._session.put(
            f"{_BASE}/playlistItems",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return PlaylistItem.model_validate(payload)

    async def delete(self, *, id: str) -> None:
        """Remove an item from a playlist."""
        await self._session.delete(f"{_BASE}/playlistItems", params={"id": id})


def _infer_parts(body: PlaylistItem) -> list[PlaylistItemPart]:
    parts: list[PlaylistItemPart] = []
    if body.snippet is not None:
        parts.append(PlaylistItemPart.SNIPPET)
    if body.content_details is not None:
        parts.append(PlaylistItemPart.CONTENT_DETAILS)
    return parts or [PlaylistItemPart.SNIPPET]
