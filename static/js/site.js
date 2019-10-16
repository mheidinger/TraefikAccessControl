(() => {
  const url = new URL(location.href);
  const siteID = parseInt(location.pathname.split("/").pop());

  function onEditSite(event) {
    event.preventDefault();

    const hostField = document.getElementById("siteHostField")
    const pathPrefixField = document.getElementById("sitePathPrefixField")
    const body = { "id": siteID, "host": hostField.value, "path_prefix": pathPrefixField.value };
    sendAPIRequest("PUT", "/api/site", body, "Successfully changed site", "Failed to change site: ");
  }

  function onCreateMapping(event) {
    event.preventDefault();

    const userField = document.getElementById("mappingUserField")
    const basicAuthField = document.getElementById("mappingBasicAuthField")
    const body = { "user_id": parseInt(userField.value), "site_id": siteID, "basic_auth_allowed": basicAuthField.checked };
    sendAPIRequest("POST", "/api/mapping", body, "Successfully added user mapping", "Failed to add user mapping: ");
  }

  function onDeleteMapping(event) {
    event.preventDefault();
    const source = event.target || event.srcElement;

    const body = { "user_id": parseInt(source.getAttribute("data-userid")), "site_id": siteID };
    sendAPIRequest("DELETE", "/api/mapping", body, "Successfully deleted user mapping", "Failed to delete user mapping: ");
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