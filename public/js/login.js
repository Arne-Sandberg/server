(function () {
  let emailField;
  let passwordField;

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("#loginForm").onsubmit = onLoginSubmit;
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
      if (res.status == 401) {
        UIkit.notification({message: "Wrong email address or password!", status: "danger"})
        return
      } else if (res.status !== 200) {
        UIkit.notification({message: "Login failed with status code " + res.status, status: "danger"})
        return;
      }
      window.location.href = "/";
    }).catch((err) => {
      UIkit.notification({message: "Login failed because of " + err, status: "danger"})
    });

    // This is done to prevent a reload
    return false;
  }
})()
