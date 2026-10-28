# Security Policy

## Supported Versions

Security fixes are intended for the latest released version on the `main` branch.

## Reporting a Vulnerability

Do not publish credentials, tunnel tokens, agent tokens, session secrets, or live IP addresses in a public issue. Open a private security advisory for this repository when private reporting is available. If private advisories are unavailable, contact the repository maintainer through a verified GitHub account channel and provide only the minimum reproduction details required.

## Deployment Requirements

- Keep `secrets/polaris.env` and `secrets/cloudflared.token` outside version control and readable only by the deployment account.
- Use Cloudflare Tunnel or another authenticated HTTPS reverse proxy. Do not expose the application port directly to the Internet.
- Configure Cloudflare Turnstile and keep its secret key server-side.
- Use a unique, long administrator password and rotate agent tokens after suspected exposure.
- Keep the host OS, Docker, and container images patched.
- Back up the `/data` directory securely. It contains node metadata and operational metrics.
