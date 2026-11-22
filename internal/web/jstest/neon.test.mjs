import { test } from "node:test";
import assert from "node:assert/strict";
import {
  newRace, update, place, ordinal, mph, laneCenter,
  MILE, TOP, NITRO_TOP, COUNTDOWN, ROAD_LEFT, ROAD_RIGHT,
} from "../static/neon-core.js";

const DT = 1 / 60;
const idle = { steer: 0, gas: false, brake: false, nitro: false };
const gas = { ...idle, gas: true };

function clearTraffic(r) {
  r.traffic = [];
  return r;
}

function run(r, input, pred, seconds) {
  const ev = [];
  for (let t = 0; t < seconds && !pred(r); t += DT) ev.push(...update(r, typeof input === "function" ? input(r) : input, DT));
  return ev;
}

test("nobody moves during the countdown", () => {
  const r = newRace(1);
  run(r, gas, () => false, COUNTDOWN - 0.1);
  assert.equal(r.you.dist, 0);
  assert.ok(r.rivals.every((x) => x.speed === 0));
});

test("gas reaches top speed; nitro goes past it and burns the tank", () => {
  const r = clearTraffic(newRace(2));
  run(r, gas, () => false, COUNTDOWN + 6);
  assert.ok(Math.abs(r.you.speed - TOP) < 0.01, `speed ${r.you.speed}`);
  const tank = r.you.nitro;
  run(r, { ...gas, nitro: true }, () => false, 1);
  assert.ok(r.you.speed > TOP && r.you.speed <= NITRO_TOP);
  assert.ok(r.you.nitro < tank);
  run(r, gas, () => false, 3);
  assert.ok(Math.abs(r.you.speed - TOP) < 0.01, "eases back down after the boost");
});

test("brakes slow you and steering stays on the road", () => {
  const r = clearTraffic(newRace(3));
  run(r, gas, () => false, COUNTDOWN + 4);
  const s = r.you.speed;
  run(r, { ...idle, brake: true }, () => false, 0.5);
  assert.ok(r.you.speed < s - 10);
  run(r, { ...gas, steer: -1 }, () => false, 5);
  assert.ok(r.you.x > ROAD_LEFT);
  run(r, { ...gas, steer: 1 }, () => false, 5);
  assert.ok(r.you.x < ROAD_RIGHT);
});

test("hitting traffic costs most of your speed", () => {
  const r = clearTraffic(newRace(4));
  run(r, gas, () => false, COUNTDOWN + 4);
  r.traffic = [{ x: r.you.x, dist: r.you.dist + 8, speed: 10, passed: false, hit: false, tone: 0 }];
  const before = r.you.speed;
  const ev = run(r, gas, (x) => x.you.crashT > 0, 2);
  assert.ok(ev.some((e) => e.type === "crash"));
  assert.ok(r.you.speed < before * 0.5);
});

test("a near miss refills nitro", () => {
  const r = clearTraffic(newRace(5));
  run(r, gas, () => false, COUNTDOWN + 4);
  r.you.nitro = 0.2;
  r.traffic = [{ x: r.you.x + 17, dist: r.you.dist + 10, speed: 10, passed: false, hit: false, tone: 0 }];
  const ev = run(r, gas, (x) => x.traffic[0].passed, 3);
  assert.ok(ev.some((e) => e.type === "nearMiss"));
  assert.ok(r.you.nitro > 0.3);
});

test("a clean, quick driver wins; a driver who ignores traffic doesn't", () => {
  const dodge = (r) => {
    const y = r.you;
    let steer = 0;
    if (r.traffic.some((c) => c.dist > y.dist && c.dist - y.dist < 30 && Math.abs(c.x - y.x) < 20)) {
      const lanes = [0, 1, 2, 3].map(laneCenter).sort((a, b) => Math.abs(a - y.x) - Math.abs(b - y.x));
      const free = lanes.find((lx) => !r.traffic.some((c) => c.dist > y.dist - 6 && c.dist - y.dist < 35 && Math.abs(c.x - lx) < 20));
      if (free !== undefined) steer = Math.sign(free - y.x);
    }
    return { steer, gas: true, brake: false, nitro: y.nitro > 0.3 };
  };
  for (const seed of [11, 12, 13]) {
    const good = newRace(seed);
    const ev = run(good, dodge, (r) => r.over, 120);
    const fin = ev.find((e) => e.type === "finish");
    assert.ok(fin, "finished");
    assert.equal(fin.place, 1, `seed ${seed}`);
    const bad = newRace(seed);
    run(bad, gas, (r) => r.over, 120);
    assert.ok(place(bad) > 1, `seed ${seed}: plowing through traffic shouldn't win`);
  }
});

test("the race ends at one mile", () => {
  const r = clearTraffic(newRace(6));
  run(r, { ...gas, nitro: true }, (x) => x.over, 120);
  assert.ok(r.over);
  assert.ok(r.you.dist >= MILE);
  assert.deepEqual(update(r, gas, DT), []);
});

test("labels", () => {
  assert.deepEqual([1, 2, 3, 4].map(ordinal), ["1ST", "2ND", "3RD", "4TH"]);
  assert.equal(mph(TOP), 89);
});
