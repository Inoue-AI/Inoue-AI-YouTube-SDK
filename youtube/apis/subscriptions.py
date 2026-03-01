"""Subscriptions API namespace.

Wraps the following YouTube endpoints:

* GET    /subscriptions — list subscriptions
* POST   /subscriptions — subscribe to a channel
* DELETE /subscriptions — unsubscribe

Reference: https://developers.google.com/youtube/v3/docs/subscriptions
"""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from youtube.apis.base import BaseAPI
from youtube.config import YOUTUBE_API_BASE_URL
from youtube.models.subscriptions import (
    Subscription,
    SubscriptionListResponse,
    SubscriptionOrder,
    SubscriptionPart,
)

_BASE = YOUTUBE_API_BASE_URL


class SubscriptionsAPI(BaseAPI):
    """Methods for managing YouTube subscriptions."""

    async def list(
        self,
        *,
        parts: list[SubscriptionPart],
        channel_id: str | None = None,
        id: str | list[str] | None = None,
        mine: bool | None = None,
        my_recent_subscribers: bool | None = None,
        my_subscribers: bool | None = None,
        for_channel_id: str | None = None,
        max_results: int = 5,
        page_token: str | None = None,
        order: SubscriptionOrder | None = None,
    ) -> SubscriptionListResponse:
        """Return a list of subscriptions matching the given filter."""
        params: dict[str, Any] = {
            "part": _join(parts),
            "maxResults": max_results,
        }
        if channel_id is not None:
            params["channelId"] = channel_id
        if id is not None:
            params["id"] = ",".join(id) if isinstance(id, list) else id
        if mine is not None:
            params["mine"] = str(mine).lower()
        if my_recent_subscribers is not None:
            params["myRecentSubscribers"] = str(my_recent_subscribers).lower()
        if my_subscribers is not None:
            params["mySubscribers"] = str(my_subscribers).lower()
        if for_channel_id is not None:
            params["forChannelId"] = for_channel_id
        if page_token is not None:
            params["pageToken"] = page_token
        if order is not None:
            params["order"] = order.value

        payload = await self._session.get(f"{_BASE}/subscriptions", params=params)
        return SubscriptionListResponse.model_validate(payload)

    async def iter_mine(
        self,
        *,
        parts: list[SubscriptionPart],
        max_results: int = 50,
        order: SubscriptionOrder | None = None,
    ) -> AsyncIterator[Subscription]:
        """Async generator that pages through the authenticated user's subscriptions."""
        page_token: str | None = None
        while True:
            page = await self.list(
                parts=parts,
                mine=True,
                max_results=max_results,
                page_token=page_token,
                order=order,
            )
            for sub in page.items:
                yield sub
            if not page.next_page_token:
                break
            page_token = page.next_page_token

    async def insert(
        self,
        body: Subscription,
        *,
        parts: list[SubscriptionPart] | None = None,
    ) -> Subscription:
        """Subscribe to a channel.

        The ``body`` must include ``snippet.resource_id`` with
        ``kind="youtube#channel"`` and the target ``channel_id``.
        """
        parts = parts or [SubscriptionPart.SNIPPET]
        params: dict[str, Any] = {"part": _join(parts)}
        payload = await self._session.post(
            f"{_BASE}/subscriptions",
            json=body.model_dump(by_alias=True, exclude_none=True),
            params=params,
        )
        return Subscription.model_validate(payload)

    async def delete(self, *, id: str) -> None:
        """Unsubscribe (delete a subscription by its ID)."""
        await self._session.delete(f"{_BASE}/subscriptions", params={"id": id})


def _join(parts: list[SubscriptionPart]) -> str:
    return ",".join(p.value for p in parts)
