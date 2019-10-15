(() => {
	const url = new URL(document.location.href);

	function onCreateUser(event) {
		event.preventDefault();

		const nameField = document.getElementById("userNameField")
		const passwordField = document.getElementById("userPasswordField")
		const adminField = document.getElementById("userAdminField")
		const body = { "username": nameField.value, "password": passwordField.value, "is_admin": adminField.checked }
		url.pathname = "/api/user"
		fetch(url.href, {
			method: "POST",
			body: JSON.stringify(body)
		}).then(() => location.reload())
	}

	function onDeleteUser(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "id": parseInt(source.getAttribute("data-userid")) }
		url.pathname = "/api/user";
		fetch(url.href, {
      method: "DELETE",
      body: JSON.stringify(body)
    }).then(() => location.reload())
	}

	function onCreateSite(event) {
		event.preventDefault();
		const hostField = document.getElementById("siteHostField")
		const pathPrefixField = document.getElementById("sitePathPrefixField")
		const body = { "host": hostField.value, "path_prefix": pathPrefixField.value }
		url.pathname = "/api/site"
		fetch(url.href, {
			method: "POST",
			body: JSON.stringify(body)
		}).then(() => location.reload())
	}

	function onDeleteSite(event) {
		event.preventDefault();
		const source = event.target || event.srcElement;

		const body = { "id": parseInt(source.getAttribute("data-siteid")) }
		url.pathname = "/api/site";
		fetch(url.href, {
      method: "DELETE",
      body: JSON.stringify(body)
    }).then(() => location.reload())
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