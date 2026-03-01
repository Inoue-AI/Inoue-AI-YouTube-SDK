"""Exception hierarchy for the YouTube Python SDK."""

from __future__ import annotations


class YouTubeSDKError(Exception):
    """Base class for every exception raised by this SDK."""


# ---------------------------------------------------------------------------
# API-level errors (YouTube returned an error payload)
# ---------------------------------------------------------------------------

class YouTubeAPIError(YouTubeSDKError):
    """The YouTube API responded with an error.

    Attributes:
        code:        HTTP status code returned by the API.
        message:     Human-readable description from the API.
        reason:      Machine-readable error reason (e.g. ``"quotaExceeded"``).
        domain:      Error domain (e.g. ``"youtube.quota"``).
    """

    def __init__(
        self,
        code: int,
        message: str,
        reason: str = "",
        domain: str = "",
    ) -> None:
        self.code = code
        self.message = message
        self.reason = reason
        self.domain = domain
        super().__init__(f"[{code}] {message} (reason={reason})")


class YouTubeAuthError(YouTubeAPIError):
    """Raised when the access token is missing, invalid, or lacks the required scope."""


class YouTubeNotFoundError(YouTubeAPIError):
    """Raised when the requested resource does not exist."""


class YouTubeForbiddenError(YouTubeAPIError):
    """Raised when the request is understood but refused (insufficient permissions)."""


class YouTubeQuotaExceededError(YouTubeAPIError):
    """Raised when the daily API quota has been exceeded."""


class YouTubeRateLimitError(YouTubeAPIError):
    """Raised when too many requests have been sent in a given time window."""


class YouTubeConflictError(YouTubeAPIError):
    """Raised on conflict errors (e.g. duplicate resource)."""


class YouTubeBadRequestError(YouTubeAPIError):
    """Raised when the request is malformed or contains invalid parameters."""


class YouTubeServerError(YouTubeAPIError):
    """Raised when YouTube returns an internal server error (5xx)."""


# ---------------------------------------------------------------------------
# Client / transport-level errors
# ---------------------------------------------------------------------------

class YouTubeUploadError(YouTubeSDKError):
    """Raised when a file upload to YouTube fails."""

    def __init__(self, message: str, http_status: int | None = None) -> None:
        self.http_status = http_status
        super().__init__(message)


class YouTubeConfigError(YouTubeSDKError):
    """Raised when the SDK client is misconfigured (e.g. missing credentials)."""


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

_REASON_MAP: dict[str, type[YouTubeAPIError]] = {
    "unauthorized": YouTubeAuthError,
    "authError": YouTubeAuthError,
    "forbidden": YouTubeForbiddenError,
    "insufficientPermissions": YouTubeForbiddenError,
    "notFound": YouTubeNotFoundError,
    "quotaExceeded": YouTubeQuotaExceededError,
    "rateLimitExceeded": YouTubeRateLimitError,
    "userRateLimitExceeded": YouTubeRateLimitError,
    "conflict": YouTubeConflictError,
    "alreadyExists": YouTubeConflictError,
    "badRequest": YouTubeBadRequestError,
    "invalidParameter": YouTubeBadRequestError,
    "mediaBodyRequired": YouTubeBadRequestError,
    "backendError": YouTubeServerError,
    "internalError": YouTubeServerError,
}

_STATUS_MAP: dict[int, type[YouTubeAPIError]] = {
    400: YouTubeBadRequestError,
    401: YouTubeAuthError,
    403: YouTubeForbiddenError,
    404: YouTubeNotFoundError,
    409: YouTubeConflictError,
    429: YouTubeRateLimitError,
}


def build_api_error(
    code: int,
    message: str,
    reason: str = "",
    domain: str = "",
) -> YouTubeAPIError:
    """Return the most specific YouTubeAPIError subclass for the error."""
    exc_class = _REASON_MAP.get(reason)
    if exc_class is None:
        exc_class = _STATUS_MAP.get(code, YouTubeAPIError)
    if code >= 500 and exc_class is YouTubeAPIError:
        exc_class = YouTubeServerError
    return exc_class(code=code, message=message, reason=reason, domain=domain)
