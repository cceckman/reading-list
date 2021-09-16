/* Helpers to update visibility on load */
const INVISIBLE_CLASS = "invisible";
function makeVisible(elem) {
  if (elem.classList.contains(INVISIBLE_CLASS)) {
    elem.className = Array.from(elem.classList).filter((cl) => cl != INVISIBLE_CLASS).join(" ");
  }
}
function makeHidden(elem) {
  if (!elem.classList.contains(INVISIBLE_CLASS)) {
    let newClasses = elem.classList;
    newClasses.add(INVISIBLE_CLASS);
    elem.className = newClasses;
  }
}

/* Handling of events from the service worker */
const SERVER_UPDATE = 'SERVER_UPDATE'; // {type: SERVER_UPDATE, server: "123.456.789.10:8080" }

function onActive(registration) {
  console.log("on active:", registration)
  let worker = registration.active;

  // Request an update of the current server.
  worker.postMessage({ type: SERVER_UPDATE });
  let registeredMessage = document.getElementById("loading-message");
  let content = document.getElementById("app-content");
  makeHidden(registeredMessage);
  makeVisible(content);
}

function onWorkerMessage(event) {
  console.log("got worker message:", event);
  if (event.data && event.data.type == SERVER_UPDATE) {
    console.log("got updated server:", event.data.server);
    let current = document.getElementById("current-server");
    current.innerText = event.data.server;
  }
}

function onSubmitServer(event) {
  navigator.serviceWorker.ready.then((registration) => {
    const worker = registration.active;
    const server = document.getElementById("server").value;
    console.log("sending set-server: ", server);
    worker.postMessage({ type: SERVER_UPDATE, server: server });
  });
  // Don't do an HTTP submit.
  event.preventDefault();
  return false;
}

const ADD_FORM = "add-form";

// Handle "Add" events via JS, rather than changing the window.
function onSubmitAdd(event) {
  const form = document.getElementById(ADD_FORM);
  let data = new FormData(form);
  const url = data.get("url");
  let status = document.getElementById("submit-progress");
  status.innerText = `Submitting ${url}...`
  makeVisible(status);

  fetch("/entries", {
    method: "POST",
    body: data,
  }).then((response) => {
    console.log("response from submit:", response)
    if(response.ok) {
      status.innerText = `Successfully submitted ${url}`;
    } else {
      throw `${response.status} ${response.statusText}`;
    }
  }).catch((e) => {
    status.innerText = `Error submitting ${url}: ${e}`;
  });
  event.preventDefault();
  return false;
}

/* Set up the service worker */
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('worker.js')
    .then((reg) => {
      console.log('Registration succeeded; got: ', reg.scope, ",");
      navigator.serviceWorker.onmessage = onWorkerMessage;
      document.getElementById("server-control").onsubmit = onSubmitServer;
      document.getElementById(ADD_FORM).onsubmit = onSubmitAdd;
      return navigator.serviceWorker.ready;
    }).then(onActive).catch((error) => {
      console.log('Registration failed: ', error);
    });
}

/*
  TODO: Use https://web.dev/get-installed-related-apps/ to suppress advertisement of installability.
  Requires serving manifest.json from a hostname-aware server, not just static-serving.
*/