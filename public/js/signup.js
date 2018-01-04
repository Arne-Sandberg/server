(function () {
  let fnameField;
  let lnameField;
  let emailField;
  let passwordField;
  let passwordConfirmField;
  let passwordConfirmHelp;
  let submitButton;

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("#signupForm").onsubmit = onSignupSubmit;
    fnameField = document.querySelector("input[name='fname']");
    lnameField = document.querySelector("input[name='lname']");
    emailField = document.querySelector("input[name='email']");
    passwordField = document.querySelector("input[name='password']");
    passwordFieldHelp = document.querySelector(".help.password-help");
    passwordConfirmField = document.querySelector("input[name='password-confirm']");
    passwordConfirmHelp = document.querySelector(".help.password-confirm-help");
    submitButton = document.querySelector("button[type='submit']")

    fnameField.addEventListener("input", (event) => {
      if (event.target.value.length > 0) {
        event.target.classList.add("uk-form-success");
      } else {
        event.target.classList.remove("uk-form-success");
      }

      checkButtonActivation();
    });

    lnameField.addEventListener("input", (event) => {
      if (event.target.value.length > 0) {
        event.target.classList.add("uk-form-success");
      } else {
        event.target.classList.remove("uk-form-success");
      }

      checkButtonActivation();
    });

    emailField.addEventListener("input", (event) => {
      if (event.target.value.length == 0) {
        event.target.classList.remove("uk-form-success");
        event.target.classList.remove("uk-form-danger");
      } else if (validateEmail(event.target.value)) {
        event.target.classList.remove("uk-form-danger");
        event.target.classList.add("uk-form-success");
      } else {
        event.target.classList.remove("uk-form-success");
        event.target.classList.add("uk-form-danger");
      }

      checkButtonActivation();
    });

    passwordField.addEventListener("input", (event) => {
      if (event.target.value.length < 6) {
        passwordField.classList.add("uk-form-danger");
        passwordField.classList.remove("uk-form-success");
        passwordFieldHelp.removeAttribute("hidden");
      } else {
        passwordFieldHelp.setAttribute("hidden", "")
        passwordField.classList.add("uk-form-success");
        passwordField.classList.remove("uk-form-danger");
      }

      checkButtonActivation();
    })

    passwordConfirmField.addEventListener("input", (event) => {
      if (event.target.value !== passwordField.value) {
        passwordConfirmField.classList.add("uk-form-danger");
        passwordConfirmField.classList.remove("uk-form-success");
        passwordConfirmHelp.removeAttribute("hidden");
      } else {
        passwordConfirmField.classList.remove("uk-form-danger");
        passwordConfirmField.classList.add("uk-form-success");
        passwordConfirmHelp.setAttribute("hidden", "")
      }

      checkButtonActivation();
    });
  })


  function onSignupSubmit() {
    // First of all, validate the passwords match
    if (passwordField.value !== passwordConfirmField.value) {
      passwordConfirmField.classList.add("uk-form-danger");
      passwordConfirmHelp.classList.add("uk-form-danger");
      passwordConfirmHelp.removeAttribute("hidden");
      return false;
    }

    fetch("/signup", {
      credentials: "include",
      method: "POST",
      body: JSON.stringify({
        firstName: fnameField.value,
        lastName: lnameField.value,
        email: emailField.value,
        password: passwordField.value,
      })
    }).then((res) => {
      if (res.status !== 200) {
        UIkit.notification({message: "Signup failed with status code " + res.status, status: "danger"})
        return;
      }
      window.location.href = "/";
    }).catch((err) => {
      UIkit.notification({message: "Signup failed beacuse of " + err, status: "danger"})
    });


    // This is done to prevent a reload
    return false;
  }

  function validateEmail(email) {
    return email.includes("@") && email.includes(".")
  }

  function checkButtonActivation() {
    if (fnameField.classList.contains("uk-form-success") &&
        lnameField.classList.contains("uk-form-success") &&
        emailField.classList.contains("uk-form-success") &&
        passwordField.classList.contains("uk-form-success") &&
        passwordConfirmField.classList.contains("uk-form-success")) {
      submitButton.removeAttribute("disabled")
    } else {
      submitButton.setAttribute("disabled", "")
    }
  }
})()