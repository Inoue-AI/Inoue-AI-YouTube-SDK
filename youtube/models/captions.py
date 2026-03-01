"""Pydantic models for the YouTube Captions resource.

Reference: https://developers.google.com/youtube/v3/docs/captions
"""

from __future__ import annotations

from enum import StrEnum

from pydantic import Field

from youtube.models.base import YouTubeBaseModel

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class CaptionPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"


class TrackKind(StrEnum):
    """Kind of caption track."""

    ASR = "ASR"
    FORCED = "forced"
    STANDARD = "standard"


class AudioTrackType(StrEnum):
    """Type of audio track the caption is associated with."""

    COMMENTARY = "commentary"
    DESCRIPTIVE = "descriptive"
    PRIMARY = "primary"
    UNKNOWN = "unknown"


class CaptionFormat(StrEnum):
    """Download format for caption tracks."""

    SBV = "sbv"
    SCC = "scc"
    SRT = "srt"
    TTML = "ttml"
    VTT = "vtt"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class CaptionSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Caption resource."""

    video_id: str | None = None
    last_updated: str | None = None
    track_kind: TrackKind | None = None
    language: str | None = None
    name: str | None = None
    audio_track_type: AudioTrackType | None = None
    is_cc: bool | None = Field(None, alias="isCC")
    is_draft: bool | None = None
    is_auto_synced: bool | None = None
    status: str | None = None
    failure_reason: str | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class Caption(YouTubeBaseModel):
    """A YouTube Caption resource."""

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: CaptionSnippet | None = None


class CaptionListResponse(YouTubeBaseModel):
    """Response from ``captions.list``."""

    kind: str | None = None
    etag: str | None = None
    items: list[Caption] = []
