const BIN_COUNT = 30;

let rawValues = [];
let logValues = [];
let globalMinLin = 0;
let globalMaxLin = 0;
let globalMinLog = 0;
let globalMaxLog = 0;

// --- 1. Fetch CSV and initialize ---
fetch("/js/dummy.csv")
  .then((response) => response.text())
  .then((csvText) => {
    const lines = csvText.trim().split("\n");
    lines.shift(); // Remove header

    // Parse floats, filter out invalid entries < 1
    rawValues = lines
      .map((line) => parseFloat(line.trim()))
      .filter((v) => !isNaN(v) && v >= 1);

    // Linear stats
    globalMinLin = Math.min(...rawValues);
    globalMaxLin = Math.max(...rawValues);

    // Log values: log10(price+1)
    logValues = rawValues.map((val) => Math.log10(val + 1));
    globalMinLog = Math.min(...logValues);
    globalMaxLog = Math.max(...logValues);

    // Init slider
    const slider = document.getElementById("accommodationPrice-slider0");
    slider.min = 0;
    slider.max = 100;
    slider.value = 50;

    // Draw chart for initial filter
    const initialFilterVal = mapLinearToExponential(50, 10, 550);
    drawHistogram(
      rawValues,
      logValues,
      initialFilterVal,
      globalMinLog,
      globalMaxLog,
      BIN_COUNT,
    );
    updatePriceOutput(50);
  })
  .catch((err) => console.error("Error fetching CSV:", err));

function handleSliderChange(sliderValue) {
  const linearValue = parseFloat(sliderValue);
  const filterVal = mapLinearToExponential(linearValue, 10, 550);

  drawHistogram(
    rawValues,
    logValues,
    filterVal,
    globalMinLog,
    globalMaxLog,
    BIN_COUNT,
  );
  updatePriceOutput(linearValue);
}

function updatePriceOutput(linearValue) {
  const mappedValue = mapLinearToExponential(linearValue, 10, 550);
  const outputElement = document.getElementById("accommodationOutput0");
  outputElement.textContent = `€${mappedValue.toFixed(2)}`;
}

// Pure log-based mapping: 0 => minVal, 100 => maxVal
function mapLinearToExponential(sliderValue, minVal, maxVal) {
  const fraction = sliderValue / 100;
  const logMin = Math.log10(minVal);
  const logMax = Math.log10(maxVal);
  const logVal = logMin + fraction * (logMax - logMin);
  return Math.pow(10, logVal);
}

// Draw the histogram
function drawHistogram(rawVals, logVals, filterVal, minLog, maxLog, binCount) {
  const chart = document.getElementById("chart");
  chart.innerHTML = "";

  if (!rawVals.length) {
    chart.textContent = "No data";
    return;
  }

  const logRange = maxLog - minLog;
  const binSize = logRange / binCount;

  let bins = Array.from({ length: binCount }, () => ({
    total: 0,
    included: 0,
  }));

  for (let i = 0; i < logVals.length; i++) {
    const lv = logVals[i];
    const rv = rawVals[i];

    let binIndex = Math.floor((lv - minLog) / binSize);
    if (binIndex < 0) binIndex = 0;
    if (binIndex >= binCount) binIndex = binCount - 1;

    bins[binIndex].total += 1;
    if (rv <= filterVal) {
      bins[binIndex].included += 1;
    }
  }

  const maxCount = Math.max(...bins.map((b) => b.total));

  bins.forEach((bin, i) => {
    const barWrapper = document.createElement("div");
    barWrapper.className = "bar";

    const totalPct = bin.total > 0 ? (bin.total / maxCount) * 100 : 0;
    barWrapper.style.height = totalPct + "%";

    if (bin.total > 0) {
      const includedPct = (bin.included / bin.total) * 100;
      const includedDiv = document.createElement("div");
      includedDiv.className = "included-portion";
      includedDiv.style.height = includedPct + "%";
      barWrapper.appendChild(includedDiv);
    }

    const binStartLog = minLog + i * binSize;
    const binEndLog = minLog + (i + 1) * binSize;
    const lowerLin = Math.pow(10, binStartLog) - 1;
    const upperLin = Math.pow(10, binEndLog) - 1;

    const label = document.createElement("div");
    label.className = "bar-label";
    label.textContent = `${formatNumber(lowerLin)} – ${formatNumber(upperLin)}`;
    barWrapper.appendChild(label);

    chart.appendChild(barWrapper);
  });
}

function formatNumber(num) {
  if (num < 1) {
    return num.toFixed(2);
  } else if (num < 100) {
    return num.toFixed(1);
  } else {
    return Math.round(num).toString();
  }
}
