"""Tests for the shared query-parameter helpers."""

from __future__ import annotations

from youtube.apis.params import join_ids, join_parts
from youtube.models.videos import VideoPart


class TestJoinParts:
    def test_single_part(self) -> None:
        assert join_parts([VideoPart.SNIPPET]) == "snippet"

    def test_multiple_parts_preserve_order(self) -> None:
        result = join_parts([VideoPart.SNIPPET, VideoPart.STATISTICS])
        assert result == "snippet,statistics"

    def test_accepts_tuple(self) -> None:
        result = join_parts((VideoPart.STATUS, VideoPart.CONTENT_DETAILS))
        assert result == "status,contentDetails"

    def test_empty(self) -> None:
        assert join_parts([]) == ""


class TestJoinIds:
    def test_single_string_unchanged(self) -> None:
        assert join_ids("abc123") == "abc123"

    def test_string_is_not_split_into_chars(self) -> None:
        # A bare str must NOT be treated as a sequence of characters.
        assert join_ids("xyz") == "xyz"

    def test_list_of_ids(self) -> None:
        assert join_ids(["a", "b", "c"]) == "a,b,c"

    def test_tuple_of_ids(self) -> None:
        assert join_ids(("a", "b")) == "a,b"

    def test_single_element_list(self) -> None:
        assert join_ids(["only"]) == "only"
