function setCookie(name, value, days) {
  const date = new Date();
  date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
  const expires = "expires=" + date.toUTCString();
  const cookieString = `${name}=${value}; ${expires}; path=/`;
  console.log("Setting cookie:", cookieString); // Debug log
  document.cookie = cookieString;
}

function getCookie(name) {
  console.log("Checking for cookie:", name); // Debug log
  const cookies = document.cookie.split("; ");
  console.log("All cookies:", cookies); // Debug log

  for (let i = 0; i < cookies.length; i++) {
    const [cookieName, cookieValue] = cookies[i].split("=");
    console.log("Checking cookie:", cookieName, cookieValue); // Debug log
    if (cookieName === name) {
      return cookieValue;
    }
  }
  return null; // Return null if not found
}

document.addEventListener("DOMContentLoaded", function () {
  // 1) Try to read the city from the cookie
  let savedCity = getCookie("selectedCity");
  if (!savedCity) {
    savedCity = "Berlin"; // fallback if no cookie
  }

  // 2) Build the HTMX request URL.
  //    For example, we do an initial flight price slider of 49 => ~€399
  //    (use your own defaults as needed).
  let flightPriceLinear = 49;
  let accomPriceLinear = 57;

  let url =
    "/filter?city[]=" +
    encodeURIComponent(savedCity) +
    "&maxFlightPriceLinear[]=" +
    flightPriceLinear +
    "&maxAccommodationPrice[]=" +
    accomPriceLinear +
    "&sort=best_weather";

  console.log("Auto-loading data with city:", savedCity, "via:", url);

  // 3) Fire an HTMX request to /filter, putting results into #results-container
  htmx.ajax("GET", url, "#flight-table");
});

// clear sort option selections
document.addEventListener("DOMContentLoaded", function () {
  const sortSelect = document.getElementById("sort");
  sortSelect.value = "best_weather"; // Default value
});
