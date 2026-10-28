const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const { sampleAcross } = require("../internal/webui/assets/chart-utils.js");

const start = Date.parse("2026-10-27T00:00:00.000Z");
const oneMinute = 60 * 1000;
const history = Array.from({ length: 24 * 60 + 1 }, (_, index) => ({
  at: new Date(start + index * oneMinute).toISOString(),
  value: index,
}));
const sampled = sampleAcross(history, 240);
assert.equal(sampled.length, 240, "24-hour history should be sampled to 240 points");
assert.equal(sampled[0].at, history[0].at, "sampling must preserve the first timestamp");
assert.equal(sampled.at(-1).at, history.at(-1).at, "sampling must preserve the last timestamp");
assert.equal(Date.parse(sampled.at(-1).at) - Date.parse(sampled[0].at), 24 * 60 * oneMinute, "sampled x-axis must span the full 24 hours");
assert.ok(sampled.some((point) => point.value >= 700 && point.value <= 740), "sampling must retain points from the middle of the 24-hour range");

const appSource = fs.readFileSync(path.join(__dirname, "../internal/webui/assets/app.js"), "utf8");
assert.match(appSource, /Array\.isArray\(node\.history\) \? node\.history : \[\]/, "dashboard must use the complete API history array");
assert.doesNotMatch(appSource, /history\.slice\(-80\)/, "dashboard must not discard the older 24-hour history tail");
assert.match(appSource, /sampleAcross\(filtered, 240\)/, "scaled charts must uniformly sample to at most 240 points");
assert.match(appSource, /sampleAcross\(values, 240\)/, "percentage charts must uniformly sample to at most 240 points");

console.log("PASS: all chart series preserve the full 24-hour range and sample to at most 240 points");
