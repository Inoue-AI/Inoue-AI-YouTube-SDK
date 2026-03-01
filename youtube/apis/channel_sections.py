"""ChannelSections API namespace.

Wraps the following YouTube endpoints:

* GET    /channelSections — list channel sections
* POST   /channelSections — create a channel section
* PUT    /channelSections — update a channel section
* DELETE /channelSections — delete a channel section

Reference: https://developers.google.com/youtube/v3/docs/channelSections
"""

from __future__ import annotations

from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.channel_sections import (
    ChannelSection,
    ChannelSectionListResponse,
    ChannelSectionPart,
)

_BASE = YOUTUBE_API_BASE_URL


class ChannelSectionsAPI(BaseAPI):
    """Methods for managing YouTube channel sections."""

    async def list(
        self,
        *,
        parts: list[ChannelSectionPart],
        channel_id: str | None = None,
        id: str | list[str] | None = None,
        mine: bool | None = None,
    ) -> ChannelSectionListResponse:
        """Return channel sections matching the given filter."""
        params: dict[str, Any] = {"part": _join(parts)}
        if channel_id is not None:
            params["channelId"] = channel_id
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if mine is not None:
            params["mine"] = str(mine).lower()

        payload = await self._session.get(f"{_BASE}/channelSections", params=params)
        return ChannelSectionListResponse.model_validate(payload)

    async def insert(
        self,
        body: ChannelSection,
        *,
        parts: list[ChannelSectionPart] | None = None,
    ) -> ChannelSection:
        """Create a new channel section (max 10 per channel)."""
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.post(
            f"{_BASE}/channelSections",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return ChannelSection.model_validate(payload)

    async def update(
        self,
        body: ChannelSection,
        *,
        parts: list[ChannelSectionPart] | None = None,
    ) -> ChannelSection:
        """Update an existing channel section."""
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.put(
            f"{_BASE}/channelSections",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return ChannelSection.model_validate(payload)

    async def delete(self, *, id: str) -> None:
        """Delete a channel section."""
        await self._session.delete(f"{_BASE}/channelSections", params={"id": id})


def _join(parts: list[ChannelSectionPart]) -> str:
    return ",".join(p.value for p in parts)


def _infer_parts(body: ChannelSection) -> list[ChannelSectionPart]:
    parts: list[ChannelSectionPart] = []
    if body.snippet is not None:
        parts.append(ChannelSectionPart.SNIPPET)
    if body.content_details is not None:
        parts.append(ChannelSectionPart.CONTENT_DETAILS)
    return parts or [ChannelSectionPart.SNIPPET]
