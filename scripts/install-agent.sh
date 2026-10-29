#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ "${EUID}" -ne 0 ]]; then
  printf '%s\n' "Run this installer as root." >&2
  exit 1
fi
if [[ ! -x ./polaris-agent ]]; then
  printf '%s\n' "Build or download ./polaris-agent before running this installer." >&2
  exit 1
fi

install -o root -g root -m 0755 ./polaris-agent /usr/local/bin/polaris-agent
if ! id polaris-agent >/dev/null 2>&1; then
  useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin polaris-agent
fi
if [[ ! -f /etc/polaris-agent.env ]]; then
  install -o root -g polaris-agent -m 0640 /dev/null /etc/polaris-agent.env
  cat > /etc/polaris-agent.env <<'EOF_ENV'
POLARIS_SERVER_URL=https://status.example.com
POLARIS_NODE_ID=replace-with-node-id
POLARIS_NODE_NAME=replace-with-node-name
POLARIS_AGENT_TOKEN=replace-with-token
# Optional: semicolon-separated NAME=host:port TCP targets.
# POLARIS_PROBE_TARGETS="CT=114.114.114.114:53;CU=123.125.81.6:53;CM=www.10086.cn:443"
EOF_ENV
  chown root:polaris-agent /etc/polaris-agent.env
  chmod 0640 /etc/polaris-agent.env
  printf '%s\n' "Edit /etc/polaris-agent.env with the values from the Polaris dashboard, then rerun this script."
  exit 0
fi

install -o root -g root -m 0644 ./deploy/systemd/polaris-agent.service /etc/systemd/system/polaris-agent.service
systemctl daemon-reload
systemctl enable --now polaris-agent.service
systemctl --no-pager --full status polaris-agent.service
