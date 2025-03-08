// --------------------- Global Variables ---------------------
let cities = [];
let additionalCityCount = 0;
let rowCount = 1;

// --------------------- Functions ---------------------
/* Hide the duration when there is more than one city as input(because we don't have ways to disply the duration from different origins yet)*/
function toggleDurationVisibility() {
  const durationContainers = document.querySelectorAll(".destination-duration");
  // Count how many .city-row elements are in the DOM right now
  const cityRowsCount = document.querySelectorAll(".city-row").length;

  durationContainers.forEach((durationContainer) => {
    if (cityRowsCount > 1) {
      console.log("Hiding duration-container");
      durationContainer.style.display = "none";
    } else {
      console.log("Showing duration-container");
      durationContainer.style.display = "block";
    }
  });
}

// Helper function to set up city search dropdown behavior
function setupCitySearch({
  input,
  dropdown,
  button,
  shouldSaveCookie = false,
}) {
  let highlightedIndex = -1;

  // Utility: Populate the dropdown list
  function populateDropdown(filteredCities) {
    dropdown.innerHTML = ""; // Clear current list
    filteredCities.forEach(({ city, country }, index) => {
      const li = document.createElement("li");
      li.textContent = `${city}`;
      li.classList.add("dropdown-item");

      // Highlight the item if it matches the current index
      if (index === highlightedIndex) {
        li.classList.add("highlighted");
      }

      li.addEventListener("click", () => {
        input.value = li.textContent; // Set input value
        dropdown.classList.add("hidden"); // Hide dropdown

        // If this search bar should save cookies, do it here
        if (shouldSaveCookie) {
          setCookie("selectedCity", li.textContent, 7);
          console.log(
            "Cookie set for city selected via dropdown:",
            li.textContent,
          );
        }

 updateFlightSliders();
        // Trigger HTMX-compatible "change" event
        input.dispatchEvent(new Event("change", { bubbles: true }));
      });

      dropdown.appendChild(li);
    });

    // Automatically highlight the first item (if any exist)
    if (filteredCities.length > 0) {
      highlightItem(0);
    } else {
      highlightedIndex = -1;
    }
  }

  // Utility: Highlight the selected dropdown item
  function highlightItem(index) {
    const items = dropdown.querySelectorAll("li");
    if (items.length === 0) return;

    // Remove highlight from all items
    items.forEach((item) => item.classList.remove("highlighted"));

    // Make sure the new index is within the range, and wrap around if not
    if (index < 0) {
      index = items.length - 1; // Wrap to the last item
    } else if (index >= items.length) {
      index = 0; // Wrap back to the first item
    }

    // Highlight the new item
    items[index].classList.add("highlighted");
    highlightedIndex = index;

    // Ensure the highlighted item is visible in the dropdown
    items[index].scrollIntoView({ block: "nearest" });
  }

  // Handle the input event
  input.addEventListener("input", () => {
    const value = input.value.toLowerCase();
    const filteredCities = cities.filter(({ city, country }) =>
      `${city}, ${country}`.toLowerCase().includes(value),
    );
    highlightedIndex = -1; // Reset the highlighted index
    populateDropdown(filteredCities);
    dropdown.classList.remove("hidden");
  });

  // Handle arrow key navigation + enter key
  input.addEventListener("keydown", (event) => {
    const items = dropdown.querySelectorAll("li");
    if (items.length === 0) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      highlightItem(highlightedIndex + 1);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      highlightItem(highlightedIndex - 1);
    } else if (event.key === "Enter") {
      event.preventDefault();
      if (highlightedIndex >= 0 && highlightedIndex < items.length) {
        items[highlightedIndex].click(); // Simulate a click
      }
    }
  });

  // Handle dropdown button click (expand/collapse)
  button.addEventListener("click", () => {
    if (dropdown.classList.contains("hidden")) {
      // Show all cities
      populateDropdown(cities);
      dropdown.classList.remove("hidden");
    } else {
      dropdown.classList.add("hidden");
    }
  });

  // Hide dropdown when user clicks anywhere outside of input, dropdown, or button
  document.addEventListener("click", (event) => {
    if (
      !input.contains(event.target) &&
      !dropdown.contains(event.target) &&
      !button.contains(event.target)
    ) {
      dropdown.classList.add("hidden");
    }
  });

  // Handle blur event
  input.addEventListener("blur", () => {
    const value = input.value.trim();
    const isValidCity = cities.some(
      ({ city }) => city.toLowerCase() === value.toLowerCase(),
    );

    if (isValidCity && shouldSaveCookie) {
      setCookie("selectedCity", value, 7);
      console.log("Valid city saved (cookie):", value);
    } else if (!isValidCity && shouldSaveCookie) {
      console.warn("Invalid city, not saving:", value);
    }
    // Trigger HTMX-compatible "change" event
    input.dispatchEvent(new Event("change", { bubbles: true }));
  });
}

// --------------------- Event Listeners ---------------------

// 1) Remove City Row (event delegation on #city-rows)

document
  .getElementById("city-rows") // Parent container wrapping all city rows
  .addEventListener("click", function (event) {
    if (event.target.classList.contains("remove-city-button")) {
      // Find the button's parent and then the top-level .form-group.city-row
      const cityRow = event.target.closest(".form-group.city-row"); // Start from the button's immediate parent
      if (cityRow && cityRow.classList.contains("operators")) {
        const outerCityRow = cityRow.parentElement; // Move to the outer .form-group.city-row
        if (outerCityRow && outerCityRow.classList.contains("city-row")) {
          outerCityRow.remove(); // Remove the entire row
          additionalCityCount--; // Adjust count
          toggleDurationVisibility(); // Update UI
        }
      }
    }
  });

// 2) Add City Row

document
  .getElementById("add-city-button")
  .addEventListener("click", function () {
    // 1) Get the city from the first row’s input
    const firstRowInput = document.querySelector(".city-row .dropdown-input");
    let firstRowCityValue = firstRowInput ? firstRowInput.value.trim() : "";

    // 2) Validate the first-row city. If invalid, default to "Berlin"
    const isValidCity = cities.some(
      ({ city }) => city.toLowerCase() === firstRowCityValue.toLowerCase(),
    );
    if (!isValidCity) {
      firstRowCityValue = "Berlin";
    }

    const cityRows = document.getElementById("city-rows");
    const div = document.createElement("div");
    div.className = "form-group city-row";

    // Build the innerHTML without any inline <script> tag.
    div.innerHTML = `
    <div class="form-group city-row operators">
      <button type="button" class="remove-city-button">-</button>
      <select class="logical-operator" name="logical_operator[]">
        <option value="AND">AND</option>
        <option value="OR">OR</option>
      </select>

      <div class="dropdown-container">
        <!-- Use unique IDs if needed; for example, include rowCount in the id -->
        <input
          id="city-search-${rowCount}"
          class="dropdown-input"
          name="city[]"
          placeholder="Search for a city"
          autocomplete="off"
        />
        <button class="dropdown-btn" type="button">
          <span class="caret">▼</span>
        </button>
        <ul class="dropdown-list hidden"></ul>
      </div>
    </div>
    <output id="priceOutput${rowCount}" class="output-range">€399</output>
    <div class="flight-price-slider">
      <div class="chart-container">
        <div class="chart" id="flight-chart${rowCount}"></div>
      </div>
      <input
        type="range"
        id="combinedPrice-slider${rowCount}"
        name="maxFlightPriceLinear[]"
        min="0"
        max="100"
        step="0.01"
        value="49"
        class="price-slider"
        hx-push-url="false"
        hx-preserve="false"
        hx-include="#combinedPrice-slider${rowCount}"
        autocomplete="off"
        oninput="window['flightSlider${rowCount}'].updateData(window.allFlightPrices[${rowCount}]);"
      />
    </div>
  `;

    // Append and process the new row with HTMX (if needed)
    cityRows.appendChild(div);
    htmx.process(div);

    // Now initialize the new flight slider programmatically:
    window["flightSlider" + rowCount] = createPriceSlider({
      sliderId: "combinedPrice-slider" + rowCount,
      outputId: "priceOutput" + rowCount,
      chartId: "flight-chart" + rowCount,
      dataArray: [],
      minVal: minFlightPrice,
      midVal: midFlightPrice,
      maxVal: maxFlightPrice,
      defaultValue: 57,
      binCount: 30,
    });

    // Initialize the dropdown search functionality for the new row
    const input = div.querySelector(".dropdown-input");
    const dropdown = div.querySelector(".dropdown-list");
    const button = div.querySelector(".dropdown-btn");

    // Set the input value to the first row's value
    input.value = firstRowCityValue;
    setupCitySearch({
      input,
      dropdown,
      button,
      shouldSaveCookie: false,
    });

    window["flightSlider" + rowCount].updateData(window.allFlightPrices);
    // Increment rowCount only once per row addition
    rowCount++;
    additionalCityCount++;
    console.log("New additionalCityCount:", additionalCityCount);
    toggleDurationVisibility();
    // Trigger the slider update with current data:
  });
// 3) DOMContentLoaded - Search bar for Origin Cities & cookie handling

document.addEventListener("DOMContentLoaded", () => {
  // Grab DOM elements for the main search bar
  const input = document.getElementById("city-search");
  const dropdown = document.getElementById("city-list");
  const button = document.getElementById("dropdown-btn");

  // 1) Fetch city list from backend
  fetch("/city-country-pairs")
    .then((response) => response.json())
    .then((data) => {
      cities = data; // Populate the global 'cities' array
      console.log("Cities loaded:", cities);

      // 2) Read cookie once (AFTER cities are fetched)
      const savedCity = getCookie("selectedCity");

      if (savedCity) {
        // Check if the cookie value is valid
        const isValidCity = cities.some(
          ({ city }) => city.toLowerCase() === savedCity.toLowerCase(),
        );

        if (isValidCity) {
          // If valid, set it
          input.value = savedCity;
          console.log("Loaded valid city from cookie:", savedCity);
        } else {
          // If invalid, default to Berlin
          console.warn("Invalid city in cookie, defaulting to Berlin.");
          input.value = "Berlin";
        }
      } else {
        // If no cookie exists, default to Berlin
        console.log("No cookie found, defaulting to Berlin.");
        input.value = "Berlin";
      }

      // Optional: if you have a function like populateDropdown to show all cities
      // populateDropdown(cities);
    })
    .catch((error) => console.error("Error loading cities:", error));

  // 3) Set up the city search logic with cookie saving enabled
  setupCitySearch({
    input,
    dropdown,
    button,
    shouldSaveCookie: true,
  });
});

// Listen for the HTMX afterSwap event on the document or a parent container
document.body.addEventListener("htmx:afterSwap", function (event) {
  // After HTMX swaps in the server response, re-check city rows
  toggleDurationVisibility();
});



