"""SDK-wide constants and configuration defaults."""

# YouTube Data API v3
YOUTUBE_API_BASE_URL = "https://www.googleapis.com/youtube/v3"

# YouTube Data API v3 — upload endpoints
YOUTUBE_UPLOAD_BASE_URL = "https://www.googleapis.com/upload/youtube/v3"

# YouTube Analytics API v2
YOUTUBE_ANALYTICS_BASE_URL = "https://youtubeanalytics.googleapis.com/v2"

# Default network timeout in seconds for API requests.
DEFAULT_TIMEOUT: float = 30.0

# Default maximum results per page (YouTube's own default is 5).
DEFAULT_MAX_RESULTS: int = 5

# Default size (in bytes) for each video chunk during resumable upload.
# Google recommends chunks of at least 8 MB for reliable uploads.
DEFAULT_VIDEO_CHUNK_SIZE: int = 10 * 1024 * 1024  # 10 MB

# Maximum file size for custom thumbnail uploads (2 MB).
MAX_THUMBNAIL_SIZE: int = 2 * 1024 * 1024

# Maximum file size for channel banner uploads (6 MB).
MAX_BANNER_SIZE: int = 6 * 1024 * 1024
