# Changelog

## 2.0.0

- First release as a WebAssembly plugin (sandboxed; installable from the directory in goblog ≥ 0.2.9). Replaces goblog's compiled-in scholar plugin.
- Semantic Scholar only: the Google Scholar source and the `source`, `scholar_id`, `profile_cache` and `article_cache` settings are gone. Publications are cached in the plugin store instead of files, with a new `cache_hours` setting (default 24).
