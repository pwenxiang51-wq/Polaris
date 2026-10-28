# Polaris

Polaris is a lightweight, self-hosted VPS monitoring dashboard for small Linux servers. It uses Go's standard library, embeds the web UI into the dashboard binary, and does not require Node.js or an external database service.

## Current architecture

- `cmd/polaris`: Go HTTP dashboard, authenticated APIs, node management, identity metadata lookup, and login/session security.
- `cmd/polaris-agent`: unprivileged Linux probe. It reads `/proc`, `statfs`, and network interfaces, then sends reports to Polaris over HTTPS every 15 seconds.
- `internal/shared`: API and history data types shared by both binaries.
- `internal/store`: current node state, hashed token records, traffic accounting, metadata overrides, and separately persisted time-series history.
- `internal/security`: PBKDF2-SHA256 password verification.
- `internal/webui`: embedded HTML, CSS, and JavaScript. The dashboard supports English/Chinese and dark/light themes.
- `Dockerfile` and `docker-compose.yml`: dashboard and outbound Cloudflare Tunnel connector. The Compose file does not publish the dashboard port to the host.

The agent never requires SSH access from the dashboard and does not require root or `CAP_NET_RAW` for TCP probes. The dashboard and agent are separate binaries.

## Features

- CPU, memory, swap, root disk, load average, uptime, network counters, receive/send rate, and OS/kernel/architecture.
- Automatic public IP detection from public addresses assigned to local interfaces, with route-selected local-address fallback. The route-selection check creates a UDP socket but sends no UDP payload.
- When local public IP detection cannot provide an address, the dashboard may use `CF-Connecting-IP` as a fallback only on `POST /api/v1/report`, and only when `POLARIS_TRUST_PROXY_HEADERS=true`. Do not enable trusted proxy headers if the origin is reachable directly from untrusted networks.
- Country/region/city/ISP/ASN metadata is looked up by the dashboard only once per newly observed IP and the attempted IP is persisted. Failed lookups may be retried after 24 hours; successful metadata remains cached by IP. The probe does not call `api.ipify.org`, `api6.ipify.org`, or `ipwho.is`. External metadata lookup can be disabled with `POLARIS_ENABLE_GEO_LOOKUP=false`. Manual identity overrides are available in the node-details dialog.
- Configurable TCP connection checks to up to 16 targets. Each target is tried 3 times with a 400 ms dial timeout. If all attempts fail, the UI displays `Timeout`; the internal latency value is not presented as `0 ms`.
- Per-node expiration date, price in USD, monthly traffic quota, total traffic quota, usage bars, and identity overrides. These fields are optional and only appear when set.
- One-time node token issuance, SHA-256 token-hash storage, PBKDF2-SHA256 password hashes, signed expiring sessions with server-side revocation, login rate limiting, CSRF/origin checks, security headers, and server-side Cloudflare Turnstile validation.
- Traffic allowance accounting is based on received and transmitted counter deltas after the first report. Counter resets add the new counter value rather than subtracting usage. Monthly usage resets by UTC calendar month; total tracked usage is retained. This is monitoring-derived usage, not a provider billing statement.
- The dashboard treats a node as online while its last report is less than 50 seconds old.

## Network probe configuration

`POLARIS_PROBE_TARGETS` is read at agent startup, so targets can be changed without recompiling the binary. Use semicolon-separated `NAME=host:port` entries. IPv6 addresses must use brackets.

Example in `/etc/polaris-agent.env`:

```dotenv
POLARIS_PROBE_TARGETS="CT=202.96.128.86:53;CU=123.125.81.6:53;CM=211.138.180.2:53"
```

The default candidates are:

| Label | Candidate | Test | Root required | Important limitation |
|---|---|---|---|---|
| CT | `202.96.128.86:53` | TCP connect | No | Candidate only; reachability from an overseas VPS is not guaranteed. TCP/53 can be filtered or rate-limited. |
| CU | `123.125.81.6:53` | TCP connect | No | Candidate only; a failed connection can mean filtering, routing failure, or target unavailability. |
| CM | `211.138.180.2:53` | TCP connect | No | Candidate only; test from the actual server before treating the result as representative. |

These are configurable candidates, not a guarantee of China Telecom/Unicom/Mobile backbone routing. TCP measures connection-establishment time to the chosen host and port; it is not a full route analysis and is not equivalent to ICMP ping. Public ICMP targets are also often filtered or rate-limited; ICMP generally requires root or `CAP_NET_RAW` on Linux. Polaris does not silently switch to ICMP.

Use the same target set across nodes when comparing them. Try multiple independently operated endpoints for each carrier if a single DNS host is unavailable. Change a target by editing `POLARIS_PROBE_TARGETS` in the agent environment file and restarting `polaris-agent.service`.

## History retention and storage

Current node state and time-series history are stored separately:

- `nodes.json`: latest metrics, node metadata, overrides, traffic-accounting counters, and metadata lookup state. It does not contain full history arrays.
- `tokens.json`: token hashes and node bindings. Raw agent tokens are never stored.
- `history/<node-id>.jsonl`: append-oriented history records, mode `0600` under a mode `0700` data directory.

History retention is 31 days (to cover the longest calendar month). Compaction keeps raw samples from the most recent hour, minute-bucket means from 1 hour through 24 hours, and hour-bucket means older than 24 hours. Compaction is periodically applied to per-node history files, not by rewriting the complete history as part of every node-state update. At startup, orphan history files for deleted/nonexistent nodes are removed. If a history file exceeds 32 MiB, Polaris reads its newest 32 MiB tail on line boundaries, compacts the retained records, overwrites the oversized file, and logs the truncation rather than refusing to start. The authenticated history API returns the last 24 hours for the dashboard charts: CPU, memory, inbound/outbound rate, and each configured TCP target's latency. The browser evenly samples the full returned range to at most 240 points per sparkline while preserving the first and last timestamps; it does not plot only the newest tail of the history.

On first startup after upgrading from an older version, Polaris migrates legacy `History` arrays embedded in `nodes.json` into `history/<node-id>.jsonl` and rewrites the current-state file without the old history arrays. Keep a backup of the data directory before upgrading. This migration does not require deleting existing tokens.

## Requirements

- Linux VPS with Docker Engine and Docker Compose plugin for the dashboard deployment.
- A domain managed in Cloudflare, a remotely managed Tunnel, and a Turnstile widget restricted to the dashboard hostname.
- Go 1.23 or later if building the binaries outside Docker.
- Python 3 for initial password-hash generation by `scripts/setup-secrets.sh`.

## Deploy the dashboard

1. Create a remotely managed Cloudflare Tunnel. Configure its public hostname (for example `status.example.com`) to route to `http://polaris:8080` inside the Compose network.
2. Create a Turnstile widget and restrict it to that exact hostname.
3. Copy this source tree to your VPS and run the secret setup script:

   ```bash
   sudo ./scripts/setup-secrets.sh
   ```

4. Start the dashboard and tunnel:

   ```bash
   docker compose up -d --build
   docker compose ps
   docker compose logs --tail=100 polaris cloudflared
   ```

5. Open `https://status.example.com` and sign in with the administrator account created during setup.

The setup script writes non-secret settings to `secrets/polaris.env` and stores the administrator password hash, session-signing secret, Turnstile secret key, and Tunnel token in separate files. Docker uses `json-file` log rotation for both Polaris and Cloudflare Tunnel, capped at 10 MiB per file and 3 files per container. The directory is mode `0700`; files use restricted permissions for their service identities. Do not commit or share the generated secret files. They are excluded by `.gitignore` and `.dockerignore`.

Cloudflare Tunnel provides an outbound connection from the host and avoids exposing the dashboard port directly. Turnstile validation happens server-side; the browser widget is not treated as authentication on its own.

## Add a monitoring probe

1. Sign in to Polaris and choose **Add probe**.
2. Enter a display name and node identifier, such as `la-edge-01`.
3. Copy the one-time token and generated environment values. The raw token is not recoverable from the dashboard; rotate it if lost.
4. Build or download the agent and install it on the target Linux host.

Build both Linux binaries with no CGO dependency:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o polaris ./cmd/polaris
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o polaris-agent ./cmd/polaris-agent
```

For ARM64, set `GOARCH=arm64`. Copy `polaris-agent` and `deploy/systemd/polaris-agent.service` to the probe host. Create `/etc/polaris-agent.env` with the dashboard values. To set custom network targets, add `POLARIS_PROBE_TARGETS` as shown above. Keep the environment file readable only by root and the `polaris-agent` group; never place the token in command-line arguments or shell history.

The installer can create a protected initial environment template:

```bash
sudo ./scripts/install-agent.sh
sudoedit /etc/polaris-agent.env
sudo ./scripts/install-agent.sh
```

Review the example values before starting the service. The probe does not need root privileges to read ordinary Linux metrics, and TCP probing does not require `CAP_NET_RAW`.

## Node metadata

Open the pencil/details control on a node card to edit optional metadata:

- Expiry date (`YYYY-MM-DD`), with a warning when 7 days or fewer remain.
- Price in USD.
- Monthly and total traffic allowances in GiB.
- Manual overrides for IPv4/IPv6, country code, country/region, city, ISP, and ASN.

Empty values clear the corresponding override or optional field. Manual identity values override automatically detected values for display but do not replace the automatically observed IP used to decide when a new metadata lookup is due.

Monthly usage is calculated from byte-counter deltas reported by the agent, and resets at the first report in a new UTC month. The first report establishes a counter baseline and is not charged as usage. A counter decrease (for example after reboot) is treated as a reset and the new counter value is added. This is approximate host traffic accounting; provider-side billing can use different interfaces or accounting rules.

## Configuration

Dashboard settings are loaded from `secrets/polaris.env`:

| Variable | Purpose |
|---|---|
| `POLARIS_LISTEN` | Internal HTTP listen address; default `0.0.0.0:8080`. Do not publish it directly. |
| `POLARIS_DATA_DIR` | Persistent data directory; `/data` in Compose. |
| `POLARIS_ADMIN_USER` | Single administrator account name. |
| `POLARIS_ADMIN_PASSWORD_HASH_FILE` | Path to the separately mounted PBKDF2-SHA256 password hash. |
| `POLARIS_SESSION_SECRET_FILE` | Path to the separately mounted session-signing secret. |
| `POLARIS_TURNSTILE_SITEKEY` | Public Turnstile site key used by the login page. |
| `POLARIS_TURNSTILE_SECRET_FILE` | Path to the server-side Turnstile secret. |
| `POLARIS_ALLOWED_HOSTNAME` | Exact public hostname expected by Turnstile. |
| `POLARIS_EXTERNAL_URL` | Full HTTPS dashboard URL. |
| `POLARIS_TRUST_PROXY_HEADERS` | Trust Cloudflare's client IP header. Enable only when the origin is reachable exclusively through the trusted tunnel/proxy. |
| `POLARIS_ENABLE_GEO_LOOKUP` | `true` by default; set `false` to disable one-time `ipwho.is` lookups. |

Agent settings are read from its service environment:

| Variable | Purpose |
|---|---|
| `POLARIS_SERVER_URL` | Root HTTPS URL of the dashboard. |
| `POLARIS_NODE_ID` | Node identifier assigned in the dashboard. |
| `POLARIS_NODE_NAME` | Display name. |
| `POLARIS_AGENT_TOKEN` | One-time issued token; keep in the protected environment file. |
| `POLARIS_PROBE_TARGETS` | Optional semicolon-separated `NAME=host:port` TCP targets; defaults to CT/CU/CM candidates above. |

## Privacy notes

The probe makes its normal HTTPS report to the configured Polaris endpoint every 15 seconds. It does not poll third-party public-IP or geo-IP APIs. Public IP is read from local network interfaces or the route-selected local address first. If local public IP detection returns nothing, the dashboard can use the trusted `CF-Connecting-IP` header from the agent report request as a fallback. When geo lookup is enabled, the dashboard contacts `ipwho.is` once when a node's selected public IP is first seen or changes; the request contains only the IP address, not the node token, node name, or traffic metrics. The attempted lookup IP is persisted to avoid retrying every report. Set `POLARIS_ENABLE_GEO_LOOKUP=false` to disable it and use manual identity overrides if needed.

## Verification

```bash
test -z "$(gofmt -l cmd internal)"
go test -race ./...
go vet ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/polaris ./cmd/polaris
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /tmp/polaris-agent ./cmd/polaris-agent
```

Local unit tests and static checks do not replace end-to-end verification on the actual VPS. After deployment, check `/healthz`, add one real probe, verify resource counters and TCP-target statuses, then measure disk/RSS on the 1 GB host.

## Backups and updates

- Back up the Compose volume holding `/data` and the secret files separately.
- Keep safe recovery credentials and rotate any lost token.
- Before updating, back up `/data`; run `docker compose up -d --build` after code review.
- Never configure a direct public origin route that bypasses the Cloudflare Tunnel policy.

## License

MIT. See [LICENSE](LICENSE).
