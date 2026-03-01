"""Videos API namespace.

Wraps the following YouTube endpoints:

* GET    /videos           — list videos
* POST   /upload/.../videos — upload (insert) a video
* PUT    /videos           — update video metadata
* DELETE /videos           — delete a video
* POST   /videos/rate      — like / dislike / remove rating
* GET    /videos/getRating — retrieve the authenticated user's rating

Reference: https://developers.google.com/youtube/v3/docs/videos
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL, YOUTUBE_UPLOAD_BASE_URL
from youtube.models.videos import (
    Rating,
    Video,
    VideoGetRatingResponse,
    VideoListResponse,
    VideoPart,
)

_BASE = YOUTUBE_API_BASE_URL
_UPLOAD = YOUTUBE_UPLOAD_BASE_URL


class VideosAPI(BaseAPI):
    """Methods for managing YouTube videos."""

    # ------------------------------------------------------------------
    # List
    # ------------------------------------------------------------------

    async def list(
        self,
        *,
        parts: list[VideoPart],
        id: str | list[str] | None = None,
        chart: str | None = None,
        my_rating: str | None = None,
        max_results: int = 5,
        page_token: str | None = None,
        region_code: str | None = None,
        video_category_id: str | None = None,
        hl: str | None = None,
    ) -> VideoListResponse:
        """Return a list of videos matching the given filter.

        Exactly one of ``id``, ``chart``, or ``my_rating`` must be provided.
        """
        params: dict[str, Any] = {
            "part": _join(parts),
            "maxResults": max_results,
        }
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if chart is not None:
            params["chart"] = chart
        if my_rating is not None:
            params["myRating"] = my_rating
        if page_token is not None:
            params["pageToken"] = page_token
        if region_code is not None:
            params["regionCode"] = region_code
        if video_category_id is not None:
            params["videoCategoryId"] = video_category_id
        if hl is not None:
            params["hl"] = hl

        payload = await self._session.get(f"{_BASE}/videos", params=params)
        return VideoListResponse.model_validate(payload)

    async def iter_by_chart(
        self,
        *,
        parts: list[VideoPart],
        chart: str = "mostPopular",
        max_results: int = 5,
        region_code: str | None = None,
        video_category_id: str | None = None,
    ) -> AsyncIterator[Video]:
        """Async generator that pages through chart results."""
        page_token: str | None = None
        while True:
            page = await self.list(
                parts=parts,
                chart=chart,
                max_results=max_results,
                page_token=page_token,
                region_code=region_code,
                video_category_id=video_category_id,
            )
            for video in page.items:
                yield video
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    # ------------------------------------------------------------------
    # Insert (upload)
    # ------------------------------------------------------------------

    async def insert(
        self,
        body: Video,
        file_path: str,
        *,
        parts: list[VideoPart] | None = None,
        notify_subscribers: bool = True,
    ) -> Video:
        """Upload a video file to YouTube.

        The ``body`` should contain at least a ``snippet`` (with ``title``)
        and optionally ``status``, ``recording_details``, etc.

        Uses the resumable upload protocol.
        """
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        if not notify_subscribers:
            params["notifySubscribers"] = "false"

        metadata = body.model_dump(by_alias=True, exclude_none=True)

        upload_url = await self._session.init_resumable_upload(
            f"{_UPLOAD}/videos",
            params=params,
            metadata=metadata,
            file_path=file_path,
        )

        payload = await self._session.upload_file_resumable(
            upload_url,
            file_path,
        )
        return Video.model_validate(payload)

    # ------------------------------------------------------------------
    # Update
    # ------------------------------------------------------------------

    async def update(
        self,
        body: Video,
        *,
        parts: list[VideoPart] | None = None,
    ) -> Video:
        """Update a video's metadata.

        The ``body`` must contain ``id`` and at least one part to update.
        """
        if parts is None:
            parts = _infer_parts(body)

        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.put(
            f"{_BASE}/videos",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Video.model_validate(payload)

    # ------------------------------------------------------------------
    # Delete
    # ------------------------------------------------------------------

    async def delete(self, *, id: str) -> None:
        """Permanently delete a video."""
        await self._session.delete(f"{_BASE}/videos", params={"id": id})

    # ------------------------------------------------------------------
    # Rate
    # ------------------------------------------------------------------

    async def rate(self, *, id: str, rating: Rating) -> None:
        """Like, dislike, or remove the authenticated user's rating on a video."""
        await self._session.post_no_content(
            f"{_BASE}/videos/rate",
            params={"id": id, "rating": rating.value},
        )

    async def get_rating(
        self,
        *,
        id: str | list[str],
    ) -> VideoGetRatingResponse:
        """Retrieve the authenticated user's rating for one or more videos."""
        video_ids = ",".join(id) if isinstance(id, list) else id
        payload = await self._session.get(
            f"{_BASE}/videos/getRating",
            params={"id": video_ids},
        )
        return VideoGetRatingResponse.model_validate(payload)

    # ------------------------------------------------------------------
    # Report abuse
    # ------------------------------------------------------------------

    async def report_abuse(
        self,
        *,
        video_id: str,
        reason_id: str,
        secondary_reason_id: str | None = None,
        comments: str | None = None,
        language: str | None = None,
    ) -> None:
        """Report a video for abusive content."""
        body: dict[str, Any] = {
            "videoId": video_id,
            "reasonId": reason_id,
        }
        if secondary_reason_id is not None:
            body["secondaryReasonId"] = secondary_reason_id
        if comments is not None:
            body["comments"] = comments
        if language is not None:
            body["language"] = language
        await self._session.post(f"{_BASE}/videos/reportAbuse", json=body)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _join(parts: list[VideoPart]) -> str:
    return ",".join(p.value for p in parts)


def _infer_parts(body: Video) -> list[VideoPart]:
    """Detect which resource parts are populated on a Video model."""
    parts: list[VideoPart] = []
    if body.snippet is not None:
        parts.append(VideoPart.SNIPPET)
    if body.status is not None:
        parts.append(VideoPart.STATUS)
    if body.recording_details is not None:
        parts.append(VideoPart.RECORDING_DETAILS)
    if body.localizations is not None:
        parts.append(VideoPart.LOCALIZATIONS)
    return parts or [VideoPart.SNIPPET]
