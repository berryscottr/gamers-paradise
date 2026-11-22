import { test } from "node:test";
import assert from "node:assert/strict";
import { newBattle, act, damage, COMMANDS, DEW_DROPS, DEW_HEAL, statusWord } from "../static/battle-core.js";

const texts = (log) => log.map((m) => m.text);

function play(seed, policy, maxTurns = 60) {
  const b = newBattle(seed);
  for (let i = 0; i < maxTurns && !b.over; i++) act(b, policy(b));
  return b;
}

const strategy = (b) => {
  if (b.you.hp <= 26 && b.drops > 0) return "drop";
  if (!b.foe.sapped) return "sap";
  if (!b.foe.status) return "toxic";
  return "lash";
};

test("the same seed plays out the same way", () => {
  const a = play(7, strategy);
  const b = play(7, strategy);
  assert.equal(a.turn, b.turn);
  assert.equal(a.over, b.over);
  assert.equal(a.foe.hp, b.foe.hp);
});

test("brute force never wins; status strategy almost always does", () => {
  let brute = 0;
  let smart = 0;
  for (let s = 1; s <= 300; s++) {
    if (play(s, () => "lash").over === "won") brute++;
    if (play(s, strategy).over === "won") smart++;
  }
  assert.equal(brute, 0);
  assert.ok(smart >= 270, `strategy won ${smart}/300`);
});

test("vine lash barely scratches the Lv 42 foe", () => {
  const b = newBattle(1);
  assert.ok(damage(b, b.you, b.foe, 30) <= 5);
  assert.ok(damage(b, b.foe, b.you, 40) >= 20);
});

test("only one major status at a time", () => {
  const b = newBattle(3);
  b.rand = () => 0.5; // moves hit, and the foe doesn't shake it off
  act(b, "toxic");
  assert.equal(b.foe.status, "PSN");
  const log = texts(act(b, "sleep"));
  assert.ok(log.some((t) => t.includes("ALREADY POISONED")), log.join(" | "));
  assert.equal(b.foe.status, "PSN");
});

test("poison gets worse every turn", () => {
  const b = newBattle(4);
  b.rand = () => 0.5;
  b.you.hp = 9999;
  b.you.maxHp = 9999;
  b.foe.status = "PSN";
  const hits = [];
  for (let i = 0; i < 3; i++) {
    const before = b.foe.hp;
    act(b, "lash");
    hits.push(before - b.foe.hp);
  }
  assert.ok(hits[1] > hits[0] && hits[2] > hits[1], `poison ticks ${hits}`);
});

test("a sleeping foe loses its turns, then wakes", () => {
  const b = newBattle(5);
  b.rand = () => 0.5;
  b.foe.status = "SLP";
  b.foe.sleepTurns = 2;
  const hp = b.you.hp;
  const log = texts(act(b, "lash")).concat(texts(act(b, "lash")));
  assert.equal(b.you.hp, hp, "no damage while asleep");
  assert.ok(log.includes("MAGMAUL IS FAST ASLEEP."));
  assert.ok(log.includes("MAGMAUL WOKE UP!"));
  assert.equal(b.foe.status, null);
});

test("paralysis halves speed so you move first", () => {
  const b = newBattle(6);
  b.rand = () => 0.5;
  assert.ok(texts(act(b, "lash"))[0].startsWith("MAGMAUL"), "foe is faster at first");
  b.foe.status = "PAR";
  assert.ok(texts(act(b, "lash"))[0].startsWith("THORNLING"), "paralyzed foe is slower");
});

test("sap siphon drains the foe and heals you every turn", () => {
  const b = newBattle(8);
  b.rand = () => 0.5;
  act(b, "sap");
  assert.ok(b.foe.sapped);
  b.you.hp = 10;
  const foeHp = b.foe.hp;
  b.foe.status = "SLP";
  b.foe.sleepTurns = 5;
  act(b, "lash");
  assert.ok(b.you.hp > 10);
  assert.ok(b.foe.hp <= foeHp - 20);
});

test("dew drops heal and run out", () => {
  const b = newBattle(9);
  b.rand = () => 0.5;
  for (let i = 0; i < DEW_DROPS; i++) {
    b.you.hp = 1;
    b.foe.status = "SLP";
    b.foe.sleepTurns = 9;
    act(b, "drop");
    assert.equal(b.you.hp, 1 + DEW_HEAL);
  }
  const log = act(b, "drop");
  assert.ok(log[0].retry, "no turn is used when the bag is empty");
  assert.equal(b.drops, 0);
});

test("commands are validated and nothing happens after the end", () => {
  const b = newBattle(10);
  assert.throws(() => act(b, "nuke"));
  b.over = "won";
  assert.deepEqual(act(b, "lash"), []);
  assert.equal(COMMANDS.length, 6);
  assert.equal(statusWord("PAR"), "PARALYZED");
});

test("every message can be drawn by the pixel font", () => {
  const allowed = /^[A-Z0-9 '\-.!?:/&+]*$/;
  for (let s = 1; s <= 50; s++) {
    const b = newBattle(s);
    for (let i = 0; i < 40 && !b.over; i++) {
      for (const m of act(b, COMMANDS[i % COMMANDS.length])) assert.match(m.text, allowed);
    }
  }
});
