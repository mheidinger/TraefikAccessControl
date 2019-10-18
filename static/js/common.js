const url = new URL(location.href);
url.searchParams.delete("error");
url.searchParams.delete("success");
history.replaceState(null, document.title, url.href);

function attachOnEnter(fieldID, func) {
	const field = document.getElementById(fieldID);
	if (field) {
		field.addEventListener("keyup", event => {
			if (event.keyCode === 13) {
				event.preventDefault();
				func(event);
			}
		});
	} else {
		console.warn("Element '" + fieldID + "' not found for attaching enter action");
	}
}

function sendAPIRequest(method, path, body, success, error) {
	const fetchUrl = new URL(location.href);
	fetchUrl.pathname = path;
	fetchUrl.search = ""

	fetch(fetchUrl.href, {
		method: method,
		body: JSON.stringify(body)
	}).then(response => response.json()
	).then(json => {
		if (json["error"]) {
			reloadError(error + json["error"]);
		} else {
			reloadSuccess(success);
		}
	});
}

function fetchResponseToPromise(response) {
	if (response.ok) {
		return response.json();
	} else {
		throw Error(response.json());
	}
}

function reloadSuccess(message) {
	let url = new URL(location.href);
	url.searchParams.delete("error");
	url.searchParams.delete("success");
	url.searchParams.append("success", message);
	location.replace(url.href);
}

function reloadError(message) {
	let url = new URL(location.href);
	url.searchParams.delete("error");
	url.searchParams.delete("success");
	url.searchParams.append("error", message);
	location.replace(url.href);
}