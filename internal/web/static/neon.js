// Neon Mile: drawing the street and reading the keyboard. Rules:
// neon-core.js.
import { screen, keyboard, loop, loadBest, saveBest, W, H } from "/static/arcade.js";
import { newRace, update, place, ordinal, mph, MILE, ROAD_LEFT, ROAD_RIGHT, LANE_W } from "/static/neon-core.js";
import { tone, jingle } from "/static/sound.js";

const canvas = document.getElementById("neon");
const s = screen(canvas);
const keys = keyboard(canvas, ["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", " ", "Enter"]);
const BEST_KEY = "gp-neon-best"; // best winning time, in hundredths
const PX = 4; // screen pixels per meter
const YOU_Y = 150; // where your car sits on screen
let r = null;
let banner = null; // { text, t }
let result = null;

function start() {
  r = newRace(Date.now() >>> 0);
  result = null;
  banner = null;
}

loop((dt) => {
  const p = keys.pressed;
  const h = keys.held;
  if (!r || r.over) {
    if (p.has("Enter")) start();
  } else {
    const before = Math.ceil(-r.t);
    const ev = update(r, {
      steer: (h.has("ArrowRight") ? 1 : 0) - (h.has("ArrowLeft") ? 1 : 0),
      gas: h.has("ArrowUp"),
      brake: h.has("ArrowDown"),
      nitro: h.has(" "),
    }, dt);
    const after = Math.ceil(-r.t);
    if (r.t < 0 && after !== before) tone(440, 0.1, 0.05);
    if (r.t >= 0 && r.t - dt < 0) tone(880, 0.25, 0.06);
    for (const e of ev) {
      if (e.type === "crash") {
        banner = { text: "CRASH!", t: 0.9 };
        tone(90, 0.3, 0.08, "sawtooth");
      } else if (e.type === "nearMiss") {
        banner = { text: "NEAR MISS! +NITRO", t: 0.7 };
        tone(990, 0.05, 0.04);
      } else if (e.type === "finish") {
        result = e;
        const hundredths = Math.round(e.time * 100);
        if (e.place === 1) {
          const best = loadBest(BEST_KEY);
          if (!best || hundredths < best) saveBest(BEST_KEY, hundredths);
          jingle([523, 659, 784, 1047]);
        } else {
          jingle([440, 392, 330]);
        }
      }
    }
  }
  if (banner) banner.t -= dt;
  keys.clear();
  draw();
});

// ---------- art ----------
const CAR = [
  "..kkkkkkkkkk..",
  ".kbbbbbbbbbbk.",
  "kbbccccccccbbk",
  "kbbcwwwwwwcbbk",
  "kbbcwwwwwwcbbk",
  "kbbbbbbbbbbbbk",
  "kbbbbbbbbbbbbk",
  "kbbbbbbbbbbbbk",
  "kbbcccccccccbk",
  "kbbcwwwwwwwcbk",
  "kbbbbbbbbbbbbk",
  "kbbbbbbbbbbbbk",
  ".kbbbbbbbbbbk.",
  ".krbbbbbbbbrk.",
];

function car(x, y, body, glass = "#9ad8ff") {
  s.sprite(CAR, Math.round(x - 7), Math.round(y), { k: "#05040a", b: body, c: "#1a1a2a", w: glass, r: "#ff3030" });
}

function hash(n) {
  const x = Math.sin(n * 127.1) * 43758.5453;
  return x - Math.floor(x);
}

function city(dist) {
  s.rect(0, 0, W, H, "#0a0818");
  // buildings scroll with the road: one block every 30 m on each side
  const first = Math.floor((dist - 10) / 30);
  for (let i = first; i < first + 8; i++) {
    const y = YOU_Y - (i * 30 - dist) * PX;
    for (const side of [0, 1]) {
      const wv = 36 + Math.floor(hash(i * 3 + side) * 30);
      const x = side ? ROAD_RIGHT + 14 : ROAD_LEFT - 14 - wv;
      const hv = 30 * PX - 10;
      s.rect(x, y - hv, wv, hv, ["#140e2a", "#1a1030", "#0e1428"][Math.floor(hash(i + side * 9) * 3)]);
      for (let wy = y - hv + 6; wy < y - 6; wy += 10) {
        for (let wx = x + 4; wx < x + wv - 6; wx += 8) {
          if (hash(wx * 0.37 + wy * 1.3 + i) > 0.55) s.rect(wx, wy, 4, 5, hash(wx + wy) > 0.8 ? "#ff7ac8" : "#f6d870");
        }
      }
    }
  }
  // sidewalks
  s.rect(ROAD_LEFT - 14, 0, 14, H, "#2a2438");
  s.rect(ROAD_RIGHT, 0, 14, H, "#2a2438");
}

function road(dist) {
  s.rect(ROAD_LEFT, 0, ROAD_RIGHT - ROAD_LEFT, H, "#16141e");
  s.rect(ROAD_LEFT, 0, 2, H, "#ff3aa8");
  s.rect(ROAD_RIGHT - 2, 0, 2, H, "#3ae0ff");
  const off = (dist * PX) % 40;
  for (let l = 1; l < 4; l++) {
    const x = ROAD_LEFT + l * LANE_W - 1;
    for (let y = -40 + off; y < H; y += 40) s.rect(x, y, 2, 20, "#5a5470");
  }
  // streetlights every 40 m, with pools of light
  const first = Math.floor(dist / 40);
  for (let i = first; i < first + 6; i++) {
    const y = YOU_Y - (i * 40 - dist) * PX;
    for (const x of [ROAD_LEFT - 8, ROAD_RIGHT + 8]) {
      s.ctx.fillStyle = "rgba(255, 220, 140, 0.08)";
      s.ctx.beginPath();
      s.ctx.ellipse(x, y, 28, 18, 0, 0, Math.PI * 2);
      s.ctx.fill();
      s.rect(x - 2, y - 2, 4, 4, "#fff0b0");
    }
  }
  // the finish line
  const fy = YOU_Y - (MILE - dist) * PX;
  if (fy > -10 && fy < H) {
    for (let x = ROAD_LEFT; x < ROAD_RIGHT; x += 8) {
      s.rect(x, fy, 4, 4, "#ffffff");
      s.rect(x + 4, fy + 4, 4, 4, "#ffffff");
    }
  }
}

function headlights(x, y) {
  s.ctx.fillStyle = "rgba(255, 246, 180, 0.12)";
  s.ctx.beginPath();
  s.ctx.moveTo(x - 5, y);
  s.ctx.lineTo(x + 5, y);
  s.ctx.lineTo(x + 18, y - 60);
  s.ctx.lineTo(x - 18, y - 60);
  s.ctx.fill();
}

function cars() {
  const d = r.you.dist;
  for (const c of r.traffic) {
    const y = YOU_Y - (c.dist - d) * PX;
    if (y < -30 || y > H + 10) continue;
    car(c.x, y, ["#4a4a5a", "#5a4a3a", "#3a4a5a"][c.tone], "#2a3a4a");
  }
  for (const rv of r.rivals) {
    const y = YOU_Y - (rv.dist - d) * PX;
    if (y < -30 || y > H + 10) continue;
    headlights(rv.x, y);
    car(rv.x, y, rv.color);
  }
  const y = r.you;
  headlights(y.x, YOU_Y);
  if (y.crashT > 0 && Math.floor(y.crashT * 12) % 2 === 0) return; // blink after a crash
  car(y.x, YOU_Y, "#3a78ff");
  if (keys.held.has(" ") && y.nitro > 0) {
    s.rect(y.x - 4, YOU_Y + 14, 2, 4 + Math.round(Math.random() * 4), "#3ae0ff");
    s.rect(y.x + 2, YOU_Y + 14, 2, 4 + Math.round(Math.random() * 4), "#3ae0ff");
  }
}

function hud() {
  s.rect(0, 0, W, 12, "rgba(5,4,10,0.8)");
  s.text(ordinal(place(r)), 4, 3, "#fff27a");
  s.text(`${mph(r.you.speed)} MPH`, W / 2, 3, "#ffffff", 1, "center");
  s.text(Math.max(0, r.t).toFixed(1), W - 4, 3, "#a0f0ff", 1, "right");
  // nitro tank
  s.text("NITRO", 4, 20, "#3ae0ff");
  s.rect(4, 30, 50, 6, "#222");
  s.rect(5, 31, Math.round(48 * r.you.nitro), 4, "#3ae0ff");
  // the mile, with everyone on it
  const bx = W - 10;
  s.rect(bx, 20, 3, 160, "#2a2438");
  const dot = (dist, c) => s.rect(bx - 2, 178 - Math.round(Math.min(1, dist / MILE) * 156), 7, 3, c);
  for (const rv of r.rivals) dot(rv.dist, rv.color);
  dot(r.you.dist, "#3a78ff");
  if (r.t < 0) s.shadowText(String(Math.ceil(-r.t)), W / 2, 70, "#fff27a", 3, "center");
  else if (r.t < 0.8) s.shadowText("GO!", W / 2, 70, "#7aff5a", 3, "center");
  if (banner && banner.t > 0) s.shadowText(banner.text, W / 2, 100, banner.text === "CRASH!" ? "#ff5a5a" : "#3ae0ff", 1, "center");
}

function draw() {
  s.begin();
  const dist = r ? r.you.dist : (performance.now() / 1000) * 30; // the title screen cruises
  city(dist);
  road(dist);
  if (!r) {
    headlights(laneCenterX(), YOU_Y);
    car(laneCenterX(), YOU_Y, "#3a78ff");
    const best = loadBest(BEST_KEY);
    s.panel("NEON MILE", ["ONE MILE. FOUR CARS. CITY TRAFFIC.", "ARROWS: DRIVE   SPACE: NITRO", "NEAR MISSES REFILL NITRO", best ? `BEST WIN: ${(best / 100).toFixed(2)}S   ENTER: RACE` : "ENTER: RACE"]);
    return;
  }
  cars();
  hud();
  if (result) {
    const best = loadBest(BEST_KEY);
    s.panel(result.place === 1 ? "YOU WIN!" : `${ordinal(result.place)} PLACE`, [
      `TIME ${result.time.toFixed(2)}S`,
      best ? `BEST WIN ${(best / 100).toFixed(2)}S` : "WIN ONE TO SET A BEST TIME",
      "",
      "ENTER: RACE AGAIN",
    ]);
  }
}

function laneCenterX() {
  return ROAD_LEFT + LANE_W * 1.5;
}

