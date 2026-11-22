// phaseFor picks the room's lighting for an hour (0-23). spec is the
// server's "hour:name,..." list (from internal/art Phases), in clock order.
// Hours before the first phase belong to the last one (night wraps past
// midnight).
export function phaseFor(spec, hour) {
  const phases = String(spec || "")
    .split(",")
    .map((p) => p.split(":"))
    .filter((p) => p.length === 2 && p[1])
    .map(([h, name]) => ({ from: Number(h), name }));
  if (phases.length === 0) return "night";
  let pick = phases[phases.length - 1].name;
  for (const p of phases) {
    if (hour >= p.from) pick = p.name;
  }
  return pick;
}
