"""Async HTTP session that wraps aiohttp for all YouTube API communication."""

from __future__ import annotations

import asyncio
import os
from typing import Any

import aiohttp

from youtube.config import DEFAULT_TIMEOUT, DEFAULT_VIDEO_CHUNK_SIZE
from youtube.exceptions import YouTubeUploadError, build_api_error


class YouTubeSession:
    """Thin async HTTP client bound to an access token and/or API key.

    The underlying :class:`aiohttp.ClientSession` is created lazily on the
    first request.

    Parameters
    ----------
    access_token:
        A valid OAuth 2.0 access token.  Required for any write operation
        and for reading private data.
    api_key:
        A Google API key.  Sufficient for read-only access to public data.
    timeout:
        Total request timeout in seconds (default: ``DEFAULT_TIMEOUT``).
    """

    def __init__(
        self,
        *,
        access_token: str | None = None,
        api_key: str | None = None,
        timeout: float = DEFAULT_TIMEOUT,
    ) -> None:
        self._access_token = access_token
        self._api_key = api_key
        self._timeout = aiohttp.ClientTimeout(total=timeout)
        self._session: aiohttp.ClientSession | None = None
        self._lock = asyncio.Lock()

    # ------------------------------------------------------------------
    # Session lifecycle
    # ------------------------------------------------------------------

    async def _get_session(self) -> aiohttp.ClientSession:
        """Get or create the underlying aiohttp session (thread-safe)."""
        async with self._lock:
            if self._session is None or self._session.closed:
                headers: dict[str, str] = {"Accept": "application/json"}
                if self._access_token:
                    headers["Authorization"] = f"Bearer {self._access_token}"
                connector = aiohttp.TCPConnector(
                    limit=100,
                    ttl_dns_cache=300,
                    ssl=True,
                )
                self._session = aiohttp.ClientSession(
                    connector=connector,
                    timeout=self._timeout,
                    headers=headers,
                )
        return self._session

    async def aclose(self) -> None:
        """Close the underlying HTTP session and release all connections."""
        async with self._lock:
            if self._session is not None and not self._session.closed:
                await self._session.close()
                self._session = None

    # ------------------------------------------------------------------
    # Auth helpers
    # ------------------------------------------------------------------

    def _inject_key(self, params: dict[str, Any] | None) -> dict[str, Any] | None:
        """Append ``key=<api_key>`` to query params when an API key is set."""
        if self._api_key:
            params = dict(params) if params else {}
            params["key"] = self._api_key
        return params

    # ------------------------------------------------------------------
    # Public request methods
    # ------------------------------------------------------------------

    async def get(
        self,
        url: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a GET request and return the parsed JSON body."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.get(url, params=params) as resp:
            return await self._parse_json(resp)

    async def post(
        self,
        url: str,
        *,
        json: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a POST request with an optional JSON body."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.post(url, json=json, params=params) as resp:
            return await self._parse_json(resp)

    async def put(
        self,
        url: str,
        *,
        json: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a PUT request with an optional JSON body."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.put(url, json=json, params=params) as resp:
            return await self._parse_json(resp)

    async def delete(
        self,
        url: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> None:
        """Issue a DELETE request (expects 204 No Content)."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.delete(url, params=params) as resp:
            if resp.status == 204:
                return
            await self._raise_for_error(resp)

    async def post_no_content(
        self,
        url: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> None:
        """Issue a POST request that returns 204 No Content (e.g. rate, setModerationStatus)."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.post(url, params=params) as resp:
            if resp.status == 204:
                return
            await self._raise_for_error(resp)

    # ------------------------------------------------------------------
    # Upload methods
    # ------------------------------------------------------------------

    async def init_resumable_upload(
        self,
        url: str,
        *,
        params: dict[str, Any],
        metadata: dict[str, Any],
        file_path: str,
        content_type: str = "video/*",
    ) -> str:
        """Initiate a resumable upload and return the upload URI.

        Reference: https://developers.google.com/youtube/v3/guides/using_code_samples#uploads
        """
        params = dict(params)
        params["uploadType"] = "resumable"
        params = self._inject_key(params)

        file_size = os.path.getsize(file_path)
        session = await self._get_session()
        headers = {
            "Content-Type": "application/json; charset=UTF-8",
            "X-Upload-Content-Length": str(file_size),
            "X-Upload-Content-Type": content_type,
        }

        async with session.post(url, json=metadata, params=params, headers=headers) as resp:
            if resp.status != 200:
                await self._raise_for_error(resp)
            upload_url = resp.headers.get("Location")
            if not upload_url:
                raise YouTubeUploadError("No Location header in resumable upload init response.")
            return upload_url

    async def upload_file_resumable(
        self,
        upload_url: str,
        file_path: str,
        *,
        content_type: str = "video/*",
        chunk_size: int = DEFAULT_VIDEO_CHUNK_SIZE,
    ) -> dict[str, Any]:
        """Upload a file in chunks via the resumable upload protocol.

        Returns the final response JSON (the created resource).
        """
        file_size = os.path.getsize(file_path)
        session = await self._get_session()

        with open(file_path, "rb") as fh:
            start = 0
            while start < file_size:
                chunk = fh.read(chunk_size)
                end = start + len(chunk) - 1
                headers = {
                    "Content-Type": content_type,
                    "Content-Length": str(len(chunk)),
                    "Content-Range": f"bytes {start}-{end}/{file_size}",
                }
                async with session.put(upload_url, data=chunk, headers=headers) as resp:
                    if resp.status in (200, 201):
                        return await self._parse_json(resp)
                    if resp.status == 308:
                        # Chunk accepted, continue uploading.
                        start = end + 1
                        continue
                    raise YouTubeUploadError(
                        f"Chunk upload failed — HTTP {resp.status}",
                        http_status=resp.status,
                    )

        raise YouTubeUploadError("Upload finished without receiving a final response.")

    async def upload_binary(
        self,
        url: str,
        *,
        params: dict[str, Any],
        data: bytes,
        content_type: str,
    ) -> dict[str, Any]:
        """Direct binary upload (thumbnails, channel banners, watermarks).

        Returns the parsed JSON response.
        """
        params = dict(params)
        params["uploadType"] = "media"
        params = self._inject_key(params)

        session = await self._get_session()
        headers = {
            "Content-Type": content_type,
            "Content-Length": str(len(data)),
        }
        async with session.post(url, data=data, params=params, headers=headers) as resp:
            if resp.status in (200, 201):
                return await self._parse_json(resp)
            if resp.status == 204:
                return {}
            await self._raise_for_error(resp)
            return {}  # unreachable, _raise_for_error always raises

    async def upload_multipart(
        self,
        url: str,
        *,
        params: dict[str, Any],
        metadata: dict[str, Any],
        data: bytes,
        content_type: str,
    ) -> dict[str, Any]:
        """Multipart upload (captions): JSON metadata + binary file.

        Returns the parsed JSON response.
        """

        params = dict(params)
        params["uploadType"] = "multipart"
        params = self._inject_key(params)

        with aiohttp.MultipartWriter("related") as mpwriter:
            json_part = mpwriter.append_json(metadata)
            json_part.set_content_disposition("form-data", name="metadata")

            file_part = mpwriter.append(data, {"Content-Type": content_type})
            file_part.set_content_disposition("form-data", name="file")

        session = await self._get_session()
        async with session.post(url, data=mpwriter, params=params) as resp:
            if resp.status in (200, 201):
                return await self._parse_json(resp)
            await self._raise_for_error(resp)
            return {}

    async def get_binary(
        self,
        url: str,
        *,
        params: dict[str, Any] | None = None,
    ) -> bytes:
        """Issue a GET request and return the raw binary response body."""
        params = self._inject_key(params)
        session = await self._get_session()
        async with session.get(url, params=params) as resp:
            if resp.status == 200:
                return await resp.read()
            await self._raise_for_error(resp)
            return b""

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    @staticmethod
    async def _parse_json(response: aiohttp.ClientResponse) -> dict[str, Any]:
        """Parse a YouTube JSON response, raising on error payloads."""
        if response.status == 204:
            return {}

        payload: dict[str, Any] = await response.json(content_type=None)

        error_obj = payload.get("error")
        if error_obj:
            errors_list = error_obj.get("errors", [])
            first = errors_list[0] if errors_list else {}
            raise build_api_error(
                code=error_obj.get("code", response.status),
                message=error_obj.get("message", "Unknown error"),
                reason=first.get("reason", ""),
                domain=first.get("domain", ""),
            )

        return payload

    @staticmethod
    async def _raise_for_error(response: aiohttp.ClientResponse) -> None:
        """Attempt to parse an error response and raise the appropriate exception."""
        try:
            payload = await response.json(content_type=None)
            error_obj = payload.get("error", {})
            errors_list = error_obj.get("errors", [])
            first = errors_list[0] if errors_list else {}
            raise build_api_error(
                code=error_obj.get("code", response.status),
                message=error_obj.get("message", f"HTTP {response.status}"),
                reason=first.get("reason", ""),
                domain=first.get("domain", ""),
            )
        except (ValueError, KeyError):
            raise build_api_error(
                code=response.status,
                message=f"HTTP {response.status}",
            ) from None
