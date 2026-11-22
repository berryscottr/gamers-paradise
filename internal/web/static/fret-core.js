// Fret Night: the song, the chart and the scoring, with no DOM or audio,
// so it can be tested (internal/web/jstest). fret.js plays and draws it.
//
// The song is an original chiptune written by code: A minor, 140 BPM, a
// four-chord loop, and melodies built from short motifs so they repeat
// the way real songs do. The same seed always makes the same song.
import { rng } from "./rng.js";

export const BPM = 140;
export const BEAT = 60 / BPM;
export const STEP = BEAT / 2; // an eighth note
export const LEAD_IN = 4 * BEAT; // one bar of count-in before the first note
export const LANES = 5;
export const PERFECT = 0.045; // seconds either side of the note
export const GOOD = 0.1;
export const SEED = 1995;
export const SONG_TITLE = "Midnight Riff";

const SCALE = [57, 59, 60, 62, 64, 65, 67]; // A natural minor, from A3
const scaleNote = (i) => SCALE[((i % 7) + 7) % 7] + 12 * Math.floor(i / 7);

// The loop: Am, F, C, G as scale degrees of their roots.
const CHORDS = [
  { name: "Am", deg: 0 },
  { name: "F", deg: 5 },
  { name: "C", deg: 2 },
  { name: "G", deg: 6 },
];

export const SECTIONS = [
  { name: "intro", bars: 4, density: 0.4, drums: false },
  { name: "verse", bars: 8, density: 0.7, drums: true },
  { name: "chorus", bars: 8, density: 0.95, drums: true },
  { name: "verse", bars: 8, density: 0.7, drums: true },
  { name: "chorus", bars: 8, density: 0.95, drums: true },
  { name: "outro", bars: 4, density: 0.5, drums: true },
];

// motif makes a two-bar rhythm (which of 16 eighth steps play) and a
// melodic contour (scale steps above the chord root) for a section.
function motif(rand, density) {
  const steps = [];
  let c = 2;
  for (let i = 0; i < 16; i++) {
    const strong = i % 4 === 0;
    const plays = strong || rand() < density * (i % 2 === 0 ? 1 : 0.7);
    if (plays) {
      c += Math.round((rand() - 0.5) * 4);
      c = Math.max(-2, Math.min(6, strong ? Math.round(c / 2) * 2 : c)); // strong steps land on chord tones
      steps.push({ step: i, contour: c });
    }
  }
  return steps;
}

// song returns every note the music plays: lead, bass and drums.
export function song(seed = SEED) {
  const rand = rng(seed);
  const lead = [];
  const bass = [];
  const drums = [];
  let bar = 0;
  for (const sec of SECTIONS) {
    const m = motif(rand, sec.density);
    for (let b = 0; b < sec.bars; b++, bar++) {
      const chord = CHORDS[bar % 4];
      const half = (b % 2) * 8; // which bar of the two-bar motif
      const t0 = LEAD_IN + bar * 4 * BEAT;
      for (const n of m) {
        if (n.step < half || n.step >= half + 8) continue;
        const idx = 5 + chord.deg + n.contour;
        lead.push({ t: t0 + (n.step - half) * STEP, midi: scaleNote(idx), dur: STEP * 0.9, contour: n.contour, step: n.step - half, section: sec.name });
      }
      const root = scaleNote(chord.deg) - 12;
      for (let beat = 0; beat < 4; beat++) {
        bass.push({ t: t0 + beat * BEAT, midi: root, dur: BEAT * 0.45 });
        if (sec.name === "chorus") bass.push({ t: t0 + beat * BEAT + STEP, midi: root + 12, dur: STEP * 0.8 });
      }
      for (let s = 0; s < 8; s++) {
        if (sec.drums || s % 2 === 0) drums.push({ t: t0 + s * STEP, kind: "hat" });
        if (!sec.drums) continue;
        if (s === 0 || s === 4 || (sec.name === "chorus" && s === 7)) drums.push({ t: t0 + s * STEP, kind: "kick" });
        if (s === 2 || s === 6) drums.push({ t: t0 + s * STEP, kind: "snare" });
      }
    }
  }
  const count = [0, 1, 2, 3].map((i) => ({ t: i * BEAT, kind: "click" }));
  const length = LEAD_IN + bar * 4 * BEAT;
  return { lead, bass, drums: count.concat(drums), length, bars: bar };
}

// laneFor maps a melody's height to one of the five lanes, so the chart
// climbs and falls with the tune.
export function laneFor(contour) {
  return Math.max(0, Math.min(LANES - 1, Math.floor((contour + 2) / 2)));
}

// chart turns the melody into notes to hit. Normal keeps notes on the
// beat; expert keeps every note and adds two-lane chords on chorus
// downbeats.
export function chart(s, difficulty = "expert") {
  const notes = [];
  for (const n of s.lead) {
    if (difficulty === "normal" && n.step % 2 !== 0) continue;
    const lane = laneFor(n.contour);
    notes.push({ t: n.t, lane });
    if (difficulty === "expert" && n.section === "chorus" && n.step === 0) {
      notes.push({ t: n.t, lane: lane >= 2 ? lane - 2 : lane + 2 });
    }
  }
  notes.sort((a, b) => a.t - b.t || a.lane - b.lane);
  return notes.map((n, id) => ({ ...n, id, judged: null }));
}

export function newGame(difficulty = "expert", seed = SEED) {
  const s = song(seed);
  return {
    difficulty,
    song: s,
    notes: chart(s, difficulty),
    status: "playing", // playing | failed | done
    score: 0,
    streak: 0,
    bestStreak: 0,
    hits: 0,
    perfects: 0,
    misses: 0,
    meter: 60, // the crowd: 0 and they leave
    next: 0, // first note that might still be unjudged
  };
}

export function multiplier(streak) {
  return Math.min(4, 1 + Math.floor(streak / 10));
}

// press judges a key press in a lane at song time t.
export function press(g, lane, t) {
  if (g.status !== "playing") return { type: "none" };
  let best = null;
  for (let i = g.next; i < g.notes.length && g.notes[i].t <= t + GOOD; i++) {
    const n = g.notes[i];
    if (n.judged || n.lane !== lane) continue;
    if (Math.abs(n.t - t) <= GOOD && (!best || Math.abs(n.t - t) < Math.abs(best.t - t))) best = n;
  }
  if (!best) {
    // a press with nothing to hit breaks the streak
    g.streak = 0;
    g.meter = Math.max(0, g.meter - 2);
    checkFail(g);
    return { type: "stray", lane };
  }
  const perfect = Math.abs(best.t - t) <= PERFECT;
  best.judged = perfect ? "perfect" : "good";
  g.streak++;
  g.bestStreak = Math.max(g.bestStreak, g.streak);
  g.hits++;
  if (perfect) g.perfects++;
  const points = (perfect ? 100 : 60) * multiplier(g.streak);
  g.score += points;
  g.meter = Math.min(100, g.meter + (perfect ? 3 : 2));
  return { type: best.judged, lane, points };
}

// update marks notes that slid past as missed and ends the song.
export function update(g, t) {
  const ev = [];
  if (g.status !== "playing") return ev;
  while (g.next < g.notes.length && (g.notes[g.next].judged || g.notes[g.next].t < t - GOOD)) {
    const n = g.notes[g.next];
    if (!n.judged) {
      n.judged = "miss";
      g.misses++;
      g.streak = 0;
      g.meter = Math.max(0, g.meter - 6);
      ev.push({ type: "miss", lane: n.lane });
    }
    g.next++;
  }
  if (checkFail(g)) ev.push({ type: "failed" });
  else if (t > g.song.length + 1) {
    g.status = "done";
    ev.push({ type: "done", stars: stars(g) });
  }
  return ev;
}

function checkFail(g) {
  if (g.meter <= 0 && g.status === "playing") {
    g.status = "failed";
    return true;
  }
  return false;
}

// stars rates a finished song by accuracy, one to five.
export function stars(g) {
  const acc = g.notes.length ? g.hits / g.notes.length : 0;
  if (acc >= 0.95) return 5;
  if (acc >= 0.85) return 4;
  if (acc >= 0.7) return 3;
  if (acc >= 0.5) return 2;
  return 1;
}

export function midiToHz(m) {
  return 440 * Math.pow(2, (m - 69) / 12);
}
