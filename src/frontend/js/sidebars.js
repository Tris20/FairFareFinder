// Wait for the DOM to load
document.addEventListener("DOMContentLoaded", () => {
  const toggleLeftButton = document.getElementById("toggleLeftSidebar");
  const toggleRightButton = document.getElementById("toggleRightSidebar");
  const leftSidebar = document.getElementById("left-sidebar");
  const rightSidebar = document.getElementById("right-sidebar");

  // Toggle left sidebar
  toggleLeftButton.addEventListener("click", () => {
    leftSidebar.classList.toggle("closed");
  });

  // Toggle right sidebar
  toggleRightButton.addEventListener("click", () => {
    rightSidebar.classList.toggle("closed");
  });
});
