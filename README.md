# tautulli-exporter

Prometheus exporter for [Tautulli](https://tautulli.com/) activity statistics.
Clean-room implementation (MIT) that keeps the metric and environment contract
of the unmaintained `nwalke/tautulli_exporter` image, so existing dashboards
keep working.

Queries Tautulli's v2 API (`cmd=get_activity`) on each scrape.

## Metrics

| metric | type | meaning |
|---|---|---|
| `tautulli_up` | gauge | last API query succeeded (1/0) |
| `tautulli_stream_count` | gauge | active streams |
| `tautulli_stream_direct_play` | gauge | direct-play streams |
| `tautulli_stream_direct_stream` | gauge | direct streams |
| `tautulli_stream_count_transcode` | gauge | transcoding streams |
| `tautulli_bandwidth_total` | gauge | total streaming bandwidth (kbps) |
| `tautulli_bandwidth_lan` | gauge | LAN streaming bandwidth (kbps) |
| `tautulli_bandwidth_wan` | gauge | WAN streaming bandwidth (kbps) |

## Configuration

| env | default | |
|---|---|---|
| `TAUTULLI_URI` | — | base URL, e.g. `http://tautulli:8181` (required) |
| `TAUTULLI_API_KEY` | — | Tautulli API key (required) |
| `SERVE_PORT` | `9487` | listen port |

## Running

```
docker run -e TAUTULLI_URI=http://tautulli:8181 -e TAUTULLI_API_KEY=... \
  -p 9487:9487 ghcr.io/jabrown93/tautulli-exporter:latest
```

Releases are cut by semantic-release from Conventional Commits; images are
multi-arch (amd64/arm64), cosign-signed, published to GHCR.
