# Decisions

Short records of choices worth remembering. Newest last.

## ADR-001: Bedroom hub (2026-11-22)

The landing page is a pixel-art bedroom at night; every object links to a
page dressed in its own era. The laptop links back to the bedroom (it shows
the room on its screen). The beanbag opens the arcade. The door is the
guestbook. Design discovery is logged in the workspace notes.

## ADR-002: Desktop only, 1024×768 (2026-11-22)

True Web 1.5: fixed 1000px layout, no viewport meta, no mobile layout.
Phones render the desktop page zoomed out, like 2006. Code underneath stays
modern: semantic HTML, CSS, ES modules.

## ADR-003: Go + Postgres on Railway, no logins (2026-11-22)

Same stack as the owner's other projects. One Railway service deploys `main`;
Postgres holds only the hit counter and guestbook. No staging: local dev
for validation, CI and protected `main` before prod. In-memory store for
local dev; refused in production.

## ADR-004: All art is drawn in code (2026-11-22)

`internal/art` draws the room, titles, button, cursor and font; `artgen`
writes PNGs, and a test fails if the committed files drift. Franchises are
referenced through objects and motifs (a d20, a laser sword, a pomegranate,
a setlist reading "One"), never characters, logos or sprites. Personal
writing lives in `content/words.txt` and is the owner's alone.

## ADR-005: Strict CSP, no third parties (2026-11-22)

`default-src 'none'`, same-origin scripts, styles and images only, no
inline code, Trusted Types required, no fetches or frames. Sound is
synthesized with Web Audio, so there are no media files either. Tests check
every page for inline script, style and handlers.

## ADR-006: Counting visitors without storing IPs (2026-11-22)

Unique visitors per day. A visitor ID is HMAC(secret, UTC day, IP, user
agent), so it rotates daily and can't be reversed without the secret. Rows
are deleted after a day; the counter keeps only a total. Obvious bots
aren't counted. Behind Railway the IP comes from `X-Real-IP` (only trusted
in production).

## ADR-007: Guestbook goes live with auto-moderation (2026-11-22)

No approval queue. Each signature is normalized, links in messages are
stripped, a word list masks slurs and profanity, and spam shapes are
refused. Limits: 5 posts per 10 minutes per IP (in memory), 3 signatures
per visitor per day, no duplicate messages within a day, a honeypot field,
same-origin POSTs only. An optional homepage link is shown with
`rel="ugc nofollow noopener"`. The owner removes entries with `cmd/gbmod`.

## ADR-008: CI and repo hardening (2026-11-22)

Public repo. `main` requires a PR and passing checks (Lint, Tests,
Security, Docker image, CodeQL), with no force-pushes or deletion. Actions
are pinned to commit SHAs, workflow tokens are read-only, checkout drops
credentials. Security job: govulncheck, gitleaks over the full history,
zizmor for workflow issues. CodeQL covers Go, JavaScript and the workflows.
Dependency review on PRs, Dependabot weekly. GitHub secret scanning with
push protection is on.

## ADR-009: Railway settings as code (2026-11-22)

`.railway/railway.ts` defines the service (Dockerfile build, `/healthz`,
deploy `main` only after GitHub checks pass). Postgres and variables are
managed in Railway. Railway retires `railway.json` on 2026-12-01.

## ADR-010: Half-Pipe Hero first (2026-11-22)

The beanbag's select screen lists four games; the skate trick combo game
ships first and the others show as coming soon. Game rules live in
`static/skate-core.js` (no DOM) so they're tested with `node --test`.

## ADR-011: All four games ship (2026-11-22)

The beanbag's arcade has four original games, each split into a rules
module with no DOM (`*-core.js`, tested with `node --test`) and a canvas
renderer, sharing `static/arcade.js` (pixel surface, font, keys, best
scores in the browser) and `static/rng.js` (seeded randomness):

- **Half-Pipe Hero**: skate trick combos.
- **Fret Night**: five-lane rhythm game to "Midnight Riff", an original
  chiptune the code writes from a seed (A minor, 140 BPM, verse/chorus).
  Music is scheduled on the Web Audio clock so hits line up with it.
  Starting a song turns sound on. Lane colors and keys are its own, not a
  copy of any guitar game's.
- **Status Effect**: a Lv 12 Thornling against a Lv 42 Magmaul, both
  original creatures. Tuned by simulation: plain attacks never win, one
  status alone almost never wins, random play wins about half the time,
  and combining status, Sap Siphon and heals wins over 90%.
- **Neon Mile**: a one-mile night street race against three rivals through
  traffic. A clean run with nitro wins; a couple of crashes lose.
