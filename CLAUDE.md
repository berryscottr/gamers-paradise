# Gamer's Paradise

A retro personal site: a Web 1.5 nostalgia trip, not a portfolio.
Tagline "Spend a day in my childhood". Handle: Sir Games-A-Lot. Public repo.

## Rules

- **Nothing about work or productivity.** It's about growing up with games.
- **The owner writes every word of personal copy** in `content/words.txt`.
  Don't write it for them. Add a section with a `?` prompt instead and ask
  them to fill it in. UI labels, game rules and the features list are fine
  to write.
- **No family members** on the site or in public files.
- **Original art only** (ADR-004). Reference franchises through objects and
  motifs; no game characters, logos, sprites, music or copied UIs.
- **Desktop only, 1024×768, true 2006** (ADR-002). No viewport meta, no
  mobile layout. Modern, sound code underneath.
- **Strict CSP** (ADR-005): no inline script, style or event handlers; no
  third-party requests. Tests enforce it.
- **No logins, no admin page.** Guestbook moderation is `cmd/gbmod`.

## Workflow

- Feature branch → PR into `main` → CI green → merge. `main` is protected
  (PRs and checks required, no force-push) and Railway deploys it only after
  CI passes. There is no staging: the owner validates locally first.
- Enable the hook once per clone: `git config core.hooksPath .githooks`.
- Art changes: edit `internal/art`, run `go run ./cmd/artgen`, commit the
  output. CI fails on drift.
- Record decisions in `docs/decisions.md`. Keep docs current.

## Map

| Path | What |
|---|---|
| `cmd/paradise` | the server |
| `cmd/artgen` | writes `internal/web/static/art` from `internal/art` |
| `cmd/gbmod` | guestbook list/delete from a terminal |
| `content/words.txt` | all personal writing, with prompts |
| `internal/art` | pixel art in code: room, phases, hotspots, titles, font |
| `internal/web` | handlers, templates, CSS, JS (`static/`), JS tests (`jstest/`) |
| `internal/web/static/*-core.js` | game rules, no DOM (skate, fret, battle, neon); renderers beside them |
| `internal/guestbook` | signature cleaning and auto-moderation |
| `internal/visitor` | daily-rotating visitor hash, bot check |
| `internal/store` | Postgres and in-memory store, migrations |
| `.railway/railway.ts` | Railway service settings as code |

Docs: `docs/development.md`, `docs/security.md`, `docs/moderation.md`,
`docs/decisions.md`. Workspace notes: `../notes/projects/indie-site.md`.
