"""YouTube Analytics API namespace.

Wraps the following endpoint:

* GET /reports — query analytics reports

Reference: https://developers.google.com/youtube/analytics/reference/reports/query
"""

from __future__ import annotations

from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_ANALYTICS_BASE_URL
from youtube.models.analytics import AnalyticsResponse

_BASE = YOUTUBE_ANALYTICS_BASE_URL


class AnalyticsAPI(BaseAPI):
    """Methods for querying YouTube Analytics reports."""

    async def query(
        self,
        *,
        ids: str = "channel==MINE",
        start_date: str,
        end_date: str,
        metrics: str | list[str],
        dimensions: str | list[str] | None = None,
        filters: str | None = None,
        sort: str | list[str] | None = None,
        max_results: int | None = None,
        start_index: int | None = None,
        currency: str | None = None,
        include_historical_channel_data: bool | None = None,
    ) -> AnalyticsResponse:
        """Query a YouTube Analytics report.

        Parameters
        ----------
        ids:
            Channel or content owner to report on.
            Default ``"channel==MINE"`` uses the authenticated user's channel.
        start_date:
            Start date in ``YYYY-MM-DD`` format.
        end_date:
            End date in ``YYYY-MM-DD`` format.
        metrics:
            Comma-separated metrics or a list, e.g.
            ``["views", "likes", "estimatedMinutesWatched"]``.
        dimensions:
            Comma-separated dimensions or a list, e.g. ``["day"]``.
        filters:
            Semicolon-separated filters, e.g.
            ``"video==VIDEO_ID;country==US"``.
        sort:
            Comma-separated sort fields or a list.  Prefix ``-`` for
            descending.
        max_results:
            Maximum number of rows to return.
        start_index:
            1-based index of the first row to return.
        currency:
            ISO 4217 currency code for revenue metrics (default ``USD``).
        include_historical_channel_data:
            Include data from before the channel was linked.
        """
        params: dict[str, Any] = {
            "ids": ids,
            "startDate": start_date,
            "endDate": end_date,
            "metrics": ",".join(metrics) if isinstance(metrics, list) else metrics,
        }
        if dimensions is not None:
            params["dimensions"] = (
                ",".join(dimensions) if isinstance(dimensions, list) else dimensions
            )
        if filters is not None:
            params["filters"] = filters
        if sort is not None:
            params["sort"] = ",".join(sort) if isinstance(sort, list) else sort
        if max_results is not None:
            params["maxResults"] = max_results
        if start_index is not None:
            params["startIndex"] = start_index
        if currency is not None:
            params["currency"] = currency
        if include_historical_channel_data is not None:
            params["includeHistoricalChannelData"] = str(include_historical_channel_data).lower()

        payload = await self._session.get(f"{_BASE}/reports", params=params)
        return AnalyticsResponse.model_validate(payload)
