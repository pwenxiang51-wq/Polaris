# Changelog

## Unreleased — 2026-10-28

- Increased retained time-series history from 30 to 31 days.
- Rendered the full 24-hour history range by uniformly sampling each sparkline to at most 240 points while preserving its first and last timestamps.
- Allowed failed IP geolocation lookups to be retried after 24 hours; successful IP metadata remains cached.
- Removed orphan per-node history files during startup.
- Changed oversized history startup handling: files above 32 MiB are read from the newest 32 MiB tail at line boundaries, compacted, persisted, and logged instead of aborting startup.
- Configured Docker JSON log rotation for Polaris and Cloudflare Tunnel at 10 MiB per file and 3 files.
- Added regression tests for 31-day retention, 24-hour chart sampling, failed geolocation retry, orphan cleanup, and oversized history recovery.

## 2026-10-28 — Muse review patch

- Removed probe-side public-IP and geolocation polling.
- Added configurable TCP probe targets and documented reachability limitations.
- Separated time-series history from current node state and added compaction.
- Added optional node expiration, USD price, monthly/total traffic quotas, and identity overrides.
- Kept the standard-library Go architecture, HTTP reporting, Cloudflare Tunnel deployment, login protection, token hashing, and existing dashboard layout.
