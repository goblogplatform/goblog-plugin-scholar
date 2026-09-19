# goblog-plugin-scholar

A [goblog](https://github.com/goblogplatform/goblog) WebAssembly plugin that adds a **Research** page (`/research`, in the nav) listing your publications from [Semantic Scholar](https://www.semanticscholar.org/): title linked to the paper, authors, year, venue and citation count, newest first. Version 2.0.0 replaces the scholar plugin that used to be compiled into goblog.

## Install

From your goblog's **Admin → Plugins → Browse**, search for *Scholar Publications* and click **Install** (goblog 0.2.9 or newer). Or download `plugin.wasm` from the [latest release](https://github.com/goblogplatform/goblog-plugin-scholar/releases/latest) into `plugins/wasm/scholar.wasm`, write `plugins/wasm/scholar.json` containing `{"allowed_hosts":["api.semanticscholar.org"]}`, and restart goblog.

## Settings

Under **Admin → Settings → Scholar Publications**:

| Setting | Default | Meaning |
|---|---|---|
| `enabled` | `false` | Set to `true` to enable the Research page |
| `semantic_scholar_id` | | Your Semantic Scholar author id: the number at the end of your `semanticscholar.org/author/...` URL (e.g. `1792904`) |
| `semantic_scholar_api_key` | | Optional; an [API key](https://www.semanticscholar.org/product/api) raises the rate limit |
| `article_limit` | `50` | Maximum number of publications to show (the newest N) |
| `cache_hours` | `24` | How long fetched publications are reused before refreshing |

Installs upgrading from goblog's compiled-in scholar plugin keep their `enabled`, `semantic_scholar_id`, `semantic_scholar_api_key` and `article_limit` values. Google Scholar is no longer supported, so the old `source` and `scholar_id` settings are ignored, as are `profile_cache` and `article_cache` (the cache now lives in goblog's database).

## How it works

The plugin talks to `api.semanticscholar.org` only (`/graph/v1/author/{id}/papers`), and goblog's sandbox blocks every other host. Fetched publications are kept in the plugin store (goblog's `plugin_store` table) and reused for `cache_hours`; an hourly `refresh` job re-fetches once the cache is stale, so visitors rarely wait on the API. If a fetch fails the stale cache is still shown; with no cache at all the page says publications are temporarily unavailable.

## Build it yourself

```bash
go test ./...   # unit tests run natively
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags="-s -w" -o plugin.wasm .
```

Needs Go 1.25 or newer (`go.mod` pins the toolchain, so `GOTOOLCHAIN=auto` fetches it). Check the result with `goblog validate-plugin plugin.wasm`. Tag a release as `vX.Y.Z`; the workflow builds and uploads `plugin.wasm` for you.

## License

Apache-2.0.
