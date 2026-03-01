"""ChannelBanners API namespace.

Wraps the following YouTube endpoint:

* POST /upload/.../channelBanners/insert — upload a channel banner image

Reference: https://developers.google.com/youtube/v3/docs/channelBanners/insert
"""

from __future__ import annotations

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_UPLOAD_BASE_URL
from youtube.models.channel_banners import ChannelBannerResource

_UPLOAD = YOUTUBE_UPLOAD_BASE_URL


class ChannelBannersAPI(BaseAPI):
    """Methods for uploading YouTube channel banner images."""

    async def insert(
        self,
        *,
        image_data: bytes,
        content_type: str = "image/jpeg",
    ) -> ChannelBannerResource:
        """Upload a channel banner image.

        Returns a :class:`ChannelBannerResource` whose ``url`` must be used
        in a subsequent ``channels.update`` call to set
        ``branding_settings.image.banner_external_url``.

        Parameters
        ----------
        image_data:
            Raw image bytes (JPEG or PNG, max 6 MB, 16:9 aspect ratio,
            recommended 2560 x 1440 px).
        content_type:
            MIME type — ``image/jpeg`` or ``image/png``.
        """
        payload = await self._session.upload_binary(
            f"{_UPLOAD}/channelBanners/insert",
            params={},
            data=image_data,
            content_type=content_type,
        )
        return ChannelBannerResource.model_validate(payload)

    async def insert_from_file(
        self,
        *,
        file_path: str,
        content_type: str = "image/jpeg",
    ) -> ChannelBannerResource:
        """Upload a channel banner from a local file."""
        with open(file_path, "rb") as fh:
            image_data = fh.read()
        return await self.insert(
            image_data=image_data,
            content_type=content_type,
        )
