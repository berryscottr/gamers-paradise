// Half-Pipe Hero: the game rules, with no DOM or canvas, so they can be
// tested (internal/web/jstest). skate.js draws it and reads the keyboard.
//
// The rider is a pendulum on a circular half-pipe: energy E sets how high
// they ride; past the coping (E > K) they launch straight up. Pumping at the
// bottom adds energy; the ramp's friction slowly takes it away.

export const G = 600; // gravity, art px/s^2
export const R = 90; // pipe radius, art px
export const K = G / R; // energy needed to reach the coping
export const E_MAX = K + 5.5;
export const PUMP = 1.1;
export const FRICTION = 0.35;
export const START_E = 3.2;
export const RUN_SECONDS = 90;
export const FLIP_TIME = 0.32;
export const SPIN_RATE = 720; // degrees per second
export const GRAB_POINTS = 15; // per tenth of a second held
export const MAX_STREAK = 5;

export const FLIPS = {
  left: { name: "Kickflip", pts: 100 },
  right: { name: "Heelflip", pts: 100 },
  up: { name: "Impossible", pts: 150 },
  down: { name: "Pop Shove-It", pts: 80 },
};

export const GRABS = {
  left: { name: "Indy", pts: 60 },
  right: { name: "Melon", pts: 60 },
  up: { name: "Nosegrab", pts: 70 },
  down: { name: "Tailgrab", pts: 70 },
};

export function newRun(best = 0) {
  return {
    mode: "ready", // ready | ramp | air | bail | over
    theta: 0,
    dir: 1,
    E: START_E,
    pumped: false,
    air: null,
    streak: 0,
    score: 0,
    time: RUN_SECONDS,
    best,
    bailT: 0,
  };
}

export function start(s) {
  Object.assign(s, newRun(s.best), { mode: "ramp" });
  return s;
}

// comboValue scores one air: the sum of its tricks, times how many there
// were (spins count as one trick).
export function comboValue(tricks, spins180, grabTime) {
  const count = tricks.length + (spins180 > 0 ? 1 : 0);
  const base =
    tricks.reduce((sum, t) => sum + t.pts, 0) +
    spins180 * 100 +
    Math.floor(grabTime * 10 + 1e-9) * GRAB_POINTS;
  const names = tricks.map((t) => t.name);
  if (spins180 > 0) names.unshift(String(spins180 * 180));
  return { points: base * count, count, name: names.join(" + ") };
}

// update advances the run by dt seconds. input: { pump, flip, grab, spin }
// where pump is a press this frame, flip a direction pressed this frame (or
// null), grab a direction being held (or null), and spin -1, 0 or 1 (held).
// It returns the events that happened, for sounds and pop-ups.
export function update(s, input, dt) {
  const ev = [];
  if (s.mode === "ready" || s.mode === "over") return ev;

  if (s.mode === "ramp") ramp(s, input, dt, ev);
  else if (s.mode === "air") air(s, input, dt, ev);
  else if (s.mode === "bail") {
    s.bailT -= dt;
    if (s.bailT <= 0) Object.assign(s, { mode: "ramp", theta: 0, dir: 1, E: START_E, pumped: false });
  }

  s.time -= dt;
  if (s.time <= 0 && s.mode !== "air") {
    s.time = 0;
    s.mode = "over";
    s.best = Math.max(s.best, s.score);
    ev.push({ type: "over", score: s.score, best: s.best });
  }
  return ev;
}

function ramp(s, input, dt, ev) {
  if (input.pump && Math.abs(s.theta) < 0.35 && !s.pumped) {
    s.E = Math.min(E_MAX, s.E + PUMP);
    s.pumped = true;
    ev.push({ type: "pump" });
  }
  if (Math.abs(s.theta) > 0.6) s.pumped = false;
  s.E = Math.max(START_E * 0.6, s.E - FRICTION * dt);

  const w2 = 2 * (s.E - K * (1 - Math.cos(s.theta)));
  if (w2 <= 0) {
    turn(s);
    return;
  }
  s.theta += s.dir * Math.sqrt(w2) * dt;
  if (Math.abs(s.theta) >= Math.PI / 2) {
    const side = Math.sign(s.theta);
    if (s.E > K + 0.05) {
      s.mode = "air";
      s.theta = (side * Math.PI) / 2;
      s.air = {
        side,
        y: 0,
        vy: R * Math.sqrt(2 * (s.E - K)),
        tricks: [],
        flipT: 0,
        grab: null,
        grabTime: 0,
        spin: 0,
        spinDir: 0,
      };
      ev.push({ type: "launch" });
    } else {
      turn(s);
    }
  }
}

// turn reverses at the highest point the rider's energy reaches.
function turn(s) {
  const side = Math.sign(s.theta) || 1;
  const c = 1 - s.E / K;
  const top = c <= 0 ? Math.PI / 2 : Math.acos(Math.min(1, c));
  // a little below the top, so friction can't pin the rider there
  s.theta = side * Math.min(top, Math.PI / 2) * 0.97;
  s.dir = -side;
}

function air(s, input, dt, ev) {
  const a = s.air;
  a.y += a.vy * dt;
  a.vy -= G * dt;

  if (input.flip && a.flipT <= 0) {
    const t = FLIPS[input.flip];
    a.tricks.push(t);
    a.flipT = FLIP_TIME;
    ev.push({ type: "trick", name: t.name });
  }
  a.flipT = Math.max(0, a.flipT - dt);

  if (input.grab) {
    if (!a.grab) {
      const g = GRABS[input.grab];
      a.grab = g.name;
      a.tricks.push(g);
      ev.push({ type: "trick", name: g.name });
    }
    a.grabTime += dt;
  } else {
    a.grab = null;
  }

  if (input.spin) {
    a.spinDir = input.spin;
    a.spin += input.spin * SPIN_RATE * dt;
  } else if (a.spinDir !== 0) {
    // let go: finish the current half turn
    const target =
      a.spinDir > 0 ? Math.ceil(a.spin / 180 - 1e-9) * 180 : Math.floor(a.spin / 180 + 1e-9) * 180;
    const next = a.spin + a.spinDir * SPIN_RATE * dt;
    if ((a.spinDir > 0 && next >= target) || (a.spinDir < 0 && next <= target)) {
      a.spin = target;
      a.spinDir = 0;
    } else {
      a.spin = next;
    }
  }

  if (a.y <= 0 && a.vy < 0) land(s, ev);
}

function land(s, ev) {
  const a = s.air;
  s.air = null;
  const clean = a.flipT <= 0 && !a.grab && a.spinDir === 0;
  if (!clean) {
    Object.assign(s, { mode: "bail", bailT: 1.2, streak: 0 });
    ev.push({ type: "bail" });
    return;
  }
  const spins = Math.round(Math.abs(a.spin) / 180);
  const c = comboValue(a.tricks, spins, a.grabTime);
  let points = 0;
  if (c.count > 0) {
    s.streak = Math.min(MAX_STREAK, s.streak + 1);
    points = c.points * s.streak;
    s.score += points;
  }
  Object.assign(s, { mode: "ramp", theta: a.side * (Math.PI / 2 - 0.001), dir: -a.side, pumped: false });
  ev.push({ type: "land", name: c.name, points, streak: s.streak });
}
