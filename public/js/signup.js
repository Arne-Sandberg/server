(function () {
  let fnameField;
  let lnameField;
  let emailField;
  let passwordField;
  let passwordConfirmField;
  let passwordConfirmHelp;

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("form.signup-form").onsubmit = onSignupSubmit;
    fnameField = document.querySelector("input[name='fname']");
    lnameField = document.querySelector("input[name='lname']");
    emailField = document.querySelector("input[name='email']");
    passwordField = document.querySelector("input[name='password']");
    passwordFieldHelp = document.querySelector(".help.password-help");
    passwordConfirmField = document.querySelector("input[name='password-confirm']");
    passwordConfirmHelp = document.querySelector(".help.password-confirm-help");

    fnameField.addEventListener("input", (event) => {
      if (event.target.value.length > 0) {
        event.target.classList.add("is-success");
      }
    });

    lnameField.addEventListener("input", (event) => {
      if (event.target.value.length > 0) {
        event.target.classList.add("is-success");
      }
    });

    emailField.addEventListener("input", (event) => {
      console.log(event.target.checkValidity())
      if (event.target.checkValidity()) {
        event.target.classList.add("is-success");
      }
    });

    passwordField.addEventListener("input", (event) => {
      if (event.target.value.length < 8) {
        passwordField.classList.add("is-danger");
        passwordField.classList.remove("is-success");
        passwordFieldHelp.classList.remove("is-invisible");
      } else {
        passwordFieldHelp.classList.add("is-invisible");
        passwordField.classList.add("is-success");
        passwordField.classList.remove("is-danger");
      }
    })

    passwordConfirmField.addEventListener("input", (event) => {
      if (event.target.value !== passwordField.value) {
        passwordConfirmField.classList.add("is-danger");
        passwordConfirmField.classList.remove("is-success");
        passwordConfirmHelp.classList.remove("is-invisible");
      } else {
        passwordConfirmField.classList.remove("is-danger");
        passwordConfirmField.classList.add("is-success");
        passwordConfirmHelp.classList.add("is-invisible");
      }
    });
  })


  function onSignupSubmit() {
    // First of all, validate the passwords match
    // TODO: also validate the passwords on input
    if (passwordField.value !== passwordConfirmField.value) {
      passwordConfirmField.classList.add("is-danger");
      passwordConfirmHelp.classList.add("is-danger");
      passwordConfirmHelp.classList.remove("is-invisible");
      return false;
    }

    fetch("/signup", {
      method: "POST",
      body: JSON.stringify({
        firstName: fnameField.value,
        lastName: lnameField.value,
        email: emailField.value,
        password: passwordField.value,
      })
    }).then((res) => {
      if (res.status !== 200) {
        alert("Signup failed with status code " + res.status);
        return;
      }
      window.location.href = "/";
    }).catch((err) => {
      alert("Signup failed, because " + err);
    });


    // This is done to prevent a reload
    return false;
  }
})()