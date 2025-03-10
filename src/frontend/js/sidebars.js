document.addEventListener("DOMContentLoaded", () => {
  // LEFT SIDEBAR TOGGLE
  const toggleLeftButton = document.getElementById("toggleLeftSidebar");
  const leftSidebar = document.getElementById("left-sidebar");
  let leftTransitionTimeoutId = null;

  toggleLeftButton.addEventListener("click", () => {
    const currentState = toggleLeftButton.getAttribute("data-state");
    if (currentState === "opened") {
      // Close left sidebar.
      if (leftTransitionTimeoutId) {
        clearTimeout(leftTransitionTimeoutId);
        leftTransitionTimeoutId = null;
      }
      toggleLeftButton.setAttribute("data-state", "closed");
      toggleLeftButton.setAttribute("aria-expanded", "false");

      leftSidebar.classList.add("closed");

      leftTransitionTimeoutId = setTimeout(() => {
        leftSidebar.style.display = "none";
        leftTransitionTimeoutId = null;
      }, 1000);
    } else {
      // Open left sidebar.
      if (leftTransitionTimeoutId) {
        clearTimeout(leftTransitionTimeoutId);
        leftTransitionTimeoutId = null;
      }
      leftSidebar.style.display = "block";
      leftSidebar.offsetWidth; // Force reflow.
      leftSidebar.classList.remove("closed");
      toggleLeftButton.setAttribute("data-state", "opened");
      toggleLeftButton.setAttribute("aria-expanded", "true");
    }
  });

  // RIGHT SIDEBAR TOGGLE
  const toggleRightButton = document.getElementById("toggleRightSidebar");
  const rightSidebar = document.getElementById("right-sidebar");
  let rightTransitionTimeoutId = null;

  toggleRightButton.addEventListener("click", () => {
    const currentState = toggleRightButton.getAttribute("data-state");
    if (currentState === "opened") {
      // Close right sidebar.
      if (rightTransitionTimeoutId) {
        clearTimeout(rightTransitionTimeoutId);
        rightTransitionTimeoutId = null;
      }
      toggleRightButton.setAttribute("data-state", "closed");
      toggleRightButton.setAttribute("aria-expanded", "false");

      rightSidebar.classList.add("closed");

      rightTransitionTimeoutId = setTimeout(() => {
        rightSidebar.style.display = "none";
        rightTransitionTimeoutId = null;
      }, 1000);
    } else {
      // Open right sidebar.
      if (rightTransitionTimeoutId) {
        clearTimeout(rightTransitionTimeoutId);
        rightTransitionTimeoutId = null;
      }
      rightSidebar.style.display = "block";
      rightSidebar.offsetWidth; // Force reflow.
      rightSidebar.classList.remove("closed");
      toggleRightButton.setAttribute("data-state", "opened");
      toggleRightButton.setAttribute("aria-expanded", "true");
    }
  });
});
