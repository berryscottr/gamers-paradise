// Status Effect: drawing the battle and reading the menu. Rules:
// battle-core.js.
import { screen, keyboard, loop, loadBest, saveBest, W } from "/static/arcade.js";
import { newBattle, act, MOVES, COMMANDS } from "/static/battle-core.js";
import { tone, jingle } from "/static/sound.js";

const canvas = document.getElementById("battle");
const s = screen(canvas);
const keys = keyboard(canvas, ["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", "Enter", "z", " "]);
const BEST_KEY = "gp-status-best"; // fewest turns to win

let b = null;
let phase = "title"; // title | choose | msg | end
let cursor = 0;
let queue = [];
let msg = null;
let msgT = 0;
const shown = { you: 0, foe: 0, foeStatus: null, sapped: false };
let shake = { who: null, t: 0 };

function start() {
  b = newBattle(Date.now() >>> 0);
  shown.you = b.you.hp;
  shown.foe = b.foe.hp;
  shown.foeStatus = null;
  shown.sapped = false;
  queue = [{ text: "A WILD MAGMAUL BLOCKS THE PATH!", youHp: b.you.hp, foeHp: b.foe.hp, foeStatus: null, sapped: false }];
  next();
}

function next() {
  msg = queue.shift() || null;
  msgT = 0;
  if (!msg) {
    if (b.over) {
      phase = "end";
      if (b.over === "won") {
        const best = loadBest(BEST_KEY);
        if (!best || b.turn < best) saveBest(BEST_KEY, b.turn);
        jingle([523, 659, 784, 1047]);
      } else {
        jingle([392, 330, 262]);
      }
    } else {
      phase = "choose";
    }
    return;
  }
  phase = "msg";
  if (msg.foeHp < shown.foe) hit("foe");
  if (msg.youHp < shown.you) hit("you");
  shown.foeStatus = msg.foeStatus;
  shown.sapped = msg.sapped;
}

function hit(who) {
  shake = { who, t: 0.3 };
  tone(who === "foe" ? 330 : 165, 0.08, 0.05, "square");
}

loop((dt) => {
  const p = keys.pressed;
  const ok = p.has("Enter") || p.has("z") || p.has(" ");
  if (phase === "title" || phase === "end") {
    if (ok) start();
  } else if (phase === "choose") {
    if (p.has("ArrowUp")) cursor = (cursor + 4) % 6;
    if (p.has("ArrowDown")) cursor = (cursor + 2) % 6;
    if (p.has("ArrowLeft") || p.has("ArrowRight")) cursor ^= 1;
    if (p.has("ArrowUp") || p.has("ArrowDown") || p.has("ArrowLeft") || p.has("ArrowRight")) tone(880, 0.03, 0.03);
    if (ok) {
      const log = act(b, COMMANDS[cursor]);
      if (log.length && log[0].retry) {
        queue = [{ ...log[0], youHp: b.you.hp, foeHp: b.foe.hp, foeStatus: b.foe.status, sapped: b.foe.sapped }];
      } else {
        queue = log;
      }
      next();
    }
  } else if (phase === "msg") {
    msgT += dt;
    if ((ok && msgT > 0.15) || msgT > 2.2) next();
  }
  // HP bars ease toward the current message's numbers
  if (msg) {
    shown.you += Math.sign(msg.youHp - shown.you) * Math.min(Math.abs(msg.youHp - shown.you), 60 * dt);
    shown.foe += Math.sign(msg.foeHp - shown.foe) * Math.min(Math.abs(msg.foeHp - shown.foe), 120 * dt);
  }
  shake.t = Math.max(0, shake.t - dt);
  keys.clear();
  draw();
});

// ---------- art: original creatures ----------
const THORNLING = [
  ".......G..G.......",
  "......GlG.GlG.....",
  ".....GlllGlllG....",
  "......GgggggG.....",
  "....GgggggggggG...",
  "...GgggggggggggG..",
  "..GggwkgggggwkggG.",
  "..GggkkgggggkkggG.",
  "..GgpgggggggggpgG.",
  "..GggggggkkgggggG.",
  "...GgggggggggggG..",
  "..tGGgggggggggGGt.",
  "..t.GlllllllllG.t.",
  "....GlllllllllG...",
  "...GggG.....GggG..",
  "...GGG.......GGG..",
];
const THORN_COLORS = { g: "#4f9a2c", G: "#2c5a16", l: "#a4d46c", k: "#1a1a1a", w: "#ffffff", t: "#e8c84a", p: "#ee8cb4" };

const MAGMAUL = [
  "........RRRR........RRRR........",
  ".......RrrrrR......RrrrrR.......",
  "......RrooorrRRRRRRrrooorR......",
  ".....RrroyyorrrrrrrroyyorrR.....",
  "....RrrrooorrrrrrrrrrooorrrR....",
  "...RrrrrrrrrrrrrrrrrrrrrrrrrR...",
  "...RrrrwwkrrrrrrrrrrrrkwwrrrR...",
  "...RrrrwkkrrrrrrrrrrrrkkwrrrR...",
  "...RrrrrrrrrrooooooorrrrrrrrR...",
  "..RrrrrrrrrookkkkkkoorrrrrrrrR..",
  "..RrrrrrrrrokywywywkorrrrrrrrR..",
  "..RrrrrrrrrrookkkkoorrrrrrrrrR..",
  ".RRrrrorrrrrrrooooorrrrrrorrrRR.",
  "RrrrroyorrrrrrrrrrrrrrrroyorrrrR",
  "RrrrrrorrrrrRRRRRRRRrrrrrrorrrrR",
  ".RrrrrrrrrrR........RrrrrrrrrrR.",
  "..RrrooorrR..........RrooorrrR..",
  "..RRRRRRRR............RRRRRRRR..",
];
const MAG_COLORS = { r: "#7a2a1a", R: "#3a120a", o: "#e8803a", y: "#ffd84a", k: "#1a1a1a", w: "#ffffff" };

function big(rows, x, y, colors, k = 2) {
  rows.forEach((row, j) => {
    for (let i = 0; i < row.length; i++) {
      const c = colors[row[i]];
      if (c) s.rect(x + i * k, y + j * k, k, k, c);
    }
  });
}

// ---------- drawing ----------
function field() {
  for (let y = 0; y < 136; y += 4) s.rect(0, y, W, 4, y < 70 ? "#cfe8f8" : "#e8f0c8");
  s.ctx.fillStyle = "#b8d890";
  s.ctx.beginPath();
  s.ctx.ellipse(232, 62, 52, 10, 0, 0, Math.PI * 2);
  s.ctx.fill();
  s.ctx.beginPath();
  s.ctx.ellipse(70, 122, 56, 11, 0, 0, Math.PI * 2);
  s.ctx.fill();
}

function infoBox(x, y, name, level, hp, maxHp, status, showNumbers) {
  s.rect(x, y, 116, showNumbers ? 30 : 22, "#f8f8e8");
  s.rect(x, y + (showNumbers ? 29 : 21), 116, 1, "#2a2a2a");
  s.rect(x + 115, y, 1, showNumbers ? 30 : 22, "#2a2a2a");
  s.text(name, x + 4, y + 3, "#1a1a1a");
  s.text(`LV${level}`, x + 112, y + 3, "#1a1a1a", 1, "right");
  s.text("HP", x + 4, y + 13, "#d84a1a");
  s.rect(x + 18, y + 13, 92, 6, "#2a2a2a");
  const f = Math.max(0, hp / maxHp);
  s.rect(x + 19, y + 14, Math.round(90 * f), 4, f > 0.5 ? "#4ac84a" : f > 0.2 ? "#e8c84a" : "#e84a4a");
  if (status) {
    s.rect(x + 74, y + 2, 20, 9, { SLP: "#8a8a9a", PSN: "#a04ac8", PAR: "#d8b020" }[status]);
    s.text(status, x + 75, y + 3, "#ffffff");
  }
  if (showNumbers) s.text(`${Math.ceil(hp)}/${maxHp}`, x + 110, y + 21, "#1a1a1a", 1, "right");
}

function effects(t) {
  const fx = 236;
  if (shown.foeStatus === "SLP") s.text("Z", fx + 30 + Math.round(Math.sin(t * 3) * 2), 10 - Math.round((t * 6) % 6), "#5a6ac8");
  if (shown.foeStatus === "PSN") {
    for (let i = 0; i < 3; i++) s.rect(fx - 20 + i * 18, 52 - Math.round(((t * 20 + i * 7) % 24)), 3, 3, "#a04ac8");
  }
  if (shown.foeStatus === "PAR" && Math.floor(t * 6) % 2 === 0) {
    s.rect(fx - 30, 30, 4, 1, "#ffd84a");
    s.rect(fx + 34, 24, 4, 1, "#ffd84a");
  }
  if (shown.sapped) {
    for (let i = 0; i < 5; i++) s.rect(fx - 26 + i * 12, 58, 2, 6, "#2c5a16");
  }
}

function textBox() {
  s.rect(0, 136, W, 64, "#f8f8e8");
  s.rect(3, 139, W - 6, 58, "#2a2a4a");
  s.rect(5, 141, W - 10, 54, "#f8f8e8");
}

function menu() {
  textBox();
  s.text("WHAT WILL", 12, 150, "#1a1a1a");
  s.text("THORNLING DO?", 12, 162, "#1a1a1a");
  s.rect(110, 141, 1, 54, "#2a2a4a");
  COMMANDS.forEach((c, i) => {
    const x = 120 + (i % 2) * 100;
    const y = 148 + Math.floor(i / 2) * 15;
    const label = c === "drop" ? `DEW DROP X${b.drops}` : MOVES[c].name;
    s.text(label, x, y, i === cursor ? "#1a1a1a" : "#5a5a6a");
    if (i === cursor) s.text(">", x - 8, y, "#d84a1a");
  });
}

function draw() {
  s.begin();
  const t = performance.now() / 1000;
  field();
  if (!b) {
    big(MAGMAUL, 200, 20, MAG_COLORS);
    big(THORNLING, 52, 86, THORN_COLORS);
    s.panel("STATUS EFFECT", ["LV 12 VS LV 42", "NO GRINDING. JUST STRATEGY.", `FASTEST WIN: ${loadBest(BEST_KEY) ? loadBest(BEST_KEY) + " TURNS" : "NONE YET"}`, "ENTER: BATTLE"]);
    return;
  }
  const sx = (who) => (shake.who === who && shake.t > 0 ? Math.round(Math.sin(t * 60) * 2) : 0);
  if (!(b.over === "won" && phase === "end")) big(MAGMAUL, 200 + sx("foe"), 20, MAG_COLORS);
  if (!(b.over === "lost" && phase === "end")) big(THORNLING, 52 + sx("you"), 86, THORN_COLORS);
  effects(t);
  infoBox(14, 12, b.foe.name, b.foe.level, shown.foe, b.foe.maxHp, shown.foeStatus, false);
  infoBox(190, 96, b.you.name, b.you.level, shown.you, b.you.maxHp, null, true);
  if (phase === "choose") {
    menu();
  } else {
    textBox();
    const text = msg ? msg.text : "";
    wrap(text, 34).forEach((line, i) => s.text(line, 14, 150 + i * 12, "#1a1a1a"));
    if (phase === "msg" && Math.floor(t * 3) % 2 === 0) s.text(">", W - 20, 184, "#d84a1a");
    if (phase === "end") s.text(b.over === "won" ? `WON IN ${b.turn} TURNS. ENTER: AGAIN` : "ENTER: TRY AGAIN", 14, 178, "#5a5a6a");
  }
}

function wrap(text, width) {
  const words = text.split(" ");
  const lines = [""];
  for (const w of words) {
    const cur = lines[lines.length - 1];
    if ((cur + " " + w).trim().length > width) lines.push(w);
    else lines[lines.length - 1] = (cur + " " + w).trim();
  }
  return lines;
}

