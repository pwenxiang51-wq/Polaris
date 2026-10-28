(function (root, factory) {
  "use strict";
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  if (root) root.PolarisChartUtils = api;
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  "use strict";

  function sampleAcross(values, maxPoints) {
    if (!Array.isArray(values) || values.length <= 0) return [];
    const limit = Math.max(2, Math.floor(Number(maxPoints) || 240));
    if (values.length <= limit) return values.slice();
    const result = [];
    for (let index = 0; index < limit; index++) {
      const sourceIndex = Math.round(index * (values.length - 1) / (limit - 1));
      result.push(values[sourceIndex]);
    }
    return result;
  }

  return { sampleAcross };
});
