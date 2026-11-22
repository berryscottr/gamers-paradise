# Security

The site has no accounts and no admin page, and stores only a visitor total
and guestbook signatures. Here's how it's protected.

## On the public internet

- **Headers on every response**: a strict Content-Security-Policy
  (`default-src 'none'`, same-origin scripts/styles/images, no inline code,
  Trusted Types, no frames, `form-action 'self'`), HSTS, `nosniff`,
  `X-Frame-Options: DENY`, COOP and CORP `same-origin` (except the 88×31
  button, which other sites may show), a restrictive Permissions-Policy.
- **Only GET, HEAD and POST**; request bodies capped at 16 KB; server
  timeouts on headers, reads, writes and idle connections.
- **No third parties**: no analytics, fonts, embeds or CDNs.
- **Guestbook** (ADR-007): same-origin POSTs only (`Sec-Fetch-Site` or
  `Origin`), rate limits, honeypot, link stripping, word masking, length
  limits, and output escaping by `html/template`. Visitor homepages must be
  http(s) URLs without credentials and are marked `ugc nofollow noopener`.
- **Privacy** (ADR-006): no IPs stored; visitor IDs are a keyed hash that
  rotates daily and is deleted after a day.
- **SQL** is parameterized everywhere (`internal/store`).
- **Container**: static binary on distroless, running as non-root.
- **Dev sites** send `X-Robots-Tag: noindex` and disallow crawling.

## In CI and the repo

- `main` is protected: PRs only, required checks, no force-push or delete.
- Actions pinned to full commit SHAs; `permissions: contents: read` by
  default; `persist-credentials: false` on checkout.
- govulncheck (Go modules and stdlib), gitleaks (full history), zizmor
  (workflow injection and permissions), CodeQL (Go, JavaScript, Actions),
  dependency review on PRs, Dependabot weekly.
- GitHub secret scanning and push protection.

## Secrets

`DATABASE_URL` and `VISITOR_SECRET` live only in Railway's variables.
Never commit them, print them, or put them in notes. The app never logs
the database URL, even when it's malformed.
