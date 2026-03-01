"""Playlists API namespace.

Wraps the following YouTube endpoints:

* GET    /playlists — list playlists
* POST   /playlists — create a playlist
* PUT    /playlists — update a playlist
* DELETE /playlists — delete a playlist

Reference: https://developers.google.com/youtube/v3/docs/playlists
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.playlists import (
    Playlist,
    PlaylistListResponse,
    PlaylistPart,
)

_BASE = YOUTUBE_API_BASE_URL


class PlaylistsAPI(BaseAPI):
    """Methods for managing YouTube playlists."""

    async def list(
        self,
        *,
        parts: list[PlaylistPart],
        channel_id: str | None = None,
        id: str | list[str] | None = None,
        mine: bool | None = None,
        max_results: int = 5,
        page_token: str | None = None,
        hl: str | None = None,
    ) -> PlaylistListResponse:
        """Return a list of playlists matching the given filter."""
        params: dict[str, Any] = {
            "part": _join(parts),
            "maxResults": max_results,
        }
        if channel_id is not None:
            params["channelId"] = channel_id
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if mine is not None:
            params["mine"] = str(mine).lower()
        if page_token is not None:
            params["pageToken"] = page_token
        if hl is not None:
            params["hl"] = hl

        payload = await self._session.get(f"{_BASE}/playlists", params=params)
        return PlaylistListResponse.model_validate(payload)

    async def iter_mine(
        self,
        *,
        parts: list[PlaylistPart],
        max_results: int = 25,
    ) -> AsyncIterator[Playlist]:
        """Async generator that pages through the authenticated user's playlists."""
        page_token: str | None = None
        while True:
            page = await self.list(parts=parts, mine=True, max_results=max_results, page_token=page_token)
            for item in page.items:
                yield item
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    async def insert(
        self,
        body: Playlist,
        *,
        parts: list[PlaylistPart] | None = None,
    ) -> Playlist:
        """Create a new playlist."""
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.post(
            f"{_BASE}/playlists",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Playlist.model_validate(payload)

    async def update(
        self,
        body: Playlist,
        *,
        parts: list[PlaylistPart] | None = None,
    ) -> Playlist:
        """Update an existing playlist.

        The ``body`` must contain ``id`` and ``snippet.title``.
        """
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.put(
            f"{_BASE}/playlists",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Playlist.model_validate(payload)

    async def delete(self, *, id: str) -> None:
        """Delete a playlist."""
        await self._session.delete(f"{_BASE}/playlists", params={"id": id})


def _join(parts: list[PlaylistPart]) -> str:
    return ",".join(p.value for p in parts)


def _infer_parts(body: Playlist) -> list[PlaylistPart]:
    parts: list[PlaylistPart] = []
    if body.snippet is not None:
        parts.append(PlaylistPart.SNIPPET)
    if body.status is not None:
        parts.append(PlaylistPart.STATUS)
    if body.localizations is not None:
        parts.append(PlaylistPart.LOCALIZATIONS)
    return parts or [PlaylistPart.SNIPPET]
