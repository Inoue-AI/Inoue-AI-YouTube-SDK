"""Shared pytest fixtures for the YouTube SDK test suite."""

from __future__ import annotations

import pytest
import pytest_asyncio

from youtube.client import YouTubeClient
from youtube.http.session import YouTubeSession

FAKE_TOKEN = "ya29.test_access_token_1234567890"
FAKE_API_KEY = "AIzaSyTestApiKey1234567890"


@pytest.fixture
def access_token() -> str:
    return FAKE_TOKEN


@pytest.fixture
def api_key() -> str:
    return FAKE_API_KEY


@pytest.fixture
def session(access_token: str) -> YouTubeSession:
    return YouTubeSession(access_token=access_token)


@pytest_asyncio.fixture
async def client(access_token: str) -> YouTubeClient:
    async with YouTubeClient(access_token=access_token) as c:
        yield c


@pytest_asyncio.fixture
async def api_key_client(api_key: str) -> YouTubeClient:
    async with YouTubeClient(api_key=api_key) as c:
        yield c
