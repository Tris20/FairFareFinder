let hasFadedIn = false; // Flag to check if the table has already faded in
if ("scrollRestoration" in history) {
  history.scrollRestoration = "manual";
}
document.addEventListener("DOMContentLoaded", function () {
  // After the fade-out animation ends, hide the website-name and show the table
  setTimeout(function () {
    document.getElementById("website-name").style.display = "none";

    // Show the table content with the fade-in animation and move it up
    const tableContainer = document.getElementById("table-container");
    tableContainer.classList.add("show");

    // Also show the header and move it with the form
    const pageHeader = document.getElementById("page-banner");
    pageHeader.classList.add("show");

    // Wait for the 'moveUp' animation to finish (animationend event)
    tableContainer.addEventListener("animationend", function (event) {
      if (event.animationName === "moveUp") {
        // Manually trigger the form submission after the animation is complete
        htmx.trigger("#flight-form", "change");

        // *** Trigger update of flight sliders once the table has moved up ***
        updateFlightSliders();
      }
    });

    // Handle HTMX table swap event for the first time only
    document
      .getElementById("flight-table")
      .addEventListener("htmx:afterSwap", function () {
        if (!hasFadedIn) {
          const flightTable = document.getElementById("flight-table");
          flightTable.classList.add("fade-in");
          hasFadedIn = true;

          // Show the footer after the table has been loaded
          const footer = document.querySelector("#footer-privacy");
          footer.classList.add("fade-in");
        }
        // Also update flight sliders after any HTMX swap on flight-table:
        updateFlightSliders();
      });

    // Show the background image with a fade-in effect
    document.getElementById("bg-image-dreams").style.opacity = "1";
  }, 2000);
});
