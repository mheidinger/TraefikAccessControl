(() => {
  const url = new URL(document.location.href);
  const siteID = parseInt(document.location.pathname.split("/").pop());

  function onEditSite(event) {
    event.preventDefault();

    const hostField = document.getElementById("siteHostField")
    const pathPrefixField = document.getElementById("sitePathPrefixField")
    const body = { "id": siteID, "host": hostField.value, "path_prefix": pathPrefixField.value };
    url.pathname = "/api/site";
		fetch(url.href, {
			method: "PUT",
			body: JSON.stringify(body)
		}).then(() => location.reload());
  }

  function onCreateMapping(event) {
    event.preventDefault();

    const userField = document.getElementById("mappingUserField")
    const basicAuthField = document.getElementById("mappingBasicAuthField")
    const body = { "user_id": parseInt(userField.value), "site_id": siteID, "basic_auth_allowed": basicAuthField.checked };
		url.pathname = "/api/mapping";
		fetch(url.href, {
			method: "POST",
			body: JSON.stringify(body)
		}).then(() => location.reload());
  }

  function onDeleteMapping(event) {
    event.preventDefault();
    const source = event.target || event.srcElement;

		const body = { "user_id": parseInt(source.getAttribute("data-userid")), "site_id": siteID };
		url.pathname = "/api/mapping";
		fetch(url.href, {
      method: "DELETE",
      body: JSON.stringify(body)
    }).then(() => location.reload());
  }

	document.addEventListener('DOMContentLoaded', function() {
    const elems = document.querySelectorAll('select');
    M.FormSelect.init(elems, {});
  });

  const editSiteButton = document.getElementById("siteEditButton");
  editSiteButton.onclick = onEditSite;

	const createMappingButton = document.getElementById("mappingAddButton");
	createMappingButton.onclick = onCreateMapping;

	const deleteMappingButtons = document.getElementsByClassName("mappingDeleteButton");
	for (const button of deleteMappingButtons) {
		button.onclick = onDeleteMapping;
  }
})();