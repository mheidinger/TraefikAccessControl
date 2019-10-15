(() => {
	const url = new URL(document.location.href);

	function onCreateBearer(event) {
		event.preventDefault();

		const nameField = document.getElementById("bearerNameField");
		const body = { "name": nameField.value };
		url.pathname = "/api/bearer";
		fetch(url.href, {
			method: "POST",
			body: JSON.stringify(body)
		}).then(() => location.reload());
	}

	function onDeleteBearer(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "name": source.getAttribute("data-tokenname") };
		url.pathname = "/api/bearer";
		fetch(url.href, {
      method: "DELETE",
      body: JSON.stringify(body)
    }).then(() => location.reload());
	}

	function onChangePassword(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const passwordField = document.getElementById("changePasswordField")
		const body = { "password": passwordField.value };
		url.pathname = "/api/user"
		fetch(url.href, {
      method: "PUT",
      body: JSON.stringify(body)
    }).then(() => location.reload());
	}

	const createBearerButton = document.getElementById("bearerCreateButton");
	createBearerButton.onclick = onCreateBearer;

	const deleteBearerButtons = document.getElementsByClassName("bearerDeleteButton");
	for (const button of deleteBearerButtons) {
		button.onclick = onDeleteBearer;
	}

	const changePasswordButton = document.getElementById("changePasswordButton")
	changePasswordButton.onclick = onChangePassword;
})();