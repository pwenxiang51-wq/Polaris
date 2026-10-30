#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SECRET_DIR="${ROOT_DIR}/secrets"
APP_ENV="${SECRET_DIR}/polaris.env"
ADMIN_HASH_FILE="${SECRET_DIR}/admin-password-hash"
SESSION_SECRET_FILE="${SECRET_DIR}/session-secret"
TURNSTILE_SECRET_FILE="${SECRET_DIR}/turnstile-secret"
TUNNEL_FILE="${SECRET_DIR}/cloudflared.token"
TELEGRAM_TOKEN_FILE="${SECRET_DIR}/telegram-bot-token"
TELEGRAM_CHAT_FILE="${SECRET_DIR}/telegram-chat-id"

if [[ "${EUID}" -ne 0 ]]; then
  printf '%s\n' "Run this setup script as root, for example with sudo." >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  printf '%s\n' "Python 3 is required to create the initial password hash." >&2
  exit 1
fi

mkdir -p "${SECRET_DIR}"
chmod 700 "${SECRET_DIR}"
chmod g-s "${SECRET_DIR}"
if [[ -e "${APP_ENV}" || -e "${ADMIN_HASH_FILE}" || -e "${SESSION_SECRET_FILE}" || -e "${TURNSTILE_SECRET_FILE}" || -e "${TUNNEL_FILE}" || -e "${TELEGRAM_TOKEN_FILE}" || -e "${TELEGRAM_CHAT_FILE}" ]]; then
  printf '%s' "Secret files already exist. Replace both files? [y/N] "
  read -r answer
  if [[ ! "${answer}" =~ ^[Yy]$ ]]; then
    printf '%s\n' "No files were changed."
    exit 0
  fi
fi

read -r -p "Public dashboard hostname (example: status.example.com): " hostname
if [[ ! "${hostname}" =~ ^[A-Za-z0-9.-]+$ || "${hostname}" == *..* ]]; then
  printf '%s\n' "The hostname format is invalid." >&2
  exit 1
fi
read -r -p "Administrator username: " admin_user
if [[ ! "${admin_user}" =~ ^[A-Za-z0-9._-]{1,64}$ ]]; then
  printf '%s\n' "Use 1-64 letters, numbers, dots, underscores, or hyphens for the username." >&2
  exit 1
fi
read -r -s -p "Administrator password (at least 14 characters; use a unique password): " password
printf '\n'
read -r -s -p "Confirm administrator password: " password_confirm
printf '\n'
if [[ "${password}" != "${password_confirm}" || ${#password} -lt 14 ]]; then
  printf '%s\n' "Passwords do not match or are shorter than 14 characters." >&2
  exit 1
fi
read -r -p "Cloudflare Turnstile site key: " turnstile_sitekey
if [[ -z "${turnstile_sitekey}" ]]; then
  printf '%s\n' "The Turnstile site key is required." >&2
  exit 1
fi
read -r -s -p "Cloudflare Turnstile secret key: " turnstile_secret
printf '\n'
if [[ -z "${turnstile_secret}" ]]; then
  printf '%s\n' "The Turnstile secret key is required." >&2
  exit 1
fi
read -r -s -p "Telegram Bot Token (optional; press Enter to skip): " telegram_token
printf '\n'
read -r -s -p "Telegram Chat ID (optional; press Enter to skip): " telegram_chat
printf '\n'
if [[ ( -n "${telegram_token}" && -z "${telegram_chat}" ) || ( -z "${telegram_token}" && -n "${telegram_chat}" ) ]]; then
  printf '%s\n' "Telegram Bot Token and Chat ID must be configured together." >&2
  exit 1
fi
read -r -s -p "Cloudflare Tunnel token: " tunnel_token
printf '\n'
if [[ -z "${tunnel_token}" ]]; then
  printf '%s\n' "The Cloudflare Tunnel token is required." >&2
  exit 1
fi

hash="$(PASSWORD_VALUE="${password}" python3 - <<'PY'
import base64
import hashlib
import os
import secrets

password = os.environ["PASSWORD_VALUE"].encode("utf-8")
salt = secrets.token_bytes(16)
iterations = 310000
derived = hashlib.pbkdf2_hmac("sha256", password, salt, iterations, dklen=32)
encode = lambda value: base64.urlsafe_b64encode(value).decode("ascii").rstrip("=")
print(f"pbkdf2_sha256:{iterations}:{encode(salt)}:{encode(derived)}")
PY
)"
session_secret="$(python3 - <<'PY'
import secrets
print(secrets.token_hex(32))
PY
)"

cat > "${APP_ENV}" <<EOF_ENV
POLARIS_LISTEN=0.0.0.0:8080
POLARIS_DATA_DIR=/data
POLARIS_ADMIN_USER=${admin_user}
POLARIS_ADMIN_PASSWORD_HASH_FILE=/run/secrets/admin_password_hash
POLARIS_SESSION_SECRET_FILE=/run/secrets/session_secret
POLARIS_TURNSTILE_SITEKEY=${turnstile_sitekey}
POLARIS_TURNSTILE_SECRET_FILE=/run/secrets/turnstile_secret
POLARIS_ALLOWED_HOSTNAME=${hostname}
POLARIS_EXTERNAL_URL=https://${hostname}
POLARIS_TRUST_PROXY_HEADERS=true
POLARIS_ENABLE_GEO_LOOKUP=true
POLARIS_TELEGRAM_BOT_TOKEN_FILE=/run/secrets/telegram_bot_token
POLARIS_TELEGRAM_CHAT_ID_FILE=/run/secrets/telegram_chat_id
# Optional dashboard attribution. Leave empty for a neutral fork default.
POLARIS_FOOTER_COPYRIGHT=
POLARIS_FOOTER_GITHUB=
POLARIS_FOOTER_TELEGRAM=
EOF_ENV
printf '%s\n' "${hash}" > "${ADMIN_HASH_FILE}"
printf '%s\n' "${session_secret}" > "${SESSION_SECRET_FILE}"
printf '%s\n' "${turnstile_secret}" > "${TURNSTILE_SECRET_FILE}"
printf '%s\n' "${tunnel_token}" > "${TUNNEL_FILE}"
if [[ -n "${telegram_token}" ]]; then printf '%s\n' "${telegram_token}" > "${TELEGRAM_TOKEN_FILE}"; printf '%s\n' "${telegram_chat}" > "${TELEGRAM_CHAT_FILE}"; else : > "${TELEGRAM_TOKEN_FILE}"; : > "${TELEGRAM_CHAT_FILE}"; fi
chmod 600 "${APP_ENV}"
chown root:10001 "${ADMIN_HASH_FILE}" "${SESSION_SECRET_FILE}" "${TURNSTILE_SECRET_FILE}"
chmod 440 "${ADMIN_HASH_FILE}" "${SESSION_SECRET_FILE}" "${TURNSTILE_SECRET_FILE}"
chown root:65532 "${TUNNEL_FILE}"
chmod 440 "${TUNNEL_FILE}" "${TELEGRAM_TOKEN_FILE}" "${TELEGRAM_CHAT_FILE}"
unset password password_confirm turnstile_secret tunnel_token telegram_token telegram_chat hash session_secret

printf '\n%s\n' "Polaris configuration and secret files were created:" \
  "  ${APP_ENV} (mode 600)" \
  "  ${ADMIN_HASH_FILE} (mode 440, dashboard group 10001)" \
  "  ${SESSION_SECRET_FILE} (mode 440, dashboard group 10001)" \
  "  ${TURNSTILE_SECRET_FILE} (mode 440, dashboard group 10001)" \
  "  ${TUNNEL_FILE} (mode 440, cloudflared group 65532)" \
  "  ${TELEGRAM_TOKEN_FILE} and ${TELEGRAM_CHAT_FILE} (mode 440; optional)" \
  "The secret directory remains mode 700. Do not commit or share these files." \
  "Next: configure the Cloudflare Tunnel public hostname to route to http://polaris:8080, then run docker compose up -d --build."
