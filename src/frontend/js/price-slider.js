// price-slider.js

// price-slider.js

(function () {
  /**
   * Factory function to create a re-usable slider + histogram.
   *
   * @param {Object} options - Configuration object
   * @param {string} options.sliderId - DOM ID of the <input type="range">
   * @param {string} options.outputId - DOM ID of the <output> or <span> for displaying price
   * @param {string} options.chartId - DOM ID of the <div> to render the histogram
   * @param {number[]} options.dataArray - The raw numeric data for the histogram
   * @param {number} [options.minVal=10] - The minimum price for your mapping
   * @param {number} [options.midVal=200] - The midpoint for your custom exponent/linear break
   * @param {number} [options.maxVal=550] - The max price for your mapping
   * @param {number} [options.defaultValue=50] - Initial slider value (0..100)
   * @param {number} [options.binCount=30] - Number of histogram bins
   *
   * @returns {Object} - An object with methods: updateData(newArr), handleSliderChange(...), etc.
   */
  function createPriceSlider({
    sliderId,
    outputId,
    chartId,
    dataArray,
    minVal = 10,
    midVal = 200,
    maxVal = 550,
    defaultValue = 50,
    binCount = 30,
  }) {
    // Local state
    let rawValues = [];
    let logValues = [];
    let sliderInitialized = false;

    // Parse + store initial data
    updateDataArray(dataArray);

    // Grab DOM elements
    const sliderEl = document.getElementById(sliderId);
    const outputEl = document.getElementById(outputId);
    const chartEl = document.getElementById(chartId);

    // Initialize slider once
    if (sliderEl) {
      sliderEl.min = 0;
      sliderEl.max = 100;
      sliderEl.value = defaultValue;
      sliderEl.addEventListener("input", (evt) => {
        handleSliderChange(parseFloat(evt.target.value));
      });
      sliderInitialized = true;

      // Draw the initial histogram
      handleSliderChange(defaultValue);
    }

    /**
     * Parse a new data array, store it, re-draw histogram at current slider value.
     */
    function updateDataArray(newArray) {
      // E.g. if it's a JSON string, parse it:
      if (typeof newArray === "string") {
        try {
          newArray = JSON.parse(newArray);
        } catch (err) {
          console.error("Failed to parse newArray as JSON:", err);
          newArray = [];
        }
      }
      // Filter out invalid or <1
      rawValues = Array.isArray(newArray) ? newArray.filter((v) => v >= 1) : [];
      // Convert to log
      logValues = rawValues.map((v) => Math.log10(v));
    }

    /**
     * Called whenever the slider moves.
     */
    function handleSliderChange(linearValue) {
      const mappedVal = mapLinearToExponential(
        linearValue,
        minVal,
        midVal,
        maxVal,
      );

      // Re-draw histogram
      drawHistogram(rawValues, logValues, mappedVal, binCount);

      // Update the displayed price
      if (outputEl) {
        outputEl.textContent = `€${mappedVal.toFixed(0)}`;
      }
    }

    /**
     * Our custom exponent/linear mapping (0..100 => minVal..maxVal).
     */
    function mapLinearToExponential(sliderValue, minVal, midVal, maxVal) {
      const percentage = sliderValue / 100;
      if (percentage <= 0.7) {
        // First 70%: exponential from minVal to midVal
        return minVal * Math.pow(midVal / minVal, percentage / 0.7);
      } else {
        // Last 30%: linear from midVal to maxVal
        const newPct = (percentage - 0.7) / 0.3;
        return midVal + (maxVal - midVal) * newPct;
      }
    }

    /**
     * Render the histogram into chartEl
     */
    function drawHistogram(rVals, lVals, filterVal, binCount) {
      if (!chartEl) return;
      chartEl.innerHTML = "";

      if (!rVals.length) {
        chartEl.textContent = "";
        return;
      }

      // We'll do the same 70/30 approach for bin edges
      const bins = [];
      for (let i = 0; i <= binCount; i++) {
        const pct = i / binCount;
        if (pct <= 0.7) {
          bins.push(minVal * Math.pow(midVal / minVal, pct / 0.7));
        } else {
          const newPct = (pct - 0.7) / 0.3;
          bins.push(midVal + (maxVal - midVal) * newPct);
        }
      }

      // Initialize bin counts
      const binData = Array.from({ length: binCount }, () => ({
        total: 0,
        included: 0,
      }));

      // Place each data point into the correct bin
      for (let i = 0; i < rVals.length; i++) {
        const rv = rVals[i];
        // Find bin
        let binIndex = -1;
        for (let j = 0; j < binCount; j++) {
          if (rv >= bins[j] && rv < bins[j + 1]) {
            binIndex = j;
            break;
          }
        }
        if (binIndex === -1 && rv >= bins[binCount]) {
          binIndex = binCount - 1;
        }

        if (binIndex >= 0) {
          binData[binIndex].total++;
          if (rv <= filterVal) {
            binData[binIndex].included++;
          }
        }
      }

      // Find max count to scale bar heights
      const maxCount = Math.max(...binData.map((b) => b.total)) || 1;

      // Render each bin
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

        // Label
        const lowerVal = bins[i];
        const upperVal = bins[i + 1];
        const label = document.createElement("div");
        label.className = "bar-label";
        label.textContent = `${formatNumber(lowerVal)} – ${formatNumber(upperVal)}`;
        barWrapper.appendChild(label);

        chartEl.appendChild(barWrapper);
      });
    }

    function formatNumber(num) {
      if (num < 1) return num.toFixed(1);
      if (num < 100) return num.toFixed(0);
      return Math.round(num).toString();
    }

    // Return an object with some methods so the caller can update or destroy later.
    return {
      updateData(newArr) {
        updateDataArray(newArr);
        if (sliderInitialized && sliderEl) {
          const currentVal = parseFloat(sliderEl.value || defaultValue);
          handleSliderChange(currentVal);
        }
      },
      handleSliderChange,
      setValue(newVal) {
        // e.g. programmatically change slider
        if (sliderEl) {
          sliderEl.value = newVal;
          handleSliderChange(newVal);
        }
      },
    };
  }

  // Expose the factory function
  window.createPriceSlider = createPriceSlider;
})();
