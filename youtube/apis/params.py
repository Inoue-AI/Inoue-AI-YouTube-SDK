"""Shared query-parameter helpers for the API namespaces.

These small utilities centralize the request-parameter encoding that every
API namespace performs: collapsing a sequence of resource ``part`` enums into
the comma-separated string YouTube expects, and normalizing the
``id``/``id-list`` overload accepted by most list endpoints.

Keeping this logic in one module removes the per-namespace ``_join`` copies
that previously diverged and makes the encoding rules testable in isolation.
"""

from __future__ import annotations

from collections.abc import Sequence
from enum import Enum


def join_parts(parts: Sequence[Enum]) -> str:
    """Join a sequence of ``part`` enum members into a comma-separated string.

    Each member is rendered via its ``.value`` so the wire format matches the
    YouTube API exactly (e.g. ``"snippet,statistics"``).
    """
    return ",".join(part.value for part in parts)


def join_ids(ids: str | Sequence[str]) -> str:
    """Normalize a single id or a sequence of ids into a comma-separated string.

    A bare ``str`` is returned unchanged; any other sequence is joined with
    commas. ``str`` is special-cased because it is itself a ``Sequence[str]``
    and must not be split into individual characters.
    """
    if isinstance(ids, str):
        return ids
    return ",".join(ids)
