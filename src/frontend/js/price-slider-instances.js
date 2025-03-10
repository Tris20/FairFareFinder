let accomSlider;
document.addEventListener("DOMContentLoaded", () => {
  window.accomSlider = createPriceSlider({
    sliderId: "accommodationPrice-slider0",
    outputId: "accommodationOutput0",
    chartId: "chart",
    dataArray: [],
    minVal: minAccomPrice,
    midVal: midAccomPrice,
    maxVal: maxAccomPrice,
    defaultValue: defaultAccomPrice,
    binCount: 30,
  });
});

let flightSlider;
document.addEventListener("DOMContentLoaded", () => {
  window.flightSlider = createPriceSlider({
    sliderId: "combinedPrice-slider0",
    outputId: "priceOutput0",
    chartId: "flight-chart",
    dataArray: [],
    minVal: minFlightPrice,
    midVal: midFlightPrice,
    maxVal: maxFlightPrice,
    defaultValue: 57,
    binCount: 30,
  });
});
