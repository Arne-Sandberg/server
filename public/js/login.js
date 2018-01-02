(function () {
  let fnameField;
  let lnameField;
  let emailField;
  let passwordField;
  let passwordConfirmField;
  let passwordConfirmHelp;

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("form.login-form").onsubmit = onLoginSubmit;
    emailField = document.querySelector("input[name='email']");
    passwordField = document.querySelector("input[name='password']");
    passwordFieldHelp = document.querySelector(".help.password-help");

    emailField.addEventListener("input", (event) => {
      console.log(event.target.checkValidity())
      if (event.target.checkValidity()) {
        event.target.classList.add("is-success");
      }
    });

  });


  function onLoginSubmit() {
    // First of all, validate the passwords match
    // TODO: also validate the passwords on input
   
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