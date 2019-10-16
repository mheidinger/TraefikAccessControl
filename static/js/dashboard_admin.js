(() => {
	const url = new URL(location.href);

	function onCreateUser(event) {
		event.preventDefault();

		const nameField = document.getElementById("userNameField");
		const passwordField = document.getElementById("userPasswordField");
		const adminField = document.getElementById("userAdminField");
		const body = { "username": nameField.value, "password": passwordField.value, "is_admin": adminField.checked };
		sendAPIRequest("POST", "/api/user", body, "Successfully created new user", "Failed to create new user: ");
	}

	function onDeleteUser(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "id": parseInt(source.getAttribute("data-userid")) };
		sendAPIRequest("DELETE", "/api/user", body, "Sucessfully deleted user", "Failed to delete user: ");
	}

	function onCreateSite(event) {
		event.preventDefault();

		const hostField = document.getElementById("siteHostField");
		const pathPrefixField = document.getElementById("sitePathPrefixField");
		const body = { "host": hostField.value, "path_prefix": pathPrefixField.value };
		sendAPIRequest("POST", "/api/site", body, "Successfully created new site", "Failed to create new site: ");
	}

	function onDeleteSite(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "id": parseInt(source.getAttribute("data-siteid")) };
		sendAPIRequest("DELETE", "/api/site", body, "Successfully deleted site", "Failed to delete site: ");
	}

	const createUserButton = document.getElementById("userCreateButton");
	createUserButton.onclick = onCreateUser;

	const deleteUserButtons = document.getElementsByClassName("userDeleteButton");
	for (const button of deleteUserButtons) {
		button.onclick = onDeleteUser;
	}
	
	const createSiteButton = document.getElementById("siteCreateButton");
	createSiteButton.onclick = onCreateSite;

	const deleteSiteButtons = document.getElementsByClassName("siteDeleteButton");
	for (const button of deleteSiteButtons) {
		button.onclick = onDeleteSite;
	}
})();