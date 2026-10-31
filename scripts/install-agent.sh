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

# Fail fast on template placeholders instead of starting a probe that will
# crash-loop with "invalid length" or report to status.example.com.
# This check only prints field names and lengths, never token values.
env_value() {
  local key="$1"
  grep -E "^${key}=" /etc/polaris-agent.env | head -n 1 | cut -d= -f2- | tr -d '"'
}
placeholder_error=0
server_url="$(env_value POLARIS_SERVER_URL)"
node_id="$(env_value POLARIS_NODE_ID)"
node_name="$(env_value POLARIS_NODE_NAME)"
agent_token="$(env_value POLARIS_AGENT_TOKEN)"
if [[ "${server_url}" == *"status.example.com"* ]]; then
  printf '%s\n' "POLARIS_SERVER_URL is still the template value (status.example.com). Edit /etc/polaris-agent.env first, then rerun this script." >&2
  placeholder_error=1
fi
if [[ "${node_id}" == *"replace-with-"* || "${node_name}" == *"replace-with-"* ]]; then
  printf '%s\n' "POLARIS_NODE_ID or POLARIS_NODE_NAME is still a replace-with- placeholder. Edit /etc/polaris-agent.env first, then rerun this script." >&2
  placeholder_error=1
fi
if [[ "${agent_token}" == *"replace-with-token"* ]]; then
  printf '%s\n' "POLARIS_AGENT_TOKEN is still replace-with-token. Paste the one-time token from the Polaris dashboard into /etc/polaris-agent.env first, then rerun this script." >&2
  placeholder_error=1
elif [[ ${#agent_token} -lt 40 || ${#agent_token} -gt 256 ]]; then
  printf '%s\n' "POLARIS_AGENT_TOKEN is ${#agent_token} characters, expected 40-256 (a dashboard token is 43). The pasted token is truncated or wrong. Re-copy it from the Polaris dashboard into /etc/polaris-agent.env, then rerun this script." >&2
  placeholder_error=1
fi
if [[ "${placeholder_error}" -ne 0 ]]; then
  exit 1
fi

install -o root -g root -m 0644 ./deploy/systemd/polaris-agent.service /etc/systemd/system/polaris-agent.service
systemctl daemon-reload
systemctl enable --now polaris-agent.service
systemctl --no-pager --full status polaris-agent.service
