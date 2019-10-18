(() => {
  const siteID = parseInt(location.pathname.split("/").pop());

  function onEditSite(event) {
    event.preventDefault();

    const hostField = document.getElementById("siteHostField");
    const pathPrefixField = document.getElementById("sitePathPrefixField");
    const anonymousAccessField = document.getElementById("siteAnonymousAccessField");
    const body = { "id": siteID, "host": hostField.value, "path_prefix": pathPrefixField.value, "anonymous_access": anonymousAccessField.checked };
    sendAPIRequest("PUT", "/api/site", body, "Successfully changed site", "Failed to change site: ");
  }

  attachOnEnter("siteHostField", onEditSite);
  attachOnEnter("sitePathPrefixField", onEditSite);

  const editSiteButton = document.getElementById("siteEditButton");
  editSiteButton.onclick = onEditSite;

  function onCreateMapping(event) {
    event.preventDefault();

    const userField = document.getElementById("mappingUserField")
    const basicAuthField = document.getElementById("mappingBasicAuthField")
    const body = { "user_id": parseInt(userField.value), "site_id": siteID, "basic_auth_allowed": basicAuthField.checked };
    sendAPIRequest("POST", "/api/mapping", body, "Successfully added user mapping", "Failed to add user mapping: ");
  }

  attachOnEnter("mappingUserField", onCreateMapping);

  const createMappingButton = document.getElementById("mappingAddButton");
  if (createMappingButton)
    createMappingButton.onclick = onCreateMapping;
  
  function onDeleteMapping(event) {
    event.preventDefault();
    const source = event.target || event.srcElement;

    const body = { "user_id": parseInt(source.getAttribute("data-userid")), "site_id": siteID };
    sendAPIRequest("DELETE", "/api/mapping", body, "Successfully deleted user mapping", "Failed to delete user mapping: ");
  }

	const deleteMappingButtons = document.getElementsByClassName("mappingDeleteButton");
	for (const button of deleteMappingButtons) {
		button.onclick = onDeleteMapping;
  }

	document.addEventListener('DOMContentLoaded', function() {
    const elems = document.querySelectorAll('select');
    M.FormSelect.init(elems, {});
  });
})();