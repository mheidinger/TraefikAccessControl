(() => {
	function onCreateBearer(event) {
		event.preventDefault();

		const nameField = document.getElementById("bearerNameField");
		const body = { "name": nameField.value };
		sendAPIRequest("POST", "/api/bearer", body, "Successfully created bearer", "Failed to create bearer: ");
	}

	attachOnEnter("bearerNameField", onCreateBearer);

	const createBearerButton = document.getElementById("bearerCreateButton");
	createBearerButton.onclick = onCreateBearer;

	function onDeleteBearer(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "name": source.getAttribute("data-tokenname") };
		sendAPIRequest("DELETE", "/api/bearer", body, "Successfully deleted bearer", "Failed to delete bearer: ");
	}

	const deleteBearerButtons = document.getElementsByClassName("bearerDeleteButton");
	for (const button of deleteBearerButtons) {
		button.onclick = onDeleteBearer;
	}

	function onChangePassword(event) {
		event.preventDefault();

		const passwordField = document.getElementById("changePasswordField");
		const body = { "password": passwordField.value };
		sendAPIRequest("PUT", "/api/user", body, "Successfully changed password", "Failed to change password: ");
	}

	attachOnEnter("changePasswordField", onChangePassword);

	const changePasswordButton = document.getElementById("changePasswordButton");
	changePasswordButton.onclick = onChangePassword;
})();