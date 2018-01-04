(function() {
	let progressBar;

	document.addEventListener("DOMContentLoaded", () => {
		progressBar = document.getElementById("js-progressbar");

		UIkit.upload(".js-upload", {
			url: "/upload",
			multiple: true,
			type: "POST",
			name: "files",

			error: function() {
				UIkit.notification({ message: "An error occured during the file upload!", status: "danger" });
			},

			loadStart: function(e) {
				progressBar.max = e.total;
				progressBar.value = e.loaded;
			},

			progress: function(e) {
				progressBar.max = e.total;
				progressBar.value = e.loaded;
			},

			loadEnd: function(e) {
				progressBar.max = e.total;
				progressBar.value = e.loaded;
			},

			completeAll: function() {
				// Reload the page so that the uploaded file is shown
				// This should be replaced by WebSockets or similar so that a reload is not needed
				UIkit.notification({ message: "File(s) successfully uploaded.<br />This page will be reloaded shortly!", status: "success" });
				setTimeout(function() { location.reload(); }, 1500);
			}
		});


	});
})();