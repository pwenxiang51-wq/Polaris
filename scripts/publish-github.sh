#!/usr/bin/env bash
set -Eeuo pipefail

OWNER="${1:-pwenxiang51-wq}"
REPO="${2:-polaris}"
FULL_NAME="${OWNER}/${REPO}"
EXPECTED_REMOTE="https://github.com/${FULL_NAME}.git"

if ! command -v gh >/dev/null 2>&1; then
  printf '%s\n' "GitHub CLI (gh) is required. Install it and run: gh auth login" >&2
  exit 1
fi
if ! gh auth status >/dev/null 2>&1; then
  printf '%s\n' "No active GitHub CLI login. Run: gh auth login" >&2
  exit 1
fi
for secret_file in \
  secrets/polaris.env \
  secrets/admin-password-hash \
  secrets/session-secret \
  secrets/turnstile-secret \
  secrets/cloudflared.token; do
  if [[ -f "${secret_file}" ]]; then
    printf 'Refusing to publish while local secret file exists: %s\n' "${secret_file}" >&2
    exit 1
  fi
done

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  git init -b main
fi
if ! git rev-parse --verify HEAD >/dev/null 2>&1; then
  git add .
  git commit -m "Initial Polaris monitoring dashboard"
fi

if gh repo view "${FULL_NAME}" >/dev/null 2>&1; then
  if git remote get-url origin >/dev/null 2>&1; then
    ACTUAL_REMOTE="$(git remote get-url origin)"
    if [[ "${ACTUAL_REMOTE}" != "${EXPECTED_REMOTE}" && "${ACTUAL_REMOTE}" != "git@github.com:${FULL_NAME}.git" ]]; then
      printf 'Refusing to push: origin points to %s, not %s.\n' "${ACTUAL_REMOTE}" "${EXPECTED_REMOTE}" >&2
      exit 1
    fi
  else
    git remote add origin "${EXPECTED_REMOTE}"
  fi
  git add .
  git commit -m "Update Polaris monitoring dashboard" || true
  git push -u origin HEAD:main
else
  if git remote get-url origin >/dev/null 2>&1; then
    printf '%s\n' "The origin remote already exists. Remove or correct it before creating a new repository." >&2
    exit 1
  fi
  gh repo create "${FULL_NAME}" --public --source . --remote origin --push \
    --description "Lightweight self-hosted VPS monitoring dashboard and Linux probe"
fi

printf 'Published repository: https://github.com/%s\n' "${FULL_NAME}"
