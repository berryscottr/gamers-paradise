// Half-Pipe Hero: drawing and keyboard. The rules live in skate-core.js.
import { GLYPHS } from "/static/art/font.js";
import { newRun, start, update, R, K, FLIP_TIME, FLIPS } from "/static/skate-core.js";
import { tone, jingle } from "/static/sound.js";

const canvas = document.getElementById("skate");
const ctx = canvas.getContext("2d");
ctx.imageSmoothingEnabled = false;
const W = 320;
const H = 200;
const CX = 160; // pipe center
const CY = 80; // coping height
const BEST_KEY = "gp-skate-best";

function loadBest() {
  try {
    return Number(localStorage.getItem(BEST_KEY)) || 0;
  } catch {
    return 0;
  }
}
function saveBest(v) {
  try {
    localStorage.setItem(BEST_KEY, String(v));
  } catch {
    // storage blocked: the best score lasts until the tab closes
  }
}

let s = newRun(loadBest());
let paused = false;
let popup = null; // { text, sub, t }

// ---------- keyboard ----------
const held = new Set();
const pressed = new Set();
const GAME_KEYS = ["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", " ", "z", "x", "a", "d", "Enter", "p"];
function key(e) {
  return e.key.length === 1 ? e.key.toLowerCase() : e.key;
}
window.addEventListener("keydown", (e) => {
  const k = key(e);
  if (!GAME_KEYS.includes(k)) return;
  if (document.activeElement !== canvas && document.activeElement !== document.body) return;
  e.preventDefault(); // keep arrows and space from scrolling the page
  if (!held.has(k)) pressed.add(k);
  held.add(k);
});
window.addEventListener("keyup", (e) => held.delete(key(e)));
window.addEventListener("blur", () => held.clear());
canvas.addEventListener("pointerdown", () => canvas.focus());
canvas.focus();

function arrow() {
  if (held.has("ArrowLeft")) return "left";
  if (held.has("ArrowRight")) return "right";
  if (held.has("ArrowUp")) return "up";
  if (held.has("ArrowDown")) return "down";
  return null;
}

function readInput() {
  const dir = arrow();
  return {
    pump: pressed.has("ArrowDown") || pressed.has(" "),
    flip: pressed.has("z") ? dir || "left" : null,
    grab: held.has("x") ? dir || "left" : null,
    spin: (held.has("d") ? 1 : 0) - (held.has("a") ? 1 : 0),
  };
}

// ---------- loop ----------
let last = performance.now();
function frame(now) {
  const dt = Math.min(0.05, (now - last) / 1000);
  last = now;
  if (pressed.has("Enter") && (s.mode === "ready" || s.mode === "over")) {
    start(s);
    popup = null;
    jingle([392, 523, 659]);
  } else if (pressed.has("p") && s.mode !== "ready" && s.mode !== "over") {
    paused = !paused;
  }
  if (!paused) {
    for (const e of update(s, readInput(), dt)) sound(e);
  }
  if (popup) popup.t -= dt;
  pressed.clear();
  draw();
  requestAnimationFrame(frame);
}
requestAnimationFrame(frame);

function sound(e) {
  switch (e.type) {
    case "pump":
      tone(220, 0.05, 0.04, "triangle");
      break;
    case "launch":
      tone(330, 0.08, 0.03, "sawtooth");
      break;
    case "trick":
      tone(660, 0.05, 0.04);
      break;
    case "land":
      if (e.points > 0) {
        popup = { text: e.name, sub: `+${e.points}  X${e.streak}`, t: 1.6 };
        jingle([523, 784]);
      } else {
        tone(196, 0.05, 0.03, "triangle");
      }
      break;
    case "bail":
      popup = { text: "BAILED", sub: "STREAK LOST", t: 1.4 };
      tone(110, 0.3, 0.06, "sawtooth");
      break;
    case "over":
      saveBest(e.best);
      jingle([659, 523, 392, 523]);
      break;
  }
}

// ---------- drawing ----------
function rect(x, y, w, h, c) {
  ctx.fillStyle = c;
  ctx.fillRect(Math.round(x), Math.round(y), w, h);
}

function text(str, x, y, c, scale = 1, align = "left") {
  const w = (str.length * 6 - 1) * scale;
  let cx = align === "center" ? Math.round(x - w / 2) : align === "right" ? x - w : x;
  ctx.fillStyle = c;
  for (const ch of str.toUpperCase()) {
    const g = GLYPHS[ch];
    if (g) {
      g.forEach((row, j) => {
        for (let i = 0; i < 5; i++) {
          if (row & (16 >> i)) ctx.fillRect(cx + i * scale, y + j * scale, scale, scale);
        }
      });
    }
    cx += 6 * scale;
  }
}

function shadowText(str, x, y, c, scale, align) {
  text(str, x + scale, y + scale, "#1a0a2a", scale, align);
  text(str, x, y, c, scale, align);
}

function scenery() {
  const bands = ["#3a6ad8", "#4a7ee0", "#5e92e8", "#76a8ee", "#92bef2", "#aed2f6"];
  bands.forEach((c, i) => rect(0, i * 22, W, 22, c));
  rect(52, 30, 18, 18, "#fff2b0");
  rect(48, 34, 26, 10, "#fff2b0");
  for (const [x, h] of [[0, 30], [40, 44], [90, 26], [150, 38], [210, 30], [270, 46]]) {
    rect(x, 132 - h, 60, h, "#6a8ac8");
  }
  rect(0, 120, W, 80, "#8a9aa8"); // concrete
  rect(0, 170, W, 30, "#6a7a88");
}

function pipe() {
  for (let x = CX - R; x <= CX + R; x++) {
    const dy = Math.sqrt(Math.max(0, R * R - (x - CX) * (x - CX)));
    const top = Math.round(CY + dy);
    rect(x, top, 1, 192 - top, "#c89a5a");
    rect(x, top, 1, 2, "#f0d0a0");
  }
  rect(18, CY, CX - R - 18, 112, "#a87a40"); // decks
  rect(CX + R + 1, CY, W - (CX + R + 1) - 18, 112, "#a87a40");
  rect(18, CY, CX - R - 18, 3, "#e0b880");
  rect(CX + R + 1, CY, W - (CX + R + 1) - 18, 3, "#e0b880");
  for (const x of [22, 44, 280, 298]) rect(x, CY - 14, 2, 14, "#555"); // railings
  rect(20, CY - 15, CX - R - 22, 2, "#777");
  rect(CX + R + 4, CY - 15, W - (CX + R + 4) - 20, 2, "#777");
  rect(CX - R - 1, CY - 1, 3, 3, "#ddd"); // coping
  rect(CX + R - 1, CY - 1, 3, 3, "#ddd");
  rect(0, 190, W, 10, "#4a5a68");
}

const RIDER = {
  ride: [
    "....hh....",
    "...hssh...",
    "...ssss...",
    "....ss....",
    "..oooooo..",
    ".s.oooo.s.",
    "...oooo...",
    "...bbbb...",
    "...b..b...",
    "..bb..bb..",
    "..kk..kk..",
  ],
  crouch: [
    "..........",
    "..........",
    "....hh....",
    "...hssh...",
    "...ssss...",
    ".soooooos.",
    "...oooo...",
    "..bbbbbb..",
    ".bb....bb.",
    ".kk....kk.",
    "..........",
  ],
  grab: [
    "..........",
    "....hh....",
    "...hssh...",
    "...ssss...",
    "..oooooo..",
    "..oooooos.",
    "...bbbb.s.",
    "..bb..bbs.",
    "..kk..kk..",
    "..........",
    "..........",
  ],
  bail: [
    "..........",
    "..........",
    "..........",
    "..........",
    "..........",
    "..........",
    "..........",
    "hh.ooooo..",
    "ssooooobbk",
    "ss.s...bbk",
    "..........",
  ],
};
const COLORS = { h: "#5a3a1a", s: "#ffd8a8", o: "#e8803a", b: "#2a4a9a", k: "#1a1a1a" };

function sprite(rows, x, y) {
  rows.forEach((row, j) => {
    for (let i = 0; i < row.length; i++) {
      const c = COLORS[row[i]];
      if (c) rect(x + i, y + j, 1, 1, c);
    }
  });
}

function board(x, y, flip) {
  // flip in [0, 1): the board turns over once
  const squash = Math.cos(flip * Math.PI * 2);
  const h = Math.max(1, Math.round(Math.abs(squash) * 2));
  rect(x - 7, y, 14, h, squash >= 0 ? "#3a2a6a" : "#e8803a");
  if (Math.abs(squash) > 0.5) {
    rect(x - 5, y + h, 2, 1, "#eee");
    rect(x + 3, y + h, 2, 1, "#eee");
  }
}

function rider() {
  if (s.mode === "bail") {
    sprite(RIDER.bail, CX - 5, 179);
    const t = Date.now() / 120;
    for (let i = 0; i < 3; i++) {
      rect(CX - 2 + Math.round(Math.cos(t + i * 2) * 8), 180 + Math.round(Math.sin(t + i * 2) * 3), 2, 2, "#fff27a");
    }
    return;
  }
  ctx.save();
  if (s.mode === "air") {
    const a = s.air;
    const x = CX + a.side * (R - 8);
    const y = CY - a.y - 4;
    ctx.translate(x, y);
    ctx.scale(Math.cos((a.spin * Math.PI) / 180) || 0.1, 1); // spins read as turning in place
    sprite(a.grab ? RIDER.grab : RIDER.ride, -5, -12);
    const flip = a.flipT > 0 ? 1 - a.flipT / FLIP_TIME : 0;
    board(0, -1, flip);
  } else {
    const x = CX + R * Math.sin(s.theta);
    const y = CY + R * Math.cos(s.theta);
    ctx.translate(x, y);
    ctx.rotate(-s.theta);
    const low = Math.abs(s.theta) < 0.35 && held.size > 0 && (held.has("ArrowDown") || held.has(" "));
    sprite(low ? RIDER.crouch : RIDER.ride, -5, -14);
    board(0, -3, 0);
  }
  ctx.restore();
}

function hud() {
  rect(0, 0, W, 14, "rgba(10,8,30,0.75)");
  text(`SCORE ${s.score}`, 4, 4, "#fff27a");
  text(`${Math.ceil(s.time)}`, W / 2, 4, s.time < 10 ? "#ff5a5a" : "#ffffff", 1, "center");
  text(`BEST ${s.best}`, W - 4, 4, "#a0f0ff", 1, "right");
  if (s.streak > 0) text(`STREAK X${s.streak}`, 4, 18, "#ffffff");
  if (s.mode === "air" && s.air.tricks.length) {
    text(s.air.tricks.map((t) => t.name).join(" + "), W / 2, 30, "#ffffff", 1, "center");
  }
  if (popup && popup.t > 0) {
    shadowText(popup.text, W / 2, 44, "#fff27a", 1, "center");
    shadowText(popup.sub, W / 2, 56, "#ffffff", 1, "center");
  }
  const meter = Math.max(0, Math.min(1, (s.E - 2) / 10));
  rect(W - 44, 18, 40, 5, "#222");
  rect(W - 43, 19, Math.round(38 * meter), 3, s.E > K ? "#7aff5a" : "#ffb84a"); // green: enough to launch
  text("SPEED", W - 80, 17, "#ffffff");
}

function screen(title, lines) {
  rect(40, 50, 240, 100, "rgba(10,8,30,0.85)");
  rect(40, 50, 240, 2, "#8a7ae8");
  rect(40, 148, 240, 2, "#8a7ae8");
  shadowText(title, W / 2, 62, "#fff27a", 2, "center");
  lines.forEach((l, i) => text(l, W / 2, 92 + i * 12, "#ffffff", 1, "center"));
}

function draw() {
  ctx.setTransform(2, 0, 0, 2, 0, 0); // 320x200 art pixels on a 640x400 canvas
  ctx.imageSmoothingEnabled = false;
  scenery();
  pipe();
  rider();
  hud();
  if (s.mode === "ready") {
    screen("HALF-PIPE HERO", ["PRESS ENTER TO START", "PUMP: DOWN OR SPACE", `FLIP: Z  GRAB: X  SPIN: A D`, `TRY: Z + LEFT FOR A ${FLIPS.left.name}`]);
  } else if (s.mode === "over") {
    screen("TIME UP", [`SCORE ${s.score}`, `BEST ${s.best}`, "", "PRESS ENTER TO GO AGAIN"]);
  } else if (paused) {
    screen("PAUSED", ["PRESS P TO KEEP SKATING"]);
  }
}
