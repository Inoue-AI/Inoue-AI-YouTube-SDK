"""Thumbnails API namespace.

Wraps the following YouTube endpoint:

* POST /upload/.../thumbnails/set — upload a custom thumbnail

Reference: https://developers.google.com/youtube/v3/docs/thumbnails/set
"""

from __future__ import annotations

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_UPLOAD_BASE_URL
from youtube.models.thumbnails import ThumbnailSetResponse

_UPLOAD = YOUTUBE_UPLOAD_BASE_URL


class ThumbnailsAPI(BaseAPI):
    """Methods for managing YouTube video thumbnails."""

    async def set(
        self,
        *,
        video_id: str,
        image_data: bytes,
        content_type: str = "image/jpeg",
    ) -> ThumbnailSetResponse:
        """Upload a custom thumbnail for a video.

        Parameters
        ----------
        video_id:
            The video to set the thumbnail for.
        image_data:
            Raw image bytes (JPEG or PNG, max 2 MB).
        content_type:
            MIME type — ``image/jpeg`` or ``image/png``.
        """
        payload = await self._session.upload_binary(
            f"{_UPLOAD}/thumbnails/set",
            params={"videoId": video_id},
            data=image_data,
            content_type=content_type,
        )
        return ThumbnailSetResponse.model_validate(payload)

    async def set_from_file(
        self,
        *,
        video_id: str,
        file_path: str,
        content_type: str = "image/jpeg",
    ) -> ThumbnailSetResponse:
        """Upload a custom thumbnail from a local file.

        Parameters
        ----------
        video_id:
            The video to set the thumbnail for.
        file_path:
            Path to the image file (JPEG or PNG, max 2 MB).
        content_type:
            MIME type — ``image/jpeg`` or ``image/png``.
        """
        with open(file_path, "rb") as fh:
            image_data = fh.read()
        return await self.set(
            video_id=video_id,
            image_data=image_data,
            content_type=content_type,
        )
