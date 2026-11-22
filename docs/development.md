# Development

## Run it

```sh
git config core.hooksPath .githooks     # once per clone
go run ./cmd/paradise                   # http://localhost:8080, in-memory store
PORT=8090 go run ./cmd/paradise         # if 8080 is taken
```

In the Claude desktop app, `.claude/launch.json` starts the same server in
the preview pane.

With Postgres (what production uses):

```sh
docker run -d --name paradise-pg -e POSTGRES_USER=paradise -e POSTGRES_PASSWORD=paradise -e POSTGRES_DB=paradise -p 5435:5432 postgres:17-alpine
DATABASE_URL="postgres://paradise:paradise@localhost:5435/paradise?sslmode=disable" go run ./cmd/paradise
```

The local site shows a dashed box wherever `content/words.txt` has no
words yet. Production shows nothing there.

To see the room in another light, change your computer's clock, or in the
browser console run `document.getElementById("room-img").src = "/static/art/room-dusk.png"`
(phases: `latenight`, `morning`, `afternoon`, `dusk`, `night`).

## Validate before a PR

There is no staging, so check the experience locally:

- Every page at 1024×768: the bedroom, Game Boy, Xbox years, shelf, right
  now, coming soon, arcade, the skate game, guestbook, Top 8, and a 404.
- Hover every object in the room; the status bar names it.
- Sign the guestbook; try a link and a bad word in the message.
- Play a run of Half-Pipe Hero; turn sound on and off.

## Environment variables

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Listen port (Railway sets it). |
| `APP_ENV` | `development` | `production`: HSTS, indexable, no writing prompts, requires the three below. |
| `DATABASE_URL` | empty | Postgres URL. Empty uses an in-memory store (refused in production). Secret. |
| `VISITOR_SECRET` | random per boot | 32+ characters; keys the visitor hash (ADR-006). Secret. |
| `SITE_URL` | `http://localhost:PORT` | Public https address, used in the link-back code. |
| `TRUST_PROXY` | `true` in production | Read the client IP from `X-Real-IP` (Railway's edge). |

## Checks (CI runs all of these)

```sh
gofmt -l .                          # must print nothing
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@2026.2.2 ./...
go test ./...                       # Postgres tests run when TEST_DATABASE_URL is set
node --test internal/web/jstest/*.test.mjs
go run ./cmd/artgen && git status   # art must not change
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
docker build -t gamers-paradise .
```

Race tests need cgo; on Windows run them in Docker:

```sh
docker run --rm -v "$PWD":/src -w /src golang:1.27.4 go test -race ./...
```

## Art

Everything visual is drawn in `internal/art` (room, phases, hotspots,
titles, button, cursor, speaker, background, font). After a change:

```sh
go run ./cmd/artgen
```

Hotspot rectangles in `room.go` generate `static/art/hotspots.css`, so the
links always line up with the drawing.

## Deploy (Railway)

Project `gamers-paradise`, environment `production`: the `gamers-paradise`
service (Dockerfile) and `Postgres`. Merging to `main` deploys after CI
passes. Service settings are code in `.railway/railway.ts`:

```sh
cd .railway && npm install && cd ..
railway config plan  --environment production   # must say 0 to destroy
railway config apply --environment production
```

On Windows run those through PowerShell with `$env:_` pointing at the real
`railway.exe` (Git Bash overwrites `$_`):

```sh
powershell -NoProfile -Command '$env:_ = "$env:APPDATA\npm\node_modules\@railway\cli\bin\railway.exe"; & $env:_ config plan --environment production'
```
