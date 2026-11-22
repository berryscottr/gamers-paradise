# Gamer's Paradise

*Spend a day in my childhood.*

A small corner of the indie web, built like it's 2006: a pixel-art bedroom
where every object is a link. The Game Boy Color on the nightstand, the Xbox
and the CRT at the foot of the bed, the shelf of games and minis, the poster
on the wall, the guestbook taped to the door, and a beanbag you can sit in
to play a game. The room's light follows your clock.

Best viewed at 1024×768 on a desktop. Phones get the real 2006 experience.

## What's inside

- **The bedroom** (`/`): drawn in code (`internal/art`), five lighting phases
  picked by the visitor's clock. The laptop on the desk shows the room
  itself.
- **Pages dressed in their own era**: Game Boy, the Xbox years, the shelf,
  right now, and a box of upcoming features.
- **Arcade**: four original games on the beanbag: *Half-Pipe Hero* (skate
  trick combos), *Fret Night* (a rhythm game to an original chiptune),
  *Status Effect* (win a battle with status effects, not grinding) and
  *Neon Mile* (a night street race).
- **Guestbook**: no logins; signatures go live with automatic moderation.
- **Hit counter**: unique visitors per day, counted without storing IPs.
- **Top 8 and an 88×31 button** for linking back.

## Run it

```sh
go run ./cmd/paradise      # http://localhost:8080
```

See [docs/development.md](docs/development.md) for everything else, and
[docs/security.md](docs/security.md) for how the site and CI are locked down.

## Credits

All art, code and writing are original. Game titles are mentioned with love
and belong to their owners; nothing on the site is affiliated with or
endorsed by them.
