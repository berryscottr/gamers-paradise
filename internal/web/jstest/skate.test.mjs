import { test } from "node:test";
import assert from "node:assert/strict";
import {
  newRun, start, update, comboValue, FLIPS, GRABS, K, E_MAX, RUN_SECONDS, MAX_STREAK,
} from "../static/skate-core.js";

const idle = { pump: false, flip: null, grab: null, spin: 0 };
const DT = 1 / 60;

// ride steps the run until pred(state) or a time limit.
function ride(s, input, pred, seconds = 10) {
  const events = [];
  for (let t = 0; t < seconds; t += DT) {
    events.push(...update(s, typeof input === "function" ? input(s) : input, DT));
    if (pred(s)) break;
  }
  return events;
}

// launch rides at full speed (as after several good pumps) until airborne.
function launch(s) {
  s.E = E_MAX;
  return ride(s, (st) => ({ ...idle, pump: Math.abs(st.theta) < 0.3 }), (st) => st.mode === "air", 30);
}

test("a new run waits for Enter", () => {
  const s = newRun(500);
  assert.equal(s.mode, "ready");
  assert.deepEqual(update(s, idle, 1), []);
  start(s);
  assert.equal(s.mode, "ramp");
  assert.equal(s.best, 500);
  assert.equal(s.time, RUN_SECONDS);
});

test("without pumping the rider never leaves the pipe", () => {
  const s = start(newRun());
  ride(s, idle, (st) => st.mode !== "ramp", 20);
  assert.equal(s.mode, "ramp");
  assert.ok(Math.abs(s.theta) < Math.PI / 2);
});

test("pumping at the bottom builds speed until launch", () => {
  const s = start(newRun());
  const ev = ride(s, (st) => ({ ...idle, pump: Math.abs(st.theta) < 0.3 }), (st) => st.mode === "air", 30);
  assert.equal(s.mode, "air");
  assert.ok(ev.some((e) => e.type === "pump"));
  assert.ok(ev.some((e) => e.type === "launch"));
  assert.ok(s.E > K && s.E <= E_MAX);
});

test("a pump only counts once per pass", () => {
  const s = start(newRun());
  const before = s.E;
  update(s, { ...idle, pump: true }, DT);
  update(s, { ...idle, pump: true }, DT);
  assert.ok(s.E - before < 1.2, "second pump in the same pass is ignored");
});

test("a clean kickflip lands and scores", () => {
  const s = start(newRun());
  launch(s);
  update(s, { ...idle, flip: "left" }, DT);
  const ev = ride(s, idle, (st) => st.mode !== "air");
  const land = ev.find((e) => e.type === "land");
  assert.ok(land, "landed");
  assert.equal(land.name, FLIPS.left.name);
  assert.equal(land.points, FLIPS.left.pts);
  assert.equal(s.streak, 1);
  assert.equal(s.score, FLIPS.left.pts);
});

test("holding a grab into the landing bails and resets the streak", () => {
  const s = start(newRun());
  s.streak = 3;
  launch(s);
  const ev = ride(s, { ...idle, grab: "up" }, (st) => st.mode !== "air");
  assert.ok(ev.some((e) => e.type === "bail"));
  assert.equal(s.mode, "bail");
  assert.equal(s.streak, 0);
  ride(s, idle, (st) => st.mode === "ramp", 3);
  assert.equal(s.mode, "ramp");
});

test("a half-finished spin finishes itself after you let go", () => {
  const s = start(newRun());
  launch(s);
  ride(s, { ...idle, spin: 1 }, (st) => st.air.spin >= 100);
  const ev = ride(s, idle, (st) => st.mode !== "air");
  const land = ev.find((e) => e.type === "land");
  assert.ok(land, "landed clean after the spin snapped to 180");
  assert.equal(land.name, "180");
});

test("combo value multiplies by the number of tricks", () => {
  const c = comboValue([FLIPS.left, GRABS.right], 2, 0.5);
  // (100 + 60 + 2*100 + 5*15) * 3 tricks
  assert.equal(c.points, (100 + 60 + 200 + 75) * 3);
  assert.equal(c.count, 3);
  assert.equal(c.name, "360 + Kickflip + Melon");
  assert.deepEqual(comboValue([], 0, 0), { points: 0, count: 0, name: "" });
});

test("the streak caps", () => {
  const s = start(newRun());
  for (let i = 0; i < MAX_STREAK + 2; i++) {
    launch(s);
    update(s, { ...idle, flip: "up" }, DT);
    ride(s, idle, (st) => st.mode !== "air");
  }
  assert.equal(s.streak, MAX_STREAK);
});

test("time runs out after the current air lands", () => {
  const s = start(newRun(10));
  s.score = 42;
  s.time = 0.01;
  const ev = ride(s, idle, (st) => st.mode === "over", 1);
  assert.equal(s.mode, "over");
  assert.equal(s.best, 42);
  assert.ok(ev.some((e) => e.type === "over"));
});
