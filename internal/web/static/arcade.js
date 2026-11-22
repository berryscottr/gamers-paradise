// Shared bits for the arcade games: a 320x200 pixel canvas shown at 2x,
// the site's pixel font, keyboard state, and best scores kept in this
// browser only.
import { GLYPHS } from "/static/art/font.js";

export const W = 320;
export const H = 200;

// screen wraps a 640x400 canvas as a 320x200 pixel surface.
export function screen(canvas) {
  const ctx = canvas.getContext("2d");
  const s = {
    ctx,
    begin() {
      ctx.setTransform(2, 0, 0, 2, 0, 0);
      ctx.imageSmoothingEnabled = false;
    },
    rect(x, y, w, h, c) {
      ctx.fillStyle = c;
      ctx.fillRect(Math.round(x), Math.round(y), w, h);
    },
    text(str, x, y, c, scale = 1, align = "left") {
      const w = (String(str).length * 6 - 1) * scale;
      let cx = align === "center" ? Math.round(x - w / 2) : align === "right" ? x - w : x;
      ctx.fillStyle = c;
      for (const ch of String(str).toUpperCase()) {
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
    },
    shadowText(str, x, y, c, scale = 1, align = "left") {
      s.text(str, x + scale, y + scale, "#1a0a2a", scale, align);
      s.text(str, x, y, c, scale, align);
    },
    // sprite draws rows of characters; colors maps characters to colors.
    sprite(rows, x, y, colors, flip = false) {
      rows.forEach((row, j) => {
        for (let i = 0; i < row.length; i++) {
          const c = colors[row[i]];
          if (c) s.rect(flip ? x + row.length - 1 - i : x + i, y + j, 1, 1, c);
        }
      });
    },
    // panel is a title card with lines of text, for start and end screens.
    panel(title, lines) {
      s.rect(40, 50, 240, 100, "rgba(10,8,30,0.88)");
      s.rect(40, 50, 240, 2, "#8a7ae8");
      s.rect(40, 148, 240, 2, "#8a7ae8");
      s.shadowText(title, W / 2, 60, "#fff27a", 2, "center");
      lines.forEach((l, i) => s.text(l, W / 2, 86 + i * 12, "#ffffff", 1, "center"));
    },
  };
  return s;
}

// keyboard tracks held keys and keys pressed since the last clear(). Only
// keys in `keys` are captured (so they don't scroll the page), and only
// while the canvas or the page body has focus.
export function keyboard(canvas, keys) {
  const held = new Set();
  const pressed = new Set();
  const name = (e) => (e.key.length === 1 ? e.key.toLowerCase() : e.key);
  window.addEventListener("keydown", (e) => {
    const k = name(e);
    if (!keys.includes(k)) return;
    if (document.activeElement !== canvas && document.activeElement !== document.body) return;
    e.preventDefault();
    if (!held.has(k)) pressed.add(k);
    held.add(k);
  });
  window.addEventListener("keyup", (e) => held.delete(name(e)));
  window.addEventListener("blur", () => held.clear());
  canvas.addEventListener("pointerdown", () => canvas.focus());
  canvas.focus();
  return {
    held,
    pressed,
    clear() {
      pressed.clear();
    },
  };
}

export function loadBest(key) {
  try {
    return Number(localStorage.getItem(key)) || 0;
  } catch {
    return 0;
  }
}

export function saveBest(key, v) {
  try {
    localStorage.setItem(key, String(v));
  } catch {
    // storage blocked: the best score lasts until the tab closes
  }
}

// loop calls step(dt) every frame with dt capped at 50ms.
export function loop(step) {
  let last = performance.now();
  const frame = (now) => {
    const dt = Math.min(0.05, (now - last) / 1000);
    last = now;
    step(dt, now);
    requestAnimationFrame(frame);
  };
  requestAnimationFrame(frame);
}
