"""Watermarks API namespace.

Wraps the following YouTube endpoints:

* POST /upload/.../watermarks/set — upload and set a watermark
* POST /watermarks/unset          — remove the watermark

Reference: https://developers.google.com/youtube/v3/docs/watermarks
"""

from __future__ import annotations

from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL, YOUTUBE_UPLOAD_BASE_URL
from youtube.models.watermarks import Watermark

_BASE = YOUTUBE_API_BASE_URL
_UPLOAD = YOUTUBE_UPLOAD_BASE_URL


class WatermarksAPI(BaseAPI):
    """Methods for managing YouTube channel watermarks."""

    async def set(
        self,
        *,
        channel_id: str,
        watermark: Watermark,
        image_data: bytes,
        content_type: str = "image/png",
    ) -> None:
        """Upload and set a watermark for a channel.

        Parameters
        ----------
        channel_id:
            The channel to set the watermark on.
        watermark:
            Watermark configuration (timing, position).
        image_data:
            Raw image bytes (JPEG or PNG, max 10 MB).
        content_type:
            MIME type — ``image/jpeg`` or ``image/png``.
        """
        params: dict[str, Any] = {"channelId": channel_id}
        await self._session.upload_binary(
            f"{_UPLOAD}/watermarks/set",
            params=params,
            data=image_data,
            content_type=content_type,
        )

    async def unset(self, *, channel_id: str) -> None:
        """Remove the watermark from a channel."""
        await self._session.post_no_content(
            f"{_BASE}/watermarks/unset",
            params={"channelId": channel_id},
        )
