(function() {
	let createDirectoryNameField;

	
	document.addEventListener("DOMContentLoaded", () => {
		let logoutButton = document.querySelector("#logoutButton");
		if (logoutButton) { //Check needed for sites without a logged in user
			logoutButton.onclick = onLogoutClick;
		}
		let createDirectoryButton = document.querySelector("#createDirectoryButton");
		if (createDirectoryButton) {
			createDirectoryButton.onclick = onCreateDirectory;
		}
		createDirectoryNameField = document.querySelector("#createDirectoryNameField");
		let x = $("header");
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

	function onCreateDirectory() {
		let dirName = createDirectoryNameField.value;
		// Prepend the current path, if we aren't in the root directory
		let fullPath = (window.location.href.includes("/d/")) ? window.location.href.substring(window.location.href.indexOf("/d/") + 2) + "/" + dirName : dirName;
		console.info(`Creating directory: ${fullPath}`);
	}
})()