# Changelog

## 2.0.0

- First release as a WebAssembly plugin (sandboxed; installable from the directory in goblog ≥ 0.3.0, the first release without the compiled-in scholar plugin this replaces).
- Semantic Scholar only: the Google Scholar source and the `source`, `scholar_id`, `profile_cache` and `article_cache` settings are gone. Publications are cached in the plugin store instead of files, with a new `cache_hours` setting (default 24): the page is served from the cache and an hourly job refreshes it. `article_limit` is capped at 500.
