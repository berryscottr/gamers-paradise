// Fret Night: playing the song and drawing the stage. Rules: fret-core.js.
import { screen, keyboard, loop, loadBest, saveBest, W } from "/static/arcade.js";
import { newGame, press, update, multiplier, stars, midiToHz, LANES, SONG_TITLE, BEAT, LEAD_IN } from "/static/fret-core.js";
import { enable, output } from "/static/sound.js";

const canvas = document.getElementById("fret");
const s = screen(canvas);
const KEYS = ["a", "s", "d", "f", "g"];
const keys = keyboard(canvas, [...KEYS, "Enter", "ArrowLeft", "ArrowRight", "Escape"]);
const COLORS = ["#3ae0ff", "#ff3aa8", "#a4f04a", "#ffb83a", "#a07aff"];
const LANE_W = 28;
const LEFT = Math.round(W / 2 - (LANES * LANE_W) / 2);
const HIT_Y = 168;
const TOP_Y = 16;

let difficulty = "expert";
let g = null; // current game, null on the title screen
let clock = null; // { ctx, start } or { perf, start }
let scheduled = 0;
let schedule = [];
const flashes = new Array(LANES).fill(0);
let callout = null; // { text, t }
const bestKey = () => `gp-fret-best-${difficulty}`;

function now() {
  if (!clock) return 0;
  if (clock.ctx) return clock.ctx.currentTime - clock.start;
  return performance.now() / 1000 - clock.start;
}

function start() {
  g = newGame(difficulty);
  const ctx = enable(); // the player chose to start music: turn sound on
  clock = ctx ? { ctx, start: ctx.currentTime + 0.3 } : { start: performance.now() / 1000 + 0.3 };
  schedule = [
    ...g.song.lead.map((n) => ({ ...n, voice: "lead" })),
    ...g.song.bass.map((n) => ({ ...n, voice: "bass" })),
    ...g.song.drums.map((n) => ({ ...n, voice: n.kind })),
  ].sort((a, b) => a.t - b.t);
  scheduled = 0;
  callout = null;
}

// ---------- music: schedule a little ahead on the audio clock ----------
function scheduleAhead() {
  const ctx = output();
  if (!clock || !clock.ctx || !g) return;
  const horizon = now() + 0.25;
  while (scheduled < schedule.length && schedule[scheduled].t < horizon) {
    const ev = schedule[scheduled++];
    if (ctx && g.status === "playing") play(ctx, ev, clock.start + ev.t);
  }
}

function play(ctx, ev, at) {
  if (at < ctx.currentTime) return;
  const out = ctx.createGain();
  out.connect(ctx.destination);
  if (ev.voice === "lead" || ev.voice === "bass") {
    const o = ctx.createOscillator();
    o.type = ev.voice === "lead" ? "square" : "triangle";
    o.frequency.value = midiToHz(ev.midi);
    const vol = ev.voice === "lead" ? 0.045 : 0.09;
    out.gain.setValueAtTime(vol, at);
    out.gain.exponentialRampToValueAtTime(0.0001, at + ev.dur);
    o.connect(out);
    o.start(at);
    o.stop(at + ev.dur + 0.02);
    return;
  }
  if (ev.voice === "kick") {
    const o = ctx.createOscillator();
    o.frequency.setValueAtTime(140, at);
    o.frequency.exponentialRampToValueAtTime(40, at + 0.12);
    out.gain.setValueAtTime(0.25, at);
    out.gain.exponentialRampToValueAtTime(0.0001, at + 0.15);
    o.connect(out);
    o.start(at);
    o.stop(at + 0.16);
    return;
  }
  // hats, snares and count-in clicks are filtered noise
  const len = ev.voice === "snare" ? 0.12 : 0.03;
  const buf = ctx.createBuffer(1, Math.ceil(ctx.sampleRate * len), ctx.sampleRate);
  const data = buf.getChannelData(0);
  for (let i = 0; i < data.length; i++) data[i] = Math.random() * 2 - 1;
  const src = ctx.createBufferSource();
  src.buffer = buf;
  const f = ctx.createBiquadFilter();
  f.type = ev.voice === "snare" ? "bandpass" : "highpass";
  f.frequency.value = ev.voice === "snare" ? 1800 : 7000;
  const vol = { snare: 0.12, hat: 0.03, click: 0.08 }[ev.voice] || 0.03;
  out.gain.setValueAtTime(vol, at);
  out.gain.exponentialRampToValueAtTime(0.0001, at + len);
  src.connect(f).connect(out);
  src.start(at);
}

// ---------- loop ----------
loop((dt) => {
  const p = keys.pressed;
  if (!g || g.status !== "playing") {
    if (p.has("ArrowLeft") || p.has("ArrowRight")) difficulty = difficulty === "expert" ? "normal" : "expert";
    if (p.has("Enter")) start();
  } else {
    const t = now();
    if (p.has("Escape")) g.status = "failed";
    KEYS.forEach((k, lane) => {
      if (!p.has(k)) return;
      const r = press(g, lane, t);
      flashes[lane] = 0.12;
      if (r.type === "perfect" || r.type === "good") callout = { text: r.type === "perfect" ? "PERFECT" : "GOOD", t: 0.4 };
    });
    for (const e of update(g, t)) {
      if (e.type === "done" && g.score > loadBest(bestKey())) saveBest(bestKey(), g.score);
    }
    scheduleAhead();
  }
  for (let i = 0; i < LANES; i++) flashes[i] = Math.max(0, flashes[i] - dt);
  if (callout) callout.t -= dt;
  keys.clear();
  draw();
});

// ---------- drawing ----------
function stage(t) {
  s.rect(0, 0, W, 200, "#0a0618");
  const beat = t > 0 ? (t % BEAT) / BEAT : 1;
  const pulse = 1 - beat;
  // spotlights pulse on the beat
  for (const [x, c] of [[40, "rgba(255,58,168,"], [280, "rgba(58,224,255,"]]) {
    s.ctx.fillStyle = c + (0.08 + pulse * 0.12) + ")";
    s.ctx.beginPath();
    s.ctx.moveTo(x - 4, 0);
    s.ctx.lineTo(x + 4, 0);
    s.ctx.lineTo(x + 40 * (x < 160 ? 1 : -1) + 30, 200);
    s.ctx.lineTo(x + 40 * (x < 160 ? 1 : -1) - 30, 200);
    s.ctx.fill();
  }
  // the crowd, bobbing on the beat
  for (let x = 0; x < W; x += 9) {
    const h = 10 + ((x * 7) % 6) + (g && g.meter > 30 ? Math.round(pulse * 2) * ((x / 9) % 2) : 0);
    s.rect(x, 200 - h, 7, h, "#160c2c");
    s.rect(x + 2, 197 - h, 3, 3, "#160c2c");
  }
}

function highway(t) {
  s.rect(LEFT - 3, TOP_Y - 4, LANES * LANE_W + 6, HIT_Y - TOP_Y + 18, "#05030c");
  for (let i = 0; i < LANES; i++) {
    const x = LEFT + i * LANE_W;
    s.rect(x, TOP_Y, LANE_W - 1, HIT_Y - TOP_Y + 10, "#120a26");
    s.rect(x + LANE_W - 1, TOP_Y, 1, HIT_Y - TOP_Y + 10, "#2a1a4a");
  }
  // beat lines scroll toward the player
  const span = lookahead();
  for (let b = Math.ceil((t - 0.2) / BEAT); b * BEAT < t + span; b++) {
    const y = noteY(b * BEAT, t, span);
    if (y >= TOP_Y && y <= HIT_Y + 6) s.rect(LEFT, y, LANES * LANE_W - 1, 1, b % 4 === 0 ? "#3a2a6a" : "#22163e");
  }
  // targets and keys
  for (let i = 0; i < LANES; i++) {
    const x = LEFT + i * LANE_W;
    const lit = flashes[i] > 0 || keys.held.has(KEYS[i]);
    s.rect(x + 2, HIT_Y - 2, LANE_W - 5, 6, lit ? COLORS[i] : "#3a2a5a");
    s.rect(x + 4, HIT_Y, LANE_W - 9, 2, "#05030c");
    s.text(KEYS[i], x + LANE_W / 2 - 3, HIT_Y + 10, lit ? "#ffffff" : COLORS[i]);
  }
}

function lookahead() {
  return difficulty === "expert" ? 1.5 : 2;
}

function noteY(nt, t, span) {
  return HIT_Y - ((nt - t) / span) * (HIT_Y - TOP_Y);
}

function notes(t) {
  const span = lookahead();
  for (let i = g.next; i < g.notes.length; i++) {
    const n = g.notes[i];
    if (n.t > t + span) break;
    if (n.judged === "perfect" || n.judged === "good") continue;
    const y = noteY(n.t, t, span);
    const x = LEFT + n.lane * LANE_W;
    s.rect(x + 3, y - 3, LANE_W - 7, 6, n.judged === "miss" ? "#444" : COLORS[n.lane]);
    s.rect(x + 5, y - 2, LANE_W - 11, 1, "#ffffff");
  }
}

function hud(t) {
  s.text("SCORE", 8, 8, "#a0a0c0");
  s.text(g.score, 8, 18, "#fff27a");
  s.text("STREAK", 8, 34, "#a0a0c0");
  s.text(g.streak, 8, 44, "#ffffff");
  s.shadowText(`X${multiplier(g.streak)}`, 8, 60, COLORS[multiplier(g.streak)], 2);
  // crowd meter
  s.text("CROWD", W - 40, 8, "#a0a0c0");
  s.rect(W - 34, 18, 12, 102, "#222");
  const h = Math.round(g.meter);
  const c = g.meter > 60 ? "#7aff5a" : g.meter > 30 ? "#ffd84a" : "#ff4a4a";
  s.rect(W - 33, 19 + 100 - h, 10, h, c);
  s.text(SONG_TITLE, W - 4, 186, "#6a5a9a", 1, "right");
  if (t < LEAD_IN) {
    const left = Math.ceil((LEAD_IN - t) / BEAT); // count-in: 4, 3, 2, 1
    if (left >= 1 && left <= 4) s.shadowText(String(left), W / 2, 80, "#fff27a", 3, "center");
  }
  if (callout && callout.t > 0) s.shadowText(callout.text, W / 2, 132, "#ffffff", 1, "center");
}

function draw() {
  s.begin();
  const t = g ? now() : 0;
  stage(t);
  highway(t);
  if (!g) {
    titleCard();
    return;
  }
  notes(t);
  hud(t);
  if (g.status === "failed") {
    s.panel("THE CROWD LEFT", [`SCORE ${g.score}   BEST STREAK ${g.bestStreak}`, "", "LEFT/RIGHT: DIFFICULTY", `ENTER: TRY AGAIN ON ${difficulty.toUpperCase()}`]);
  } else if (g.status === "done") {
    s.panel("ENCORE!", [`SCORE ${g.score}   BEST ${loadBest(bestKey())}`, `HIT ${g.hits}/${g.notes.length}   BEST STREAK ${g.bestStreak}`, `${stars(g)} OF 5 STARS`, "ENTER: PLAY AGAIN"]);
  }
}

function titleCard() {
  s.panel("FRET NIGHT", [
    `SONG: ${SONG_TITLE.toUpperCase()}`,
    `DIFFICULTY: < ${difficulty.toUpperCase()} >`,
    `BEST: ${loadBest(bestKey())}`,
    "KEYS: A S D F G   ENTER: START",
  ]);
}
