import { test } from "node:test";
import assert from "node:assert/strict";
import { phaseFor } from "../static/phase.js";

const spec = "2:latenight,5:morning,11:afternoon,17:dusk,21:night";

test("each hour maps to the right phase", () => {
  const want = {
    0: "night", 1: "night", 2: "latenight", 4: "latenight", 5: "morning", 10: "morning",
    11: "afternoon", 16: "afternoon", 17: "dusk", 20: "dusk", 21: "night", 23: "night",
  };
  for (const [h, name] of Object.entries(want)) {
    assert.equal(phaseFor(spec, Number(h)), name, `hour ${h}`);
  }
});

test("a missing spec falls back to night", () => {
  assert.equal(phaseFor("", 12), "night");
  assert.equal(phaseFor(undefined, 12), "night");
});
