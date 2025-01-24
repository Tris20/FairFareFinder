const BIN_COUNT = 30;

let rawValues = [];
let logValues = [];
let globalMinLog = 0;
let globalMaxLog = 0;

fetch("/js/dummy.csv")
  .then((response) => response.text())
  .then((csvText) => {
    const lines = csvText.trim().split("\n");
    lines.shift(); // Remove header

    // Parse floats, filter out invalid entries < 1
    rawValues = lines
      .map((line) => parseFloat(line.trim()))
      .filter((v) => !isNaN(v) && v >= 1);

    // Compute log10(price+1) if you want to shift zero out,
    // or log10(price) if all prices > 1. Let's do log10(v) for simplicity:
    logValues = rawValues.map((v) => Math.log10(v));

    globalMinLog = Math.log10(10);
    globalMaxLog = Math.log10(550);

    // Initialize slider: 0..100
    const slider = document.getElementById("accommodationPrice-slider0");
    slider.min = 0;
    slider.max = 100;
    slider.value = 50;

    // Draw initial histogram (assume 10..550)
    const initialVal = mapLinearToExponential(50, 10, 550);
    drawHistogram(rawValues, logValues, initialVal, BIN_COUNT);
    updatePriceOutput(50);
  })
  .catch((err) => console.error("Error:", err));

function handleSliderChange(sliderValue) {
  const filterVal = mapLinearToExponential(sliderValue, 10, 550);
  drawHistogram(rawValues, logValues, filterVal, BIN_COUNT);
  updatePriceOutput(sliderValue);
}

function updatePriceOutput(linearValue) {
  const mapped = mapLinearToExponential(linearValue, 10, 550);
  document.getElementById("accommodationOutput0").textContent =
    `€${mapped.toFixed(2)}`;
}

// --- Pure log slider mapping: 0 => minVal, 100 => maxVal
function mapLinearToExponential(sliderValue, minVal, maxVal) {
  const fraction = sliderValue / 100;
  const logMin = Math.log10(minVal);
  const logMax = Math.log10(maxVal);
  const logVal = logMin + fraction * (logMax - logMin);
  return Math.pow(10, logVal);
}

// Example histogram: log-based bins or however you prefer
function drawHistogram(rawVals, logVals, filterVal, binCount) {
  const chart = document.getElementById("chart");
  chart.innerHTML = "";

  if (!rawVals.length) {
    chart.textContent = "No data";
    return;
  }

  // Create bins in log space

  // Force the histogram range to 10..550
  const minVal = 10;
  const maxVal = 550;

  const minLog = Math.log10(minVal);
  const maxLog = Math.log10(maxVal);

  const logRange = maxLog - minLog;
  const binSize = logRange / binCount;

  const bins = Array.from({ length: binCount }, () => ({
    total: 0,
    included: 0,
  }));

  for (let i = 0; i < logVals.length; i++) {
    const lv = logVals[i];
    const rv = rawVals[i];

    let binIndex = Math.floor((lv - minLog) / binSize);
    if (binIndex < 0) binIndex = 0;
    if (binIndex >= binCount) binIndex = binCount - 1;

    bins[binIndex].total++;
    if (rv <= filterVal) {
      bins[binIndex].included++;
    }
  }

  const maxCount = Math.max(...bins.map((b) => b.total));

  bins.forEach((bin, i) => {
    const barWrapper = document.createElement("div");
    barWrapper.className = "bar";
    // scale bar by total
    const totalPct = (bin.total / maxCount) * 100;
    barWrapper.style.height = totalPct + "%";

    if (bin.total > 0) {
      const includedPct = (bin.included / bin.total) * 100;
      const includedDiv = document.createElement("div");
      includedDiv.className = "included-portion";
      includedDiv.style.height = includedPct + "%";
      barWrapper.appendChild(includedDiv);
    }

    // Label
    const binStartLog = minLog + i * binSize;
    const binEndLog = binStartLog + binSize;
    const lowerVal = Math.pow(10, binStartLog);
    const upperVal = Math.pow(10, binEndLog);

    const label = document.createElement("div");
    label.className = "bar-label";
    label.textContent = `${formatNumber(lowerVal)} – ${formatNumber(upperVal)}`;
    barWrapper.appendChild(label);

    chart.appendChild(barWrapper);
  });
}

function formatNumber(num) {
  if (num < 1) return num.toFixed(1);
  if (num < 100) return num.toFixed(0);
  return Math.round(num).toString();
}
