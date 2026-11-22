# Guestbook moderation

Signatures go live right away (ADR-007). If one gets through that
shouldn't, hide it from your own terminal. There's no admin page.

## Hide a signature

Railway's Postgres has a public URL for connecting from outside. `railway
run` injects the Postgres service's variables into a command without
printing them:

```sh
railway run --service Postgres sh -c 'DATABASE_URL="$DATABASE_PUBLIC_URL" go run ./cmd/gbmod list'
railway run --service Postgres sh -c 'DATABASE_URL="$DATABASE_PUBLIC_URL" go run ./cmd/gbmod delete 17'
```

`list` shows the newest 50 with their numbers; `delete` hides one (it stays
in the database, marked deleted).

## Tuning the filter

- Word lists: `internal/guestbook/words.go` (`anywhere` for words masked even
  inside other words, `whole` for standalone words).
- Link detection: `linkRE` in `internal/guestbook/guestbook.go`.
- Limits: `signaturesPerDay` in `internal/web/pages.go`, the burst limiter
  in `web.New`.

Add a test case in `guestbook_test.go` for anything you change.
