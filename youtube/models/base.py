"""Shared base types used across all YouTube API response models."""

from __future__ import annotations

from pydantic import BaseModel, ConfigDict
from pydantic.alias_generators import to_camel


class YouTubeBaseModel(BaseModel):
    """Base class for all SDK models (both request bodies and responses).

    Configuration:
        - ``alias_generator=to_camel`` — accepts camelCase keys from the API
          while keeping snake_case attribute names in Python.
        - ``populate_by_name=True`` — allows constructing models with either
          the Python field name or the camelCase alias.
        - ``frozen=True`` — instances are immutable once created.
        - ``extra="ignore"`` — unknown fields returned by future API versions
          are silently ignored (forward compatibility).
    """

    model_config = ConfigDict(
        alias_generator=to_camel,
        populate_by_name=True,
        frozen=True,
        extra="ignore",
    )
