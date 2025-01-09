// --------------------- Global Variables ---------------------
let cities = [];
let additionalCityCount = 0;
let rowCount = 1;

// --------------------- Functions ---------------------
function toggleDurationVisibility() {
  // Select all elements with the class "destination-duration"
  const durationContainers = document.querySelectorAll(".destination-duration");

  durationContainers.forEach((durationContainer) => {
    if (additionalCityCount > 0) {
      console.log("Hiding duration-container");
      durationContainer.style.display = "none";
    } else {
      console.log("Showing duration-container");
      durationContainer.style.display = "block";
    }
  });
}

// --------------------- Event Listeners ---------------------

// 1) Remove City Row (event delegation on #city-rows)
document
  .getElementById("city-rows")
  .addEventListener("click", function (event) {
    if (event.target.classList.contains("remove-city-button")) {
      const cityRow = event.target.closest(".city-row");
      cityRow.remove();
      additionalCityCount--;
      toggleDurationVisibility();
    }
  });

// 2) Add City Row
document
  .getElementById("add-city-button")
  .addEventListener("click", function () {
    const cityRows = document.getElementById("city-rows");
    const div = document.createElement("div");
    div.className = "form-group city-row";
    div.innerHTML = `
    <button type="button" class="remove-city-button">-</button>
    <select name="logical_operator[]">
      <option value="AND">AND</option>
      <option value="OR">OR</option>
    </select>
    <label>City:</label>
    <div class="dropdown-container">
      <input
        id="city-search"
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
    <label for="priceOutput${rowCount}"></label>
    <output id="priceOutput${rowCount}" class="output-range">€399.00</output>
    <input
      type="range"
      id="combinedPrice-slider${rowCount}"
      name="maxPriceLinear[]"
      min="0"
      max="100"
      step="1"
      value="49"
      class="price-slider"
      hx-get="/update-slider-price"
      hx-target="#priceOutput${rowCount}"
      hx-trigger="input"
      hx-push-url="false"
      hx-preserve="false"
      hx-include="#combinedPrice-slider${rowCount}"
      autocomplete="off"
    />
  `;
    cityRows.appendChild(div);
    htmx.process(div);
    rowCount++;

    // Initialize dropdown for the new input
    const input = div.querySelector(".dropdown-input");
    const dropdown = div.querySelector(".dropdown-list");
    const button = div.querySelector(".dropdown-btn");

    // Reuse the dropdown logic for the new city input
    let highlightedIndex = -1;

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

          // Trigger HTMX-compatible "change" event
          input.dispatchEvent(new Event("change", { bubbles: true }));
        });

        dropdown.appendChild(li);
      });
    }

    input.addEventListener("input", () => {
      const value = input.value.toLowerCase();
      const filteredCities = cities.filter(({ city, country }) =>
        `${city}, ${country}`.toLowerCase().includes(value),
      );
      highlightedIndex = -1; // Reset the highlighted index
      populateDropdown(filteredCities);
      dropdown.classList.remove("hidden");
    });

    button.addEventListener("click", () => {
      if (dropdown.classList.contains("hidden")) {
        populateDropdown(cities); // Populate with all cities
        dropdown.classList.remove("hidden");
      } else {
        dropdown.classList.add("hidden");
      }
    });

    input.addEventListener("blur", () => {
      setTimeout(() => dropdown.classList.add("hidden"), 200); // Hide dropdown

      const value = input.value.trim();
      const isValid = cities.some(
        ({ city, country }) =>
          `${city}, ${country}`.toLowerCase() === value.toLowerCase(),
      );

      if (!isValid) {
        console.error("Invalid city on blur:", value);
        input.value = ""; // Clear invalid input
        return;
      }

      console.log("Valid city on blur:", value);

      // Trigger HTMX-compatible "change" event
      input.dispatchEvent(new Event("change", { bubbles: true }));
    });

    rowCount++;
    additionalCityCount++;
    console.log("New additionalCityCount:", additionalCityCount);
    toggleDurationVisibility();
  });

// 3) DOMContentLoaded - Search bar for Origin Cities & cookie handling
document.addEventListener("DOMContentLoaded", () => {
  // Check if cookie for search
  const input = document.getElementById("city-search");
  const savedCity = getCookie("selectedCity");
  if (savedCity) {
    // Validate the saved city before using it
    const isValidCity = cities.some(
      ({ city }) => city.toLowerCase() === savedCity.toLowerCase(),
    );

    if (isValidCity) {
      input.value = savedCity; // Set input to the saved city from the cookie
      console.log("Loaded valid city from cookie:", savedCity);
    } else {
      console.warn("Invalid city in cookie, defaulting to Berlin.");
      input.value = "Berlin"; // Set to default value if cookie is invalid
    }
  } else {
    console.log("No cookie found, defaulting to Berlin.");
    input.value = "Berlin"; // Set to default value if no cookie exists
  }

  // Search bar inputs
  const dropdown = document.getElementById("city-list");
  const button = document.getElementById("dropdown-btn");
  let highlightedIndex = -1; // Index of the highlighted item

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

        // Set the cookie when a valid city is selected
        setCookie("selectedCity", li.textContent, 7);
        console.log(
          "Cookie set for city selected via dropdown:",
          li.textContent,
        );

        // Trigger HTMX-compatible "change" event
        input.dispatchEvent(new Event("change", { bubbles: true }));
      });

      dropdown.appendChild(li);
    });
  }

  function highlightItem(index) {
    const items = dropdown.querySelectorAll("li");
    if (items.length === 0) return;

    // Remove highlight from all items
    items.forEach((item) => item.classList.remove("highlighted"));

    // Highlight the new item
    if (index >= 0 && index < items.length) {
      items[index].classList.add("highlighted");
      highlightedIndex = index;

      // Ensure the highlighted item is visible in the dropdown
      items[index].scrollIntoView({ block: "nearest" });
    }
  }

  input.addEventListener("input", () => {
    const value = input.value.toLowerCase();
    const filteredCities = cities.filter(({ city, country }) =>
      `${city}, ${country}`.toLowerCase().includes(value),
    );
    highlightedIndex = -1; // Reset the highlighted index
    populateDropdown(filteredCities);
    dropdown.classList.remove("hidden");
  });

  input.addEventListener("keydown", (event) => {
    const items = dropdown.querySelectorAll("li");
    if (items.length === 0) return;

    if (event.key === "ArrowDown") {
      // Highlight the next item
      event.preventDefault(); // Prevent cursor movement
      highlightItem(highlightedIndex + 1);
    } else if (event.key === "ArrowUp") {
      // Highlight the previous item
      event.preventDefault(); // Prevent cursor movement
      highlightItem(highlightedIndex - 1);
    } else if (event.key === "Enter") {
      // Select the highlighted item
      event.preventDefault(); // Prevent form submission
      if (highlightedIndex >= 0 && highlightedIndex < items.length) {
        items[highlightedIndex].click(); // Simulate a click
      }
    }
  });

  button.addEventListener("click", () => {
    if (dropdown.classList.contains("hidden")) {
      populateDropdown(cities); // Populate with all cities
      dropdown.classList.remove("hidden");
    } else {
      dropdown.classList.add("hidden");
    }
  });

  input.addEventListener("blur", () => {
    const value = input.value.trim();
    const isValidCity = cities.some(
      ({ city }) => city.toLowerCase() === value.toLowerCase(),
    );

    if (isValidCity) {
      setCookie("selectedCity", value, 7); // Save only valid city
      console.log("Valid city saved:", value);
    } else {
      console.warn("Invalid city, not saving:", value);
    }
    // Trigger HTMX-compatible "change" event
    input.dispatchEvent(new Event("change", { bubbles: true }));
  });

  // Cookie Monsters
  // Fetch cities from the backend
  fetch("/city-country-pairs")
    .then((response) => response.json())
    .then((data) => {
      cities = data; // Populate cities array
      console.log("Cities loaded:", cities);

      // Validate and load the selectedCity cookie
      const savedCity = getCookie("selectedCity");
      if (savedCity) {
        const isValidCity = cities.some(
          ({ city }) => city.toLowerCase() === savedCity.toLowerCase(),
        );

        if (isValidCity) {
          input.value = savedCity;
          console.log("Loaded valid city from cookie:", savedCity);
        } else {
          console.warn("Invalid city in cookie, defaulting to Berlin.");
          input.value = "Berlin";
        }
      } else {
        console.log("No cookie found, defaulting to Berlin.");
        input.value = "Berlin";
      }

      // Populate dropdown with initial cities
      populateDropdown(cities);
    })
    .catch((error) => console.error("Error loading cities:", error));
});
