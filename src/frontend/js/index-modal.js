function openModal(cityName) {
  const modal = document.getElementById(`modal-${cityName}`);
  if (modal) {
    modal.style.display = "flex";
  }
}

function closeModal(destinationCity) {
  const modal = document.getElementById(`modal-${destinationCity}`);
  if (modal) {
    modal.style.display = "none";
  }
}

function closeModalOnOutsideClick(event, destinationCity) {
  const modalContent = event.currentTarget.querySelector(".modal-content");

  // Check if the click is outside the modal-content
  if (!modalContent.contains(event.target)) {
    closeModal(destinationCity); // Close the modal
  }
}
