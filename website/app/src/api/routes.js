export function getRequest(url, callback, errorCallback) {
  doRequest("GET", url, null, null, callback, errorCallback);
}

export function doRequest(method, url, data, files, callback, errorCallback) {
  var headers = {};

  if (files && files.length > 0) {
    var body = new FormData();
    body.append("file", files[0]);
    body.append("json", data ? JSON.stringify(data) : null);
  } else {
    headers["Content-Type"] = "application/json";
    var body = data ? JSON.stringify(data) : null;
  }

  fetch(url, {
    method: method,
    headers: headers,
    body: body,
  })
    .then((response) => {
      callback(response);
    })
    .catch((error) => {
      errorCallback(error);
    });
}

export function doLogin() {
  let url = new URL(window.location.href);
  url.pathname = "/login";
  url.searchParams.append("path", window.location.pathname);

  window.location.href = url;
}
