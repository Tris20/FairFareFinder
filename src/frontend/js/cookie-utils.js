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
