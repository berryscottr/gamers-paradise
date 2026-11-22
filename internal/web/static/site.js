// Every page: the sound toggle and link blips. The bedroom: lighting from
// the visitor's clock, the status bar, and the laptop's loading screen.
import { setup, blip } from "/static/sound.js";
import { phaseFor } from "/static/phase.js";

const room = document.getElementById("room");
const btn = document.getElementById("sound");
if (btn) setup(btn, { room: Boolean(room) });

for (const a of document.querySelectorAll("#nav a, .hs, .cart a")) {
  a.addEventListener("mouseenter", blip);
}

if (room) {
  const img = document.getElementById("room-img");
  const status = document.getElementById("status");
  const idle = status.textContent;

  const apply = () => {
    const phase = phaseFor(room.dataset.phases, new Date().getHours());
    const src = `/static/art/room-${phase}.png`;
    if (!img.src.endsWith(src)) img.src = src;
    room.dataset.phase = phase;
  };
  apply();
  setInterval(apply, 60 * 1000); // the sky keeps up if they stay a while

  for (const hs of room.querySelectorAll(".hs")) {
    const show = () => {
      status.textContent = hs.dataset.label;
    };
    const hide = () => {
      status.textContent = idle;
    };
    hs.addEventListener("mouseenter", show);
    hs.addEventListener("focus", show);
    hs.addEventListener("mouseleave", hide);
    hs.addEventListener("blur", hide);
  }

  // The laptop shows this room. Clicking it "boots a Flash game" and zooms
  // back into the bedroom: the loop.
  const laptop = room.querySelector(".hs-laptop");
  const loading = document.getElementById("loading");
  if (laptop && loading) {
    laptop.addEventListener("click", (e) => {
      e.preventDefault();
      room.classList.add("zoom");
      setTimeout(() => {
        loading.hidden = false;
      }, 700);
      setTimeout(() => {
        window.location.assign("/");
      }, 2000);
    });
  }
}
