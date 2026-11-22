// Status Effect: the battle rules, with no DOM, so they can be tested
// (internal/web/jstest). battle.js draws it.
//
// Your Lv 12 Thornling faces a Lv 42 Magmaul. Hitting it barely scratches
// it; putting it to sleep, poisoning it, paralyzing it and siphoning its
// sap is how you win. Original creatures and moves.
import { rng } from "./rng.js";

export const MOVES = {
  lash: { name: "VINE LASH", kind: "damage", power: 30, acc: 0.95 },
  sleep: { name: "SLEEP SPORE", kind: "status", status: "SLP", acc: 0.75 },
  toxic: { name: "TOXIC THORNS", kind: "status", status: "PSN", acc: 0.9 },
  stun: { name: "STUN POLLEN", kind: "status", status: "PAR", acc: 0.85 },
  sap: { name: "SAP SIPHON", kind: "sap", acc: 0.9 },
};
export const COMMANDS = ["lash", "sleep", "toxic", "stun", "sap", "drop"];
export const DEW_DROPS = 3;
export const DEW_HEAL = 25;

const FOE_MOVES = {
  slam: { name: "MAGMA SLAM", power: 40, acc: 0.9 },
  roar: { name: "CINDER ROAR", power: 25, acc: 1 },
};

function creature(name, level, hp, atk, def, spd) {
  return { name, level, hp, maxHp: hp, atk, def, spd, status: null, sleepTurns: 0, toxicN: 0, sapped: false };
}

export function newBattle(seed = Date.now()) {
  return {
    rand: rng(seed),
    you: creature("THORNLING", 12, 60, 10, 20, 14),
    foe: creature("MAGMAUL", 42, 160, 38, 25, 18),
    drops: DEW_DROPS,
    turn: 0,
    over: null, // null | "won" | "lost"
  };
}

// damage uses the classic level-power-attack-defense shape.
export function damage(b, attacker, defender, power) {
  const base = Math.floor((((2 * attacker.level) / 5 + 2) * power * (attacker.atk / defender.def)) / 50) + 2;
  return Math.max(1, Math.floor(base * (0.85 + b.rand() * 0.15)));
}

function speed(c) {
  return c.status === "PAR" ? Math.floor(c.spd / 2) : c.spd;
}

// act plays one whole turn with your command and returns the messages to
// show, in order. Each message may carry HP changes for the UI.
export function act(b, cmd) {
  if (b.over) return [];
  if (!COMMANDS.includes(cmd)) throw new Error(`unknown command ${cmd}`);
  if (cmd === "drop" && b.drops <= 0) return [{ text: "NO DEW DROPS LEFT!", retry: true }];
  b.turn++;
  // Each message carries the HP and status right after it happened, so the
  // UI can animate the bars in step with the text.
  const log = [];
  log.push = (...ms) =>
    Array.prototype.push.apply(
      log,
      ms.map((m) => ({ ...m, youHp: b.you.hp, foeHp: b.foe.hp, foeStatus: b.foe.status, sapped: b.foe.sapped })),
    );
  const youFirst = speed(b.you) >= speed(b.foe);
  const order = youFirst ? ["you", "foe"] : ["foe", "you"];
  for (const who of order) {
    if (b.over) break;
    if (who === "you") yourTurn(b, cmd, log);
    else foeTurn(b, log);
    settle(b, log);
  }
  if (!b.over) endOfTurn(b, log);
  return log;
}

function yourTurn(b, cmd, log) {
  const { you, foe } = b;
  if (cmd === "drop") {
    b.drops--;
    const heal = Math.min(DEW_HEAL, you.maxHp - you.hp);
    you.hp += heal;
    log.push({ text: `THORNLING DRANK A DEW DROP! +${heal} HP` });
    return;
  }
  const m = MOVES[cmd];
  log.push({ text: `THORNLING USED ${m.name}!` });
  if (b.rand() > m.acc) {
    log.push({ text: "BUT IT MISSED!" });
    return;
  }
  if (m.kind === "damage") {
    const d = Math.min(foe.hp, damage(b, you, foe, m.power));
    foe.hp -= d;
    log.push({ text: d <= 5 ? `ONLY ${d} DAMAGE. MAGMAUL IS WAY TOO STRONG.` : `${d} DAMAGE!` });
  } else if (m.kind === "status") {
    if (foe.status) {
      log.push({ text: `MAGMAUL IS ALREADY ${statusWord(foe.status)}!` });
      return;
    }
    foe.status = m.status;
    if (m.status === "SLP") foe.sleepTurns = 2 + Math.floor(b.rand() * 3);
    if (m.status === "PSN") foe.toxicN = 0;
    log.push({ text: `MAGMAUL IS ${statusWord(m.status)}!` });
  } else if (m.kind === "sap") {
    if (foe.sapped) {
      log.push({ text: "MAGMAUL IS ALREADY SAPPED!" });
      return;
    }
    foe.sapped = true;
    log.push({ text: "ROOTS DIG INTO MAGMAUL!" });
  }
}

function foeTurn(b, log) {
  const { you, foe } = b;
  if (foe.status === "SLP") {
    foe.sleepTurns--;
    if (foe.sleepTurns <= 0) {
      foe.status = null;
      log.push({ text: "MAGMAUL WOKE UP!" });
    } else {
      log.push({ text: "MAGMAUL IS FAST ASLEEP." });
    }
    return;
  }
  if (foe.status === "PAR" && b.rand() < 0.25) {
    log.push({ text: "MAGMAUL IS PARALYZED! IT CAN'T MOVE!" });
    return;
  }
  if ((foe.status === "PSN" || foe.status === "PAR") && b.rand() < 0.3) {
    log.push({ text: `MAGMAUL SHOOK IT OFF! NO LONGER ${statusWord(foe.status)}.` });
    foe.status = null;
    foe.toxicN = 0;
    return;
  }
  const m = b.rand() < 0.7 ? FOE_MOVES.slam : FOE_MOVES.roar;
  log.push({ text: `MAGMAUL USED ${m.name}!` });
  if (b.rand() > m.acc) {
    log.push({ text: "IT MISSED!" });
    return;
  }
  const d = Math.min(you.hp, damage(b, foe, you, m.power));
  you.hp -= d;
  log.push({ text: `THORNLING TOOK ${d} DAMAGE!` });
}

function endOfTurn(b, log) {
  const { you, foe } = b;
  if (foe.status === "PSN") {
    foe.toxicN++;
    const d = Math.min(foe.hp, Math.floor((foe.maxHp * foe.toxicN) / 16));
    foe.hp -= d;
    log.push({ text: `MAGMAUL IS HURT BY POISON! -${d}` });
    settle(b, log);
  }
  if (!b.over && foe.sapped) {
    const d = Math.min(foe.hp, Math.floor(foe.maxHp / 8));
    foe.hp -= d;
    const heal = Math.min(d, you.maxHp - you.hp);
    you.hp += heal;
    log.push({ text: `SAP SIPHON DRAINS ${d}! THORNLING +${heal} HP` });
    settle(b, log);
  }
}

function settle(b, log) {
  if (b.over) return;
  if (b.foe.hp <= 0) {
    b.over = "won";
    log.push({ text: "MAGMAUL FAINTED!" }, { text: "STATUS EFFECTS ARE OP. YOU WIN!" });
  } else if (b.you.hp <= 0) {
    b.over = "lost";
    log.push({ text: "THORNLING FAINTED..." }, { text: "TRY A LITTLE MORE STRATEGY." });
  }
}

export function statusWord(s) {
  return { SLP: "ASLEEP", PSN: "POISONED", PAR: "PARALYZED" }[s] || "";
}
