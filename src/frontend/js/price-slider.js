// price-slider.js
(function () {
  // We'll keep some private state:
  let rawValues = [];
  let logValues = [];

  let sliderInitialized = false;

  /**
   * Called to initialize the slider once.
   */
  function initSlider() {
    const slider = document.getElementById("accommodationPrice-slider0");
    if (!slider) return;

    slider.min = 0;
    slider.max = 100;
    slider.value = 53.57;
    slider.addEventListener("input", () => handleSliderChange(slider.value));

    sliderInitialized = true;
  }

  /**
   * 1) Save new accommodation prices
   * 2) (lazy) initialize slider if needed
   * 3) Draw histogram
   */
  function updateAccommodationPrices(newArray) {
    console.log("Inside updateAccommodationPrices, got:", newArray);

    if (typeof newArray === "string") {
      try {
        newArray = JSON.parse(newArray);
        console.log("Parsed newArray:", newArray);
      } catch (error) {
        console.error("Failed to parse newArray as JSON:", error);
        newArray = []; // Fallback to an empty array on failure
      }
    }

    console.log("newArray:", newArray);
    console.log("typeof :", typeof newArray);
    console.log("typeof first element is:", typeof newArray[0]);

    rawValues = newArray
      .map((str) => parseFloat(str)) // parse each string
      .filter((num) => !isNaN(num) && num >= 1);
    // fallback if not actually an array
    if (!Array.isArray(newArray)) newArray = [];
    rawValues = newArray.filter((v) => v >= 1);
    logValues = rawValues.map((v) => Math.log10(v));
    console.log("Inside updateAccommodationPrices, got RAW:", rawValues);

    // If we haven't set up the slider yet, do so
    if (!sliderInitialized) {
      initSlider();

      // The very first time we do want to set a default:
      const defaultVal = 53.57;
      drawHistogram(
        rawValues,
        logValues,
        mapLinearToExponential(defaultVal, 10, 550),
        30,
      );
      //updatePriceOutput(defaultVal);
    } else {
      // Already initialized; preserve the slider's current value
      const slider = document.getElementById("accommodationPrice-slider0");
      if (slider) {
        const currentVal = slider.value;
        drawHistogram(
          rawValues,
          logValues,
          mapLinearToExponential(currentVal, 10, 550),
          30,
        );
      }
    }
  }

  function handleSliderChange(sliderValue) {
    const filterVal = mapLinearToExponential(sliderValue, 10, 550);
    drawHistogram(rawValues, logValues, filterVal, 30);
    updatePriceOutput(sliderValue);
  }

  function updatePriceOutput(linearValue) {
    const mapped = mapLinearToExponential(linearValue, 10, 550);
    const outEl = document.getElementById("accommodationOutput0");
    if (outEl) {
      outEl.textContent = `€${mapped.toFixed(0)}`;
    }
  }

  function mapLinearToExponential(sliderValue, minVal, maxVal) {
    const midVal = 200.0; // Midpoint value at 70% of the slider
    const percentage = sliderValue / 100;

    if (percentage <= 0.7) {
      // First 70%: Exponential mapping from minVal to midVal
      return minVal * Math.pow(midVal / minVal, percentage / 0.7);
    } else {
      // Last 30%: Linear interpolation from midVal to maxVal
      const newPercentage = (percentage - 0.7) / 0.3;
      return midVal + (maxVal - midVal) * newPercentage;
    }
  }

  function drawHistogram(rVals, lVals, filterVal, binCount) {
    const chart = document.getElementById("chart");
    if (!chart) return;
    chart.innerHTML = "";

    if (!rVals.length) {
      chart.textContent = "No data";
      return;
    }

    const minVal = 10;
    const maxVal = 550;
    const midVal = 200.0;

    // Calculate bin boundaries using the 70/30 split logic
    const bins = [];
    for (let i = 0; i <= binCount; i++) {
      const percentage = i / binCount;
      if (percentage <= 0.7) {
        bins.push(minVal * Math.pow(midVal / minVal, percentage / 0.7));
      } else {
        const newPercentage = (percentage - 0.7) / 0.3;
        bins.push(midVal + (maxVal - midVal) * newPercentage);
      }
    }

    // Initialize bin counts
    const binData = Array.from({ length: binCount }, () => ({
      total: 0,
      included: 0,
    }));

    // Populate bins with data
    for (let i = 0; i < rVals.length; i++) {
      const rv = rVals[i];
      const lv = lVals[i];

      // Find the correct bin for the value
      let binIndex = -1;
      for (let j = 0; j < binCount; j++) {
        if (rv >= bins[j] && rv < bins[j + 1]) {
          binIndex = j;
          break;
        }
      }
      if (binIndex === -1 && rv >= bins[binCount]) binIndex = binCount - 1; // Handle edge case

      if (binIndex >= 0) {
        binData[binIndex].total++;
        if (rv <= filterVal) {
          binData[binIndex].included++;
        }
      }
    }

    // Determine the maximum count for normalization
    const maxCount = Math.max(...binData.map((b) => b.total));

    // Render the histogram
    binData.forEach((bin, i) => {
      const barWrapper = document.createElement("div");
      barWrapper.className = "bar";
      const totalPct = (bin.total / maxCount) * 100;
      barWrapper.style.height = totalPct + "%";

      if (bin.total > 0) {
        const includedPct = (bin.included / bin.total) * 100;
        const includedDiv = document.createElement("div");
        includedDiv.className = "included-portion";
        includedDiv.style.height = includedPct + "%";
        barWrapper.appendChild(includedDiv);
      }

      // Label the bin with its range
      const lowerVal = bins[i];
      const upperVal = bins[i + 1];
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

  // Expose just one global function
  window.updateAccommodationPrices = updateAccommodationPrices;
  window.handleSliderChange = handleSliderChange;
})();
