(function () {
  let emailField;
  let passwordField;

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("form").onsubmit = onLoginSubmit;
    emailField = document.querySelector("input[name='email']");
    passwordField = document.querySelector("input[name='password']");
  });

  function onLoginSubmit() {
    fetch("/login", {
      credentials: "include",
      method: "POST",
      body: JSON.stringify({
        email: emailField.value,
        password: passwordField.value,
      })
    }).then((res) => {
      if (res.status !== 200) {
        alert("Login failed with status code " + res.status);
        return;
      }
      window.location.href = "/";
    }).catch((err) => {
      alert("Login failed, because " + err);
    });

    // This is done to prevent a reload
    return false;
  }
})()
