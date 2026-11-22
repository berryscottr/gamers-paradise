// Sound for the whole site, synthesized with Web Audio (no audio files).
// Starts muted; the speaker button turns it on and the choice is remembered
// in this browser only.

const KEY = "gp-sound";
let ctx = null;
let on = false;
let hum = null;

function load() {
  try {
    return localStorage.getItem(KEY) === "on";
  } catch {
    return false;
  }
}

function save(v) {
  try {
    localStorage.setItem(KEY, v ? "on" : "off");
  } catch {
    // private mode or storage blocked: sound just won't be remembered
  }
}

function audio() {
  if (!ctx) {
    const AC = window.AudioContext || window.webkitAudioContext;
    if (!AC) return null;
    ctx = new AC();
  }
  if (ctx.state === "suspended") ctx.resume();
  return ctx;
}

export function isOn() {
  return on;
}

// tone plays one short square-wave note.
export function tone(freq, dur = 0.06, vol = 0.05, type = "square") {
  if (!on) return;
  const a = audio();
  if (!a) return;
  const o = a.createOscillator();
  const g = a.createGain();
  o.type = type;
  o.frequency.value = freq;
  g.gain.setValueAtTime(vol, a.currentTime);
  g.gain.exponentialRampToValueAtTime(0.0001, a.currentTime + dur);
  o.connect(g).connect(a.destination);
  o.start();
  o.stop(a.currentTime + dur + 0.02);
}

export function blip() {
  tone(880, 0.04, 0.03);
}

export function jingle(notes) {
  notes.forEach((f, i) => setTimeout(() => tone(f, 0.09, 0.05), i * 90));
}

// The CRT's high whine and low hum, very quiet.
function startHum() {
  const a = audio();
  if (!a || hum) return;
  const g = a.createGain();
  g.gain.value = 0.012;
  const low = a.createOscillator();
  low.type = "sine";
  low.frequency.value = 60;
  const high = a.createOscillator();
  high.type = "sine";
  high.frequency.value = 15700;
  const hg = a.createGain();
  hg.gain.value = 0.15;
  low.connect(g);
  high.connect(hg).connect(g);
  g.connect(a.destination);
  low.start();
  high.start();
  hum = { low, high };
}

function stopHum() {
  if (!hum) return;
  hum.low.stop();
  hum.high.stop();
  hum = null;
}

function paint(btn) {
  btn.setAttribute("aria-pressed", String(on));
  btn.title = on ? "Sound is on. Click to mute." : "Sound is off. Click to turn it on.";
  const img = btn.querySelector("img");
  if (img) {
    img.src = on ? "/static/art/speaker-on.png" : "/static/art/speaker-off.png";
    img.alt = on ? "Sound on" : "Sound off";
  }
}

// setup wires the speaker button. Browsers only allow audio after a click,
// so a remembered "on" waits for the first click anywhere on the page.
export function setup(btn, { room = false } = {}) {
  const wanted = load();
  paint(btn);
  btn.addEventListener("click", () => {
    on = !on;
    save(on);
    paint(btn);
    if (on) {
      if (room) startHum();
      jingle([523, 659, 784]);
    } else {
      stopHum();
    }
  });
  if (wanted) {
    const wake = (e) => {
      if (on || btn.contains(e.target)) return; // the button's own click decides
      on = true;
      paint(btn);
      if (room) startHum();
    };
    document.addEventListener("pointerdown", wake, { once: true });
    document.addEventListener("keydown", wake, { once: true });
  }
}
