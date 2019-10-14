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

	const createUserButton = document.getElementById("userCreateBtn");
	createUserButton.onclick = onCreateUser;

	const deleteUserButtons = document.getElementsByClassName("userDeleteButton");
	for (const button of deleteUserButtons) {
		button.onclick = onDeleteUser;
	}
})();