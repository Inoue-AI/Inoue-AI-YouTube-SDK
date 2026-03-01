"""Pydantic models for the YouTube Analytics API.

Reference: https://developers.google.com/youtube/analytics/reference/reports/query
"""

from __future__ import annotations

from typing import Any

from youtube.models.base import YouTubeBaseModel


class ColumnHeader(YouTubeBaseModel):
    """Metadata for a single column in an analytics report."""

    name: str | None = None
    data_type: str | None = None
    column_type: str | None = None


class AnalyticsResponse(YouTubeBaseModel):
    """Response from ``reports.query``.

    ``rows`` is a list of lists where each inner list contains the values
    for one row, matching the order of ``column_headers``.  Values may be
    strings, integers, or floats depending on the metric.
    """

    kind: str | None = None
    column_headers: list[ColumnHeader] = []
    rows: list[list[Any]] = []
