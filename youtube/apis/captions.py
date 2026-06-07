"""Captions API namespace.

Wraps the following YouTube endpoints:

* GET    /captions       — list caption tracks
* POST   /upload/.../captions — upload a caption track
* PUT    /upload/.../captions — update a caption track
* GET    /captions/{id}  — download a caption track
* DELETE /captions       — delete a caption track

Reference: https://developers.google.com/youtube/v3/docs/captions
"""

from __future__ import annotations

from collections.abc import Sequence
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.apis.params import join_ids, join_parts
from youtube.config import YOUTUBE_API_BASE_URL, YOUTUBE_UPLOAD_BASE_URL
from youtube.models.captions import (
    Caption,
    CaptionFormat,
    CaptionListResponse,
    CaptionPart,
)

_BASE = YOUTUBE_API_BASE_URL
_UPLOAD = YOUTUBE_UPLOAD_BASE_URL


class CaptionsAPI(BaseAPI):
    """Methods for managing YouTube caption tracks."""

    async def list(
        self,
        *,
        parts: Sequence[CaptionPart],
        video_id: str,
        id: str | Sequence[str] | None = None,
    ) -> CaptionListResponse:
        """Return caption tracks for a video."""
        params: dict[str, Any] = {
            "part": join_parts(parts),
            "videoId": video_id,
        }
        if id is not None:
            params["id"] = join_ids(id)

        payload = await self._session.get(f"{_BASE}/captions", params=params)
        return CaptionListResponse.model_validate(payload)

    async def insert(
        self,
        body: Caption,
        caption_data: bytes,
        *,
        parts: Sequence[CaptionPart] | None = None,
        content_type: str = "application/octet-stream",
    ) -> Caption:
        """Upload a new caption track.

        The ``body`` must include ``snippet.video_id``,
        ``snippet.language``, and ``snippet.name``.
        """
        parts = parts or [CaptionPart.SNIPPET]
        params: dict[str, Any] = {"part": join_parts(parts)}
        metadata = body.model_dump(by_alias=True, exclude_none=True)
        payload = await self._session.upload_multipart(
            f"{_UPLOAD}/captions",
            params=params,
            metadata=metadata,
            data=caption_data,
            content_type=content_type,
        )
        return Caption.model_validate(payload)

    async def insert_from_file(
        self,
        body: Caption,
        file_path: str,
        *,
        parts: Sequence[CaptionPart] | None = None,
        content_type: str = "application/octet-stream",
    ) -> Caption:
        """Upload a new caption track from a local file."""
        with open(file_path, "rb") as fh:
            caption_data = fh.read()
        return await self.insert(
            body,
            caption_data,
            parts=parts,
            content_type=content_type,
        )

    async def update(
        self,
        body: Caption,
        caption_data: bytes | None = None,
        *,
        parts: Sequence[CaptionPart] | None = None,
        content_type: str = "application/octet-stream",
    ) -> Caption:
        """Update an existing caption track.

        If ``caption_data`` is provided, the caption file is replaced.
        Otherwise only the metadata (e.g. ``is_draft``) is updated.
        """
        parts = parts or [CaptionPart.SNIPPET]
        params: dict[str, Any] = {"part": join_parts(parts)}
        metadata = body.model_dump(by_alias=True, exclude_none=True)

        if caption_data is not None:
            payload = await self._session.upload_multipart(
                f"{_UPLOAD}/captions",
                params=params,
                metadata=metadata,
                data=caption_data,
                content_type=content_type,
            )
        else:
            payload = await self._session.put(
                f"{_BASE}/captions",
                json=metadata,
                params=params,
            )
        return Caption.model_validate(payload)

    async def download(
        self,
        *,
        id: str,
        tfmt: CaptionFormat | None = None,
        tlang: str | None = None,
    ) -> bytes:
        """Download a caption track as raw bytes.

        Parameters
        ----------
        id:
            Caption track ID.
        tfmt:
            Output format (srt, vtt, etc.).  If omitted, the original format
            is returned.
        tlang:
            BCP-47 language code to translate the captions into.
        """
        params: dict[str, Any] = {}
        if tfmt is not None:
            params["tfmt"] = tfmt.value
        if tlang is not None:
            params["tlang"] = tlang

        return await self._session.get_binary(
            f"{_BASE}/captions/{id}",
            params=params or None,
        )

    async def delete(self, *, id: str) -> None:
        """Delete a caption track."""
        await self._session.delete(f"{_BASE}/captions", params={"id": id})
