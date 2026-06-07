"""OAuth 2.0 token-refresh helper for the YouTube SDK.

The SDK consumes pre-obtained OAuth access tokens; it does **not** drive the
interactive authorization-code redirect flow (that belongs to the frontend).
A long-running backend, however, must be able to exchange a stored
*refresh token* for a fresh *access token* when the previous one expires.
That single, server-side operation lives here.

The client talks to Google's OAuth 2.0 token endpoint
(https://oauth2.googleapis.com/token) and returns a typed
:class:`TokenResponse`. Credentials are accepted as explicit arguments (which
callers typically source from environment variables) and are never logged.

Example::

    from youtube.oauth import OAuthClient

    async with OAuthClient(
        client_id=os.environ["YT_OAUTH_CLIENT_ID"],
        client_secret=os.environ["YT_OAUTH_CLIENT_SECRET"],
    ) as oauth:
        token = await oauth.refresh_access_token(os.environ["YT_REFRESH_TOKEN"])
        # token.access_token is now ready to pass to YouTubeClient(...)
"""

from __future__ import annotations

import asyncio
from types import TracebackType
from typing import Any, Self

import aiohttp
from pydantic import Field

from youtube.config import DEFAULT_TIMEOUT
from youtube.exceptions import YouTubeAuthError, YouTubeConfigError
from youtube.models.base import YouTubeBaseModel

#: Google's OAuth 2.0 token endpoint.
GOOGLE_TOKEN_ENDPOINT = "https://oauth2.googleapis.com/token"


class TokenResponse(YouTubeBaseModel):
    """A successful response from the OAuth 2.0 token endpoint.

    Attributes
    ----------
    access_token:
        The freshly minted OAuth 2.0 access token.
    expires_in:
        Lifetime of ``access_token`` in seconds.
    scope:
        Space-delimited scopes granted to the token (may be omitted).
    token_type:
        Always ``"Bearer"`` for YouTube.
    refresh_token:
        A new refresh token. Google only returns this in some flows; on a
        plain refresh it is usually absent, so the original refresh token
        remains valid.
    """

    access_token: str
    expires_in: int
    scope: str | None = None
    token_type: str = "Bearer"
    refresh_token: str | None = None


class _TokenErrorResponse(YouTubeBaseModel):
    """The OAuth error envelope returned on a failed token exchange."""

    error: str = Field(default="invalid_request")
    error_description: str | None = None


class OAuthClient:
    """Async client for the server-side OAuth 2.0 token-refresh exchange.

    Parameters
    ----------
    client_id:
        The OAuth 2.0 client identifier issued by Google Cloud.
    client_secret:
        The OAuth 2.0 client secret. Pass it from an environment variable;
        never hardcode it.
    token_endpoint:
        Override the token URL (used by tests). Defaults to Google's endpoint.
    timeout:
        Total request timeout in seconds.
    """

    def __init__(
        self,
        *,
        client_id: str,
        client_secret: str,
        token_endpoint: str = GOOGLE_TOKEN_ENDPOINT,
        timeout: float = DEFAULT_TIMEOUT,
    ) -> None:
        if not client_id or not client_secret:
            raise YouTubeConfigError("OAuthClient requires both client_id and client_secret.")
        self._client_id = client_id
        self._client_secret = client_secret
        self._token_endpoint = token_endpoint
        self._timeout = aiohttp.ClientTimeout(total=timeout)
        self._session: aiohttp.ClientSession | None = None
        self._lock = asyncio.Lock()

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    async def __aenter__(self) -> Self:
        return self

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None,
        exc_val: BaseException | None,
        exc_tb: TracebackType | None,
    ) -> None:
        await self.aclose()

    async def _get_session(self) -> aiohttp.ClientSession:
        """Get or lazily create the underlying aiohttp session (thread-safe)."""
        async with self._lock:
            if self._session is None or self._session.closed:
                connector = aiohttp.TCPConnector(limit=10, ttl_dns_cache=300, ssl=True)
                self._session = aiohttp.ClientSession(
                    connector=connector,
                    timeout=self._timeout,
                )
            return self._session

    async def aclose(self) -> None:
        """Close the underlying HTTP session and release all connections."""
        async with self._lock:
            if self._session is not None and not self._session.closed:
                await self._session.close()
                self._session = None

    # ------------------------------------------------------------------
    # Token operations
    # ------------------------------------------------------------------

    async def refresh_access_token(self, refresh_token: str) -> TokenResponse:
        """Exchange a refresh token for a new access token.

        Parameters
        ----------
        refresh_token:
            The stored OAuth 2.0 refresh token for the user/channel.

        Returns
        -------
        TokenResponse
            The new access token and its metadata.

        Raises
        ------
        YouTubeConfigError
            If ``refresh_token`` is empty.
        YouTubeAuthError
            If Google rejects the refresh (e.g. revoked or invalid grant).
        """
        if not refresh_token:
            raise YouTubeConfigError("refresh_access_token requires a refresh_token.")

        form: dict[str, str] = {
            "client_id": self._client_id,
            "client_secret": self._client_secret,
            "refresh_token": refresh_token,
            "grant_type": "refresh_token",
        }
        return await self._post_token(form)

    async def exchange_authorization_code(
        self,
        code: str,
        redirect_uri: str,
        *,
        code_verifier: str | None = None,
    ) -> TokenResponse:
        """Exchange an authorization code for access + refresh tokens.

        Provided for completeness so a backend that receives the redirect
        callback can complete the flow. The interactive consent step that
        produces ``code`` is the frontend's responsibility.

        Parameters
        ----------
        code:
            The authorization code returned to ``redirect_uri``.
        redirect_uri:
            The exact redirect URI registered with the OAuth client.
        code_verifier:
            The PKCE code verifier, if the auth request used PKCE.
        """
        if not code:
            raise YouTubeConfigError("exchange_authorization_code requires a code.")
        if not redirect_uri:
            raise YouTubeConfigError("exchange_authorization_code requires a redirect_uri.")

        form: dict[str, str] = {
            "client_id": self._client_id,
            "client_secret": self._client_secret,
            "code": code,
            "redirect_uri": redirect_uri,
            "grant_type": "authorization_code",
        }
        if code_verifier is not None:
            form["code_verifier"] = code_verifier
        return await self._post_token(form)

    # ------------------------------------------------------------------
    # Internal
    # ------------------------------------------------------------------

    async def _post_token(self, form: dict[str, str]) -> TokenResponse:
        """POST a form-encoded token request and parse the response."""
        session = await self._get_session()
        async with session.post(self._token_endpoint, data=form) as resp:
            payload: dict[str, Any] = await resp.json(content_type=None)
            if resp.status != 200 or "error" in payload:
                err = _TokenErrorResponse.model_validate(payload)
                raise YouTubeAuthError(
                    code=resp.status,
                    message=err.error_description or err.error,
                    reason=err.error,
                    domain="oauth2",
                )
            return TokenResponse.model_validate(payload)
