"""Pydantic models for the YouTube Videos resource.

Reference: https://developers.google.com/youtube/v3/docs/videos
"""

from __future__ import annotations

from enum import StrEnum

from youtube.models.base import YouTubeBaseModel
from youtube.models.common import (
    Localization,
    PageInfo,
    PrivacyStatus,
    ThumbnailDetails,
)

# ---------------------------------------------------------------------------
# Enumerations
# ---------------------------------------------------------------------------

class VideoPart(StrEnum):
    """Resource parts accepted by the ``part`` parameter."""

    ID = "id"
    SNIPPET = "snippet"
    CONTENT_DETAILS = "contentDetails"
    STATUS = "status"
    STATISTICS = "statistics"
    PLAYER = "player"
    TOPIC_DETAILS = "topicDetails"
    RECORDING_DETAILS = "recordingDetails"
    FILE_DETAILS = "fileDetails"
    PROCESSING_DETAILS = "processingDetails"
    SUGGESTIONS = "suggestions"
    LIVE_STREAMING_DETAILS = "liveStreamingDetails"
    LOCALIZATIONS = "localizations"


class UploadStatus(StrEnum):
    """Processing status of an uploaded video."""

    DELETED = "deleted"
    FAILED = "failed"
    PROCESSED = "processed"
    REJECTED = "rejected"
    UPLOADED = "uploaded"


class VideoLicense(StrEnum):
    """License type for a video."""

    CREATIVE_COMMON = "creativeCommon"
    YOUTUBE = "youtube"


class Rating(StrEnum):
    """Viewer rating value for ``videos.rate`` / ``videos.getRating``."""

    LIKE = "like"
    DISLIKE = "dislike"
    NONE = "none"


# ---------------------------------------------------------------------------
# Nested resource parts
# ---------------------------------------------------------------------------

class VideoSnippet(YouTubeBaseModel):
    """The ``snippet`` part of a Video resource."""

    published_at: str | None = None
    channel_id: str | None = None
    title: str | None = None
    description: str | None = None
    thumbnails: ThumbnailDetails | None = None
    channel_title: str | None = None
    tags: list[str] | None = None
    category_id: str | None = None
    live_broadcast_content: str | None = None
    default_language: str | None = None
    localized: Localization | None = None
    default_audio_language: str | None = None


class VideoContentDetails(YouTubeBaseModel):
    """The ``contentDetails`` part of a Video resource."""

    duration: str | None = None
    dimension: str | None = None
    definition: str | None = None
    caption: str | None = None
    licensed_content: bool | None = None
    projection: str | None = None


class VideoStatus(YouTubeBaseModel):
    """The ``status`` part of a Video resource."""

    upload_status: UploadStatus | None = None
    failure_reason: str | None = None
    rejection_reason: str | None = None
    privacy_status: PrivacyStatus | None = None
    publish_at: str | None = None
    license: VideoLicense | None = None
    embeddable: bool | None = None
    public_stats_viewable: bool | None = None
    made_for_kids: bool | None = None
    self_declared_made_for_kids: bool | None = None


class VideoStatistics(YouTubeBaseModel):
    """The ``statistics`` part of a Video resource.

    Note: YouTube returns all counts as **strings** to handle large numbers.
    """

    view_count: str | None = None
    like_count: str | None = None
    dislike_count: str | None = None
    favorite_count: str | None = None
    comment_count: str | None = None


class VideoPlayer(YouTubeBaseModel):
    """The ``player`` part of a Video resource."""

    embed_html: str | None = None
    embed_height: int | None = None
    embed_width: int | None = None


class VideoTopicDetails(YouTubeBaseModel):
    """The ``topicDetails`` part of a Video resource."""

    topic_ids: list[str] | None = None
    relevant_topic_ids: list[str] | None = None
    topic_categories: list[str] | None = None


class VideoRecordingDetails(YouTubeBaseModel):
    """The ``recordingDetails`` part of a Video resource."""

    recording_date: str | None = None


class VideoLiveStreamingDetails(YouTubeBaseModel):
    """The ``liveStreamingDetails`` part of a Video resource."""

    actual_start_time: str | None = None
    actual_end_time: str | None = None
    scheduled_start_time: str | None = None
    scheduled_end_time: str | None = None
    concurrent_viewers: str | None = None
    active_live_chat_id: str | None = None


# ---------------------------------------------------------------------------
# Top-level resource & list response
# ---------------------------------------------------------------------------

class Video(YouTubeBaseModel):
    """A YouTube Video resource.

    Used as both the response model *and* the request body for
    ``videos.insert`` / ``videos.update``.
    """

    kind: str | None = None
    etag: str | None = None
    id: str | None = None
    snippet: VideoSnippet | None = None
    content_details: VideoContentDetails | None = None
    status: VideoStatus | None = None
    statistics: VideoStatistics | None = None
    player: VideoPlayer | None = None
    topic_details: VideoTopicDetails | None = None
    recording_details: VideoRecordingDetails | None = None
    live_streaming_details: VideoLiveStreamingDetails | None = None
    localizations: dict[str, Localization] | None = None


class VideoListResponse(YouTubeBaseModel):
    """Response from ``videos.list``."""

    kind: str | None = None
    etag: str | None = None
    next_page_token: str | None = None
    prev_page_token: str | None = None
    page_info: PageInfo | None = None
    items: list[Video] = []


class VideoRating(YouTubeBaseModel):
    """A single item from the ``videos.getRating`` response."""

    video_id: str | None = None
    rating: Rating | None = None


class VideoGetRatingResponse(YouTubeBaseModel):
    """Response from ``videos.getRating``."""

    kind: str | None = None
    etag: str | None = None
    items: list[VideoRating] = []
