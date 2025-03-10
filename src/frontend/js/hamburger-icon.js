const button = document.querySelector(".button-three");
const flightForm = document.getElementById("flight-form");

// Global variable to track a pending timeout.
let transitionTimeoutId = null;

button.addEventListener("click", () => {
  const currentState = button.getAttribute("data-state");

  if (currentState === "opened") {
    // If the form is open, transition to closed.
    // Clear any pending timeout if a user clicks during an animation.
    if (transitionTimeoutId) {
      clearTimeout(transitionTimeoutId);
      transitionTimeoutId = null;
    }

    button.setAttribute("data-state", "closed");
    button.setAttribute("aria-expanded", "false");
    flightForm.classList.add("hidden");

    // After 1 second (transition duration), hide the element from layout.
    transitionTimeoutId = setTimeout(() => {
      flightForm.style.display = "none";
      transitionTimeoutId = null;
    }, 1000);
  } else {
    // If the form is closed, transition to open.
    if (transitionTimeoutId) {
      // Cancel the pending hide if it's in progress.
      clearTimeout(transitionTimeoutId);
      transitionTimeoutId = null;
    }

    // Make sure the form is part of the layout before animating.
    flightForm.style.display = "flex";

    // Force a reflow so that the display change is registered.
    flightForm.offsetWidth;

    flightForm.classList.remove("hidden");
    button.setAttribute("data-state", "opened");
    button.setAttribute("aria-expanded", "true");
  }
});
