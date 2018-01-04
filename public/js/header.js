(function() {
	document.addEventListener("DOMContentLoaded", () => {
		let logoutButton = document.querySelector("#logoutButton");
		if (logoutButton) { //Check needed for sites without a logged in user
			logoutButton.onclick = onLogoutClick;
		}
	});

	function onLogoutClick() {
		fetch("/logout", {
			credentials: "include",
			method: "POST",
		}).then((res) => {
			if (res.status !== 200) {
				UIkit.notification({ message: "Logout failed with status code " + res.status, status: "danger" })
				return;
			}
			window.location.href = "/login";
		}).catch((err) => {
			UIkit.notification({ message: "Logout failed because of " + err, status: "danger" })
		});

		// This is done to prevent a reload
		return false;
	}
})()