import { test } from "node:test";
import assert from "node:assert/strict";
import {
  song, chart, newGame, press, update, multiplier, stars, laneFor, midiToHz,
  LANES, PERFECT, GOOD, LEAD_IN, BEAT,
} from "../static/fret-core.js";

test("the song is repeatable, about 70 seconds, and starts after the count-in", () => {
  const a = song();
  const b = song();
  assert.deepEqual(a, b);
  assert.ok(a.length > 60 && a.length < 80, `length ${a.length}`);
  assert.ok(a.lead.every((n) => n.t >= LEAD_IN));
  assert.equal(a.drums.filter((d) => d.kind === "click").length, 4);
  assert.notDeepEqual(song(1).lead, a.lead, "another seed, another melody");
});

test("melody notes stay in a singable range", () => {
  for (const n of song().lead) assert.ok(n.midi >= 50 && n.midi <= 90, `midi ${n.midi}`);
});

test("charts use every lane, and expert has more notes than normal", () => {
  const s = song();
  const expert = chart(s, "expert");
  const normal = chart(s, "normal");
  assert.ok(expert.length > normal.length);
  for (let l = 0; l < LANES; l++) assert.ok(expert.some((n) => n.lane === l), `lane ${l} unused`);
  for (let i = 1; i < expert.length; i++) assert.ok(expert[i].t >= expert[i - 1].t, "chart is in time order");
  assert.ok(expert.some((n, i) => i > 0 && n.t === expert[i - 1].t), "expert has chords");
  assert.ok(!normal.some((n, i) => i > 0 && n.t === normal[i - 1].t), "normal has no chords");
});

test("lanes follow the melody's height", () => {
  assert.equal(laneFor(-2), 0);
  assert.equal(laneFor(0), 1);
  assert.equal(laneFor(6), 4);
  assert.equal(laneFor(99), 4);
});

test("perfect and good hits score with the multiplier", () => {
  const g = newGame("expert");
  const n = g.notes[0];
  const r = press(g, n.lane, n.t + PERFECT / 2);
  assert.equal(r.type, "perfect");
  assert.equal(r.points, 100);
  const n2 = g.notes.find((x) => x.t > n.t);
  const r2 = press(g, n2.lane, n2.t - (PERFECT + GOOD) / 2);
  assert.equal(r2.type, "good");
  assert.equal(g.streak, 2);
  assert.equal(multiplier(9), 1);
  assert.equal(multiplier(10), 2);
  assert.equal(multiplier(100), 4);
});

test("a stray press or a missed note breaks the streak and costs crowd", () => {
  const g = newGame("normal");
  const n = g.notes[0];
  press(g, n.lane, n.t);
  assert.equal(g.streak, 1);
  const before = g.meter;
  assert.equal(press(g, (n.lane + 1) % LANES, n.t + 0.5).type, "stray");
  assert.equal(g.streak, 0);
  assert.ok(g.meter < before);
  const ev = update(g, g.notes[1].t + GOOD + 0.01);
  assert.ok(ev.some((e) => e.type === "miss"));
  assert.equal(g.misses >= 1, true);
});

test("ignoring the whole song empties the crowd meter", () => {
  const g = newGame("expert");
  let failed = false;
  for (let t = 0; t < g.song.length + 2 && !failed; t += 0.05) {
    failed = update(g, t).some((e) => e.type === "failed");
  }
  assert.ok(failed);
  assert.equal(g.status, "failed");
  assert.equal(press(g, 0, 10).type, "none", "no presses after failing");
});

test("hitting every note finishes with five stars", () => {
  const g = newGame("expert");
  let done = null;
  let i = 0;
  for (let t = 0; t < g.song.length + 2; t += 0.01) {
    while (i < g.notes.length && g.notes[i].t <= t) press(g, g.notes[i].lane, g.notes[i++].t);
    const ev = update(g, t).find((e) => e.type === "done");
    if (ev) done = ev;
  }
  assert.ok(done, "song finished");
  assert.equal(g.hits, g.notes.length);
  assert.equal(stars(g), 5);
  assert.equal(done.stars, 5);
});

test("A4 is 440 Hz", () => {
  assert.equal(midiToHz(69), 440);
  assert.ok(Math.abs(BEAT - 60 / 140) < 1e-12);
});
