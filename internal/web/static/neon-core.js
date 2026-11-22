// Neon Mile: the race rules, with no DOM, so they can be tested
// (internal/web/jstest). neon.js draws it.
//
// Four cars, one mile of city street at night, and slower traffic in the
// way. Near misses refill your nitro; hitting traffic costs you most of
// your speed.
import { rng } from "./rng.js";

export const MILE = 1609; // meters
export const ROAD_LEFT = 80; // px on the 320px screen
export const ROAD_RIGHT = 240;
export const LANE_W = 40;
export const CAR_W = 14;
export const CAR_LEN = 6; // meters
export const TOP = 40; // m/s
export const NITRO_TOP = 55;
export const ACCEL = 12;
export const NITRO_ACCEL = 20;
export const COAST = 4;
export const BRAKE = 25;
export const NITRO_BURN = 0.35; // share of the tank per second
export const NEAR_MISS = 22; // px between centers that counts as close
export const COUNTDOWN = 3;

export function laneCenter(i) {
  return ROAD_LEFT + LANE_W / 2 + i * LANE_W;
}

export function newRace(seed = Date.now()) {
  const rand = rng(seed);
  const traffic = [];
  for (let d = 70; d < MILE + 200; d += 22 + rand() * 26) {
    traffic.push({ x: laneCenter(Math.floor(rand() * 4)), dist: d, speed: 13 + rand() * 9, passed: false, hit: false, tone: Math.floor(rand() * 3) });
  }
  const rivals = [
    { name: "VIPER", color: "#ff3aa8", base: 40 },
    { name: "GHOST", color: "#a4f04a", base: 41 },
    { name: "BLAZE", color: "#ffb83a", base: 39 },
  ].map((r, i) => ({ ...r, x: laneCenter([0, 2, 3][i]), dist: i === 1 ? 8 : 0, speed: 0, finish: null }));
  return {
    rand,
    t: -COUNTDOWN, // the race starts at t = 0
    you: { x: laneCenter(1), dist: 0, speed: 0, nitro: 0.5, crashT: 0, finish: null },
    rivals,
    traffic,
    over: false,
  };
}

// update advances the race. input: { steer: -1|0|1, gas, brake, nitro }.
export function update(r, input, dt) {
  const ev = [];
  if (r.over) return ev;
  r.t += dt;
  if (r.t < 0) return ev; // countdown
  driveYou(r, input, dt, ev);
  for (const c of r.traffic) c.dist += c.speed * dt;
  for (const rv of r.rivals) driveRival(r, rv, dt);
  collide(r, ev);
  for (const rv of r.rivals) {
    if (rv.finish === null && rv.dist >= MILE) rv.finish = r.t;
  }
  if (r.you.finish === null && r.you.dist >= MILE) {
    r.you.finish = r.t;
    r.over = true;
    ev.push({ type: "finish", place: place(r), time: r.t });
  }
  return ev;
}

function driveYou(r, input, dt, ev) {
  const y = r.you;
  if (y.crashT > 0) {
    y.crashT = Math.max(0, y.crashT - dt);
    y.speed = Math.max(0, y.speed - COAST * dt);
  } else {
    const boosting = input.nitro && y.nitro > 0 && !input.brake;
    const top = boosting ? NITRO_TOP : TOP;
    if (boosting) {
      y.nitro = Math.max(0, y.nitro - NITRO_BURN * dt);
      if (y.nitro === 0) ev.push({ type: "empty" });
    }
    if (input.brake) y.speed -= BRAKE * dt;
    else if (input.gas || boosting) y.speed += (boosting ? NITRO_ACCEL : ACCEL) * dt;
    else y.speed -= COAST * dt;
    if (y.speed > top) y.speed = Math.max(top, y.speed - BRAKE * dt); // ease down after a boost
    y.speed = Math.max(0, y.speed);
    const grip = 70 + 50 * Math.min(1, y.speed / TOP);
    y.x += (input.steer || 0) * grip * dt;
  }
  y.x = Math.max(ROAD_LEFT + CAR_W / 2 + 2, Math.min(ROAD_RIGHT - CAR_W / 2 - 2, y.x));
  y.dist += y.speed * dt;
}

function driveRival(r, rv, dt) {
  if (rv.finish !== null) {
    rv.dist += rv.speed * dt;
    return;
  }
  // Rubber band: keep the race close.
  const gap = rv.dist - r.you.dist;
  let target = rv.base + (gap < -60 ? 4 : 0) - (gap > 80 ? 3 : 0);
  // Look for traffic ahead in this lane; change lanes or slow down.
  const blocker = r.traffic.find((c) => c.dist > rv.dist && c.dist - rv.dist < 24 && Math.abs(c.x - rv.x) < 24);
  if (blocker) {
    const lanes = [0, 1, 2, 3].map(laneCenter).filter((lx) => Math.abs(lx - rv.x) <= LANE_W + 1 && Math.abs(lx - rv.x) > 1);
    const free = lanes.find((lx) => !r.traffic.some((c) => Math.abs(c.dist - rv.dist) < 26 && Math.abs(c.x - lx) < 24));
    if (free !== undefined) rv.targetX = free;
    else target = Math.min(target, blocker.speed + 12); // boxed in: ease off, squeeze past
  }
  if (rv.targetX !== undefined) {
    const dx = rv.targetX - rv.x;
    rv.x += Math.sign(dx) * Math.min(Math.abs(dx), 80 * dt);
    if (Math.abs(dx) < 1) rv.targetX = undefined;
  }
  const accel = rv.speed < target ? ACCEL * 0.9 : -BRAKE;
  rv.speed = Math.max(0, rv.speed + accel * dt);
  if (accel > 0) rv.speed = Math.min(rv.speed, target);
  rv.dist += rv.speed * dt;
}

function collide(r, ev) {
  const y = r.you;
  for (const c of r.traffic) {
    const dd = c.dist - y.dist;
    const dx = Math.abs(c.x - y.x);
    if (!c.hit && y.crashT === 0 && Math.abs(dd) < CAR_LEN && dx < CAR_W) {
      c.hit = true;
      c.passed = true;
      y.crashT = 1;
      y.speed *= 0.3;
      ev.push({ type: "crash" });
      continue;
    }
    if (!c.passed && dd < -CAR_LEN / 2) {
      c.passed = true;
      if (dx < NEAR_MISS && !c.hit && y.speed > 15) {
        y.nitro = Math.min(1, y.nitro + 0.15);
        ev.push({ type: "nearMiss" });
      }
    }
  }
}

// place is your current position, 1 to 4.
export function place(r) {
  const y = r.you;
  let p = 1;
  for (const rv of r.rivals) {
    if (rv.finish !== null && (y.finish === null || rv.finish < y.finish)) p++;
    else if (rv.finish === null && y.finish === null && rv.dist > y.dist) p++;
  }
  return p;
}

export function ordinal(n) {
  return n + (["", "ST", "ND", "RD"][n] || "TH");
}

export function mph(speed) {
  return Math.round(speed * 2.237);
}
