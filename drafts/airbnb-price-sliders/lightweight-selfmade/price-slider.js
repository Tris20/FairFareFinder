const BIN_COUNT = 50;

let rawValues = [];
let logValues = [];
let globalMinLin = 0;
let globalMaxLin = 0;
let globalMinLog = 0;
let globalMaxLog = 0;

fetch("dummy.csv")
  .then((response) => response.text())
  .then((csvText) => {
    const lines = csvText.trim().split("\n");
    // Remove the header
    lines.shift();

    // Filter out values < 10
    rawValues = lines
      .map((line) => parseFloat(line.trim()))
      .filter((v) => !isNaN(v) && v >= 1);

    // 1) Linear min/max
    globalMinLin = Math.min(...rawValues);
    globalMaxLin = Math.max(...rawValues);

    // 2) Log transform each value: log10(v + 1)
    logValues = rawValues.map((v) => Math.log10(v + 1));

    // 3) Determine log min/max
    globalMinLog = Math.min(...logValues);
    globalMaxLog = Math.max(...logValues);

    // Setup the slider in log space
    const slider = document.getElementById("value-slider");
    slider.min = globalMinLog; // e.g. log10( min + 1 )
    slider.max = globalMaxLog; // e.g. log10( max + 1 )
    slider.value = globalMaxLog; // start at the highest log
    // We already set step="0.01" above in HTML

    // Initial draw: show full range (which corresponds to slider at max log)
    const initialFilterVal = Math.pow(10, globalMaxLog) - 1;
    drawHistogram(
      rawValues,
      logValues,
      initialFilterVal,
      globalMinLog,
      globalMaxLog,
      BIN_COUNT,
    );
  })
  .catch((err) => console.error("Error fetching CSV:", err));

/**
 * Draw a log-scale histogram with an overlay for the current slider filter.
 *
 * @param {number[]} rawVals  - Original data (linear).
 * @param {number[]} logVals  - Log-transformed data (log10(v+1)).
 * @param {number}   filterVal- The current slider cutoff in linear space.
 * @param {number}   minLog   - The min of the log-transformed data.
 * @param {number}   maxLog   - The max of the log-transformed data.
 * @param {number}   binCount - How many bins to create in log space.
 */
function drawHistogram(rawVals, logVals, filterVal, minLog, maxLog, binCount) {
  const chart = document.getElementById("chart");
  chart.innerHTML = "";

  if (rawVals.length === 0) {
    chart.textContent = "No data";
    return;
  }

  const logRange = maxLog - minLog;
  const binSize = logRange / binCount;

  // Each bin: { total, included }
  let bins = Array.from({ length: binCount }, () => ({
    total: 0,
    included: 0,
  }));

  // Populate bins
  for (let i = 0; i < logVals.length; i++) {
    const lv = logVals[i]; // log10 of (value + 1)
    const rv = rawVals[i]; // raw linear value
    let binIndex = Math.floor((lv - minLog) / binSize);
    // Clamp to [0, binCount - 1]
    if (binIndex < 0) binIndex = 0;
    if (binIndex >= binCount) binIndex = binCount - 1;

    bins[binIndex].total += 1;

    // If the raw value is <= filterVal, it’s included
    if (rv <= filterVal) {
      bins[binIndex].included += 1;
    }
  }

  // Max count in any bin (for scaling bar heights)
  const maxCount = Math.max(...bins.map((b) => b.total));

  // Create bars
  bins.forEach((bin, i) => {
    const barWrapper = document.createElement("div");
    barWrapper.className = "bar";

    // Outer bar height
    const totalPct = bin.total > 0 ? (bin.total / maxCount) * 100 : 0;
    barWrapper.style.height = totalPct + "%";

    // Nested "included" portion
    if (bin.total > 0) {
      const includedPct = (bin.included / bin.total) * 100;
      const includedDiv = document.createElement("div");
      includedDiv.className = "included-portion";
      includedDiv.style.height = includedPct + "%";
      barWrapper.appendChild(includedDiv);
    }

    // Label in linear space
    const binStartLog = minLog + i * binSize;
    const binEndLog = minLog + (i + 1) * binSize;

    // Convert back to linear
    const lowerLin = Math.pow(10, binStartLog) - 1;
    const upperLin = Math.pow(10, binEndLog) - 1;

    const label = document.createElement("div");
    label.className = "bar-label";
    label.textContent = `${formatNumber(lowerLin)} – ${formatNumber(upperLin)}`;
    barWrapper.appendChild(label);

    chart.appendChild(barWrapper);
  });
}

/**
 * Called when the slider changes
 * sliderValue is in [globalMinLog, globalMaxLog].
 */
function onSliderChange(sliderValue) {
  // Convert the log slider value back to linear
  const logVal = parseFloat(sliderValue);
  const filterVal = Math.pow(10, logVal) - 1;

  drawHistogram(
    rawValues,
    logValues,
    filterVal,
    globalMinLog,
    globalMaxLog,
    BIN_COUNT,
  );
}

// Helper to format numbers nicely
function formatNumber(num) {
  if (num < 1) {
    return num.toFixed(2);
  } else if (num < 100) {
    return num.toFixed(1);
  } else {
    return Math.round(num).toString();
  }
}
