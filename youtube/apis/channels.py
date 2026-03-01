"""Channels API namespace.

Wraps the following YouTube endpoints:

* GET /channels — list channels
* PUT /channels — update channel metadata

Reference: https://developers.google.com/youtube/v3/docs/channels
"""

from __future__ import annotations

from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.channels import (
    Channel,
    ChannelListResponse,
    ChannelPart,
)

_BASE = YOUTUBE_API_BASE_URL


class ChannelsAPI(BaseAPI):
    """Methods for reading and updating YouTube channels."""

    async def list(
        self,
        *,
        parts: list[ChannelPart],
        id: str | list[str] | None = None,
        for_handle: str | None = None,
        for_username: str | None = None,
        mine: bool | None = None,
        managed_by_me: bool | None = None,
        max_results: int = 5,
        page_token: str | None = None,
        hl: str | None = None,
    ) -> ChannelListResponse:
        """Return a list of channels matching the given filter.

        Exactly one of ``id``, ``for_handle``, ``for_username``, ``mine``,
        or ``managed_by_me`` must be provided.
        """
        params: dict[str, Any] = {
            "part": _join(parts),
            "maxResults": max_results,
        }
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if for_handle is not None:
            params["forHandle"] = for_handle
        if for_username is not None:
            params["forUsername"] = for_username
        if mine is not None:
            params["mine"] = str(mine).lower()
        if managed_by_me is not None:
            params["managedByMe"] = str(managed_by_me).lower()
        if page_token is not None:
            params["pageToken"] = page_token
        if hl is not None:
            params["hl"] = hl

        payload = await self._session.get(f"{_BASE}/channels", params=params)
        return ChannelListResponse.model_validate(payload)

    async def update(
        self,
        body: Channel,
        *,
        parts: list[ChannelPart] | None = None,
    ) -> Channel:
        """Update a channel's metadata.

        The ``body`` must contain ``id``.  Only ``brandingSettings``,
        ``localizations``, and ``status.selfDeclaredMadeForKids`` are
        writable.

        **Important:** properties not included in the request body are
        deleted — existing values are NOT preserved.
        """
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.put(
            f"{_BASE}/channels",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Channel.model_validate(payload)


def _join(parts: list[ChannelPart]) -> str:
    return ",".join(p.value for p in parts)


def _infer_parts(body: Channel) -> list[ChannelPart]:
    parts: list[ChannelPart] = []
    if body.branding_settings is not None:
        parts.append(ChannelPart.BRANDING_SETTINGS)
    if body.localizations is not None:
        parts.append(ChannelPart.LOCALIZATIONS)
    if body.status is not None:
        parts.append(ChannelPart.STATUS)
    return parts or [ChannelPart.BRANDING_SETTINGS]
