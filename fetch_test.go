package main

import (
	"errors"
	"strings"
	"testing"
)

func TestFetchAll(t *testing.T) {
	calls := []string{}
	get := func(url string, headers map[string]string) (int, []byte, error) {
		calls = append(calls, url)
		if headers["x-api-key"] != "k" {
			t.Errorf("api key header missing: %v", headers)
		}
		if strings.Contains(url, "offset=0") {
			return 200, []byte(`{"offset":0,"next":2,"data":[{"title":"a","year":1},{"title":"b","year":2}]}`), nil
		}
		return 200, []byte(`{"offset":2,"next":null,"data":[{"title":"c","year":3}]}`), nil
	}
	as, err := fetchAll(get, "123", "k", 10)
	if err != nil || len(as) != 3 {
		t.Fatalf("got %d articles, %v", len(as), err)
	}
	if as[0].Title != "c" || as[2].Title != "a" {
		t.Errorf("should be sorted newest first, got %+v", as)
	}
	if len(calls) != 2 || !strings.HasPrefix(calls[0], "https://api.semanticscholar.org/graph/v1/author/123/papers?") || !strings.Contains(calls[0], "limit=100") || !strings.Contains(calls[1], "offset=2") {
		t.Errorf("calls = %v", calls)
	}
	// limit truncates after every page has been read, so "limit" means the
	// newest N rather than the API's first N.
	calls = nil
	as, _ = fetchAll(get, "123", "k", 1)
	if len(as) != 1 || len(calls) != 2 || as[0].Title != "c" {
		t.Errorf("limit=1: %d articles, %d calls, first=%+v", len(as), len(calls), as)
	}
	// non-200 -> error
	bad := func(string, map[string]string) (int, []byte, error) { return 429, []byte("slow down"), nil }
	if _, err := fetchAll(bad, "123", "", 5); err == nil || !strings.Contains(err.Error(), "429") {
		t.Errorf("429 should error with status, got %v", err)
	}
	failing := func(string, map[string]string) (int, []byte, error) { return 0, nil, errors.New("net down") }
	if _, err := fetchAll(failing, "123", "", 5); err == nil {
		t.Error("transport error should propagate")
	}
	if _, err := fetchAll(get, "", "", 5); err == nil {
		t.Error("empty id should error")
	}
	// no API key -> no header
	noKey := func(_ string, headers map[string]string) (int, []byte, error) {
		if _, ok := headers["x-api-key"]; ok {
			t.Errorf("x-api-key must be absent without a key: %v", headers)
		}
		return 200, []byte(`{"next":null,"data":[]}`), nil
	}
	if as, err := fetchAll(noKey, "123", "", 5); err != nil || len(as) != 0 {
		t.Errorf("empty author: %v %v", as, err)
	}
	// a "next" that never ends stops after maxPages
	endless := 0
	forever := func(string, map[string]string) (int, []byte, error) {
		endless++
		return 200, []byte(`{"next":1,"data":[{"title":"x"}]}`), nil
	}
	if as, err := fetchAll(forever, "123", "", 1000); err != nil || endless != maxPages || len(as) != maxPages {
		t.Errorf("endless pagination: %d calls, %d articles, %v", endless, len(as), err)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("héllo wörld", 5); got != "héllo…" {
		t.Errorf("truncate on rune boundary = %q", got)
	}
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate no-op = %q", got)
	}
}
