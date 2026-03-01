"""Search API namespace.

Wraps the following YouTube endpoint:

* GET /search — search for videos, channels, and playlists

Reference: https://developers.google.com/youtube/v3/docs/search/list
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.search import (
    EventType,
    SafeSearch,
    SearchListResponse,
    SearchOrder,
    SearchResult,
    SearchResultType,
    VideoCaption,
    VideoDefinition,
    VideoDimension,
    VideoDuration,
    VideoType,
)

_BASE = YOUTUBE_API_BASE_URL


class SearchAPI(BaseAPI):
    """Methods for searching YouTube."""

    async def list(
        self,
        *,
        q: str | None = None,
        type: list[SearchResultType] | SearchResultType | None = None,
        channel_id: str | None = None,
        for_mine: bool | None = None,
        order: SearchOrder | None = None,
        max_results: int = 5,
        page_token: str | None = None,
        published_after: str | None = None,
        published_before: str | None = None,
        region_code: str | None = None,
        relevance_language: str | None = None,
        safe_search: SafeSearch | None = None,
        event_type: EventType | None = None,
        video_caption: VideoCaption | None = None,
        video_category_id: str | None = None,
        video_definition: VideoDefinition | None = None,
        video_dimension: VideoDimension | None = None,
        video_duration: VideoDuration | None = None,
        video_embeddable: bool | None = None,
        video_license: str | None = None,
        video_type: VideoType | None = None,
    ) -> SearchListResponse:
        """Search for YouTube resources.

        Note: each call costs **100 quota units**.
        """
        params: dict[str, Any] = {
            "part": "snippet",
            "maxResults": max_results,
        }
        if q is not None:
            params["q"] = q
        if type is not None:
            if isinstance(type, list):
                params["type"] = ",".join(t.value for t in type)
            else:
                params["type"] = type.value
        if channel_id is not None:
            params["channelId"] = channel_id
        if for_mine is not None:
            params["forMine"] = str(for_mine).lower()
        if order is not None:
            params["order"] = order.value
        if page_token is not None:
            params["pageToken"] = page_token
        if published_after is not None:
            params["publishedAfter"] = published_after
        if published_before is not None:
            params["publishedBefore"] = published_before
        if region_code is not None:
            params["regionCode"] = region_code
        if relevance_language is not None:
            params["relevanceLanguage"] = relevance_language
        if safe_search is not None:
            params["safeSearch"] = safe_search.value
        if event_type is not None:
            params["eventType"] = event_type.value
        if video_caption is not None:
            params["videoCaption"] = video_caption.value
        if video_category_id is not None:
            params["videoCategoryId"] = video_category_id
        if video_definition is not None:
            params["videoDefinition"] = video_definition.value
        if video_dimension is not None:
            params["videoDimension"] = video_dimension.value
        if video_duration is not None:
            params["videoDuration"] = video_duration.value
        if video_embeddable is not None:
            params["videoEmbeddable"] = str(video_embeddable).lower()
        if video_license is not None:
            params["videoLicense"] = video_license
        if video_type is not None:
            params["videoType"] = video_type.value

        payload = await self._session.get(f"{_BASE}/search", params=params)
        return SearchListResponse.model_validate(payload)

    async def iter_results(
        self,
        *,
        q: str | None = None,
        type: list[SearchResultType] | SearchResultType | None = None,
        channel_id: str | None = None,
        order: SearchOrder | None = None,
        max_results: int = 25,
        published_after: str | None = None,
        published_before: str | None = None,
        max_pages: int = 5,
    ) -> AsyncIterator[SearchResult]:
        """Async generator that pages through search results.

        ``max_pages`` limits the total number of API calls to avoid excessive
        quota usage (each page = 100 quota units).
        """
        page_token: str | None = None
        for _ in range(max_pages):
            page = await self.list(
                q=q,
                type=type,
                channel_id=channel_id,
                order=order,
                max_results=max_results,
                page_token=page_token,
                published_after=published_after,
                published_before=published_before,
            )
            for result in page.items:
                yield result
            if not page.next_page_token:
                break
            page_token = page.next_page_token
