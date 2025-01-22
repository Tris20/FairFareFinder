const BIN_COUNT = 30;

let rawValues = [];
let logValues = [];
let globalMinLin = 0;
let globalMaxLin = 0;
let globalMinLog = 0;
let globalMaxLog = 0;

// Fetch and process the CSV data
fetch("/js/dummy.csv")
  .then((response) => response.text())
  .then((csvText) => {
    const lines = csvText.trim().split("\n");
    lines.shift(); // Remove header

    rawValues = lines
      .map((line) => parseFloat(line.trim()))
      .filter((v) => !isNaN(v) && v >= 1);

    globalMinLin = Math.min(...rawValues);
    globalMaxLin = Math.max(...rawValues);

    logValues = rawValues.map((v) => Math.log10(v + 1));
    globalMinLog = Math.min(...logValues);
    globalMaxLog = Math.max(...logValues);

    const slider = document.getElementById("accommodationPrice-slider0");
    slider.min = globalMinLog;
    slider.max = globalMaxLog;
    slider.value = globalMaxLog;

    const initialFilterVal = Math.pow(10, globalMaxLog) - 1;
    drawHistogram(
      rawValues,
      logValues,
      initialFilterVal,
      globalMinLog,
      globalMaxLog,
      BIN_COUNT,
    );

    updatePriceOutput(globalMaxLog); // Initialize price output
  })
  .catch((err) => console.error("Error fetching CSV:", err));

function handleSliderChange(sliderValue) {
  const logVal = parseFloat(sliderValue);
  const filterVal = Math.pow(10, logVal) - 1;

  // Update the chart
  drawHistogram(
    rawValues,
    logValues,
    filterVal,
    globalMinLog,
    globalMaxLog,
    BIN_COUNT,
  );

  // Update price output
  updatePriceOutput(logVal);
}

// Function to update the price output display
function updatePriceOutput(logValue) {
  const linearValue = Math.pow(10, logValue) - 1;
  const outputElement = document.getElementById("accommodationOutput0");
  outputElement.textContent = `€${linearValue.toFixed(2)}`;
}

// Draw the histogram (unchanged)
function drawHistogram(rawVals, logVals, filterVal, minLog, maxLog, binCount) {
  const chart = document.getElementById("chart");
  chart.innerHTML = "";

  if (rawVals.length === 0) {
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
    if (rv <= filterVal) bins[binIndex].included += 1;
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

// Helper function to format numbers
function formatNumber(num) {
  if (num < 1) {
    return num.toFixed(2);
  } else if (num < 100) {
    return num.toFixed(1);
  } else {
    return Math.round(num).toString();
  }
}
