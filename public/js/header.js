(function() {
	document.addEventListener("DOMContentLoaded", () => {
		document.querySelector("#logoutButton").onclick = onLogoutClick;
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