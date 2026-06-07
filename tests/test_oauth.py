"""Tests for the OAuth 2.0 token-refresh helper (mocked HTTP)."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses

from youtube.exceptions import YouTubeAuthError, YouTubeConfigError
from youtube.oauth import GOOGLE_TOKEN_ENDPOINT, OAuthClient, TokenResponse

_TOKEN_RE = re.compile(r"^https://oauth2\.googleapis\.com/token$")

SAMPLE_TOKEN_RESPONSE = {
    "access_token": "ya29.NEW_ACCESS_TOKEN",
    "expires_in": 3599,
    "scope": "https://www.googleapis.com/auth/youtube",
    "token_type": "Bearer",
}

SAMPLE_ERROR_RESPONSE = {
    "error": "invalid_grant",
    "error_description": "Token has been expired or revoked.",
}


class TestTokenResponseModel:
    """Model-level tests for TokenResponse."""

    def test_parses_minimal_response(self) -> None:
        token = TokenResponse.model_validate({"access_token": "abc", "expires_in": 3600})
        assert token.access_token == "abc"
        assert token.expires_in == 3600
        assert token.token_type == "Bearer"
        assert token.refresh_token is None

    def test_ignores_unknown_fields(self) -> None:
        token = TokenResponse.model_validate(
            {"access_token": "abc", "expires_in": 1, "id_token": "JWT", "extra": 1}
        )
        assert token.access_token == "abc"


class TestOAuthClientConfig:
    """Constructor / argument validation."""

    def test_requires_client_id_and_secret(self) -> None:
        with pytest.raises(YouTubeConfigError):
            OAuthClient(client_id="", client_secret="s")
        with pytest.raises(YouTubeConfigError):
            OAuthClient(client_id="i", client_secret="")

    async def test_refresh_requires_refresh_token(self) -> None:
        async with OAuthClient(client_id="i", client_secret="s") as oauth:
            with pytest.raises(YouTubeConfigError):
                await oauth.refresh_access_token("")

    async def test_exchange_requires_code_and_redirect(self) -> None:
        async with OAuthClient(client_id="i", client_secret="s") as oauth:
            with pytest.raises(YouTubeConfigError):
                await oauth.exchange_authorization_code("", "https://cb")
            with pytest.raises(YouTubeConfigError):
                await oauth.exchange_authorization_code("code", "")


class TestRefreshAccessToken:
    """Token-refresh exchange with mocked Google endpoint."""

    async def test_refresh_success(self) -> None:
        async with OAuthClient(client_id="cid", client_secret="csecret") as oauth:
            with aioresponses() as mocked:
                mocked.post(_TOKEN_RE, payload=SAMPLE_TOKEN_RESPONSE)
                token = await oauth.refresh_access_token("rt-123")
                assert isinstance(token, TokenResponse)
                assert token.access_token == "ya29.NEW_ACCESS_TOKEN"
                assert token.expires_in == 3599

    async def test_refresh_sends_correct_form(self) -> None:
        captured: dict[str, str] = {}

        def _callback(url: object, **kwargs: object) -> None:
            data = kwargs.get("data")
            assert isinstance(data, dict)
            captured.update(data)

        async with OAuthClient(client_id="cid", client_secret="csecret") as oauth:
            with aioresponses() as mocked:
                mocked.post(_TOKEN_RE, payload=SAMPLE_TOKEN_RESPONSE, callback=_callback)
                await oauth.refresh_access_token("rt-123")

        assert captured["grant_type"] == "refresh_token"
        assert captured["refresh_token"] == "rt-123"
        assert captured["client_id"] == "cid"
        assert captured["client_secret"] == "csecret"

    async def test_refresh_invalid_grant_raises_auth_error(self) -> None:
        async with OAuthClient(client_id="cid", client_secret="csecret") as oauth:
            with aioresponses() as mocked:
                mocked.post(_TOKEN_RE, status=400, payload=SAMPLE_ERROR_RESPONSE)
                with pytest.raises(YouTubeAuthError) as exc_info:
                    await oauth.refresh_access_token("revoked-token")
        assert exc_info.value.reason == "invalid_grant"
        assert exc_info.value.code == 400

    async def test_exchange_authorization_code_with_pkce(self) -> None:
        captured: dict[str, str] = {}

        def _callback(url: object, **kwargs: object) -> None:
            data = kwargs.get("data")
            assert isinstance(data, dict)
            captured.update(data)

        async with OAuthClient(client_id="cid", client_secret="csecret") as oauth:
            with aioresponses() as mocked:
                mocked.post(
                    _TOKEN_RE,
                    payload={**SAMPLE_TOKEN_RESPONSE, "refresh_token": "rt-new"},
                    callback=_callback,
                )
                token = await oauth.exchange_authorization_code(
                    "auth-code", "https://app/callback", code_verifier="verifier"
                )

        assert token.refresh_token == "rt-new"
        assert captured["grant_type"] == "authorization_code"
        assert captured["code"] == "auth-code"
        assert captured["redirect_uri"] == "https://app/callback"
        assert captured["code_verifier"] == "verifier"


class TestModuleConstants:
    def test_endpoint_constant(self) -> None:
        assert GOOGLE_TOKEN_ENDPOINT == "https://oauth2.googleapis.com/token"
