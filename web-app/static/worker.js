// TODO: This gets reset with every reload, i.e. when the worker file changes content.
// We can store this in one of the web-app storage areas and retrieve it.
var SERVER = "localhost:8081";

// Message handling:
// {type: SERVER_UPDATE, server?: "server"}
// If 'server' field is not present, don't change; just query the current value
// Generates a SERVER_UPDATE message back
const SERVER_UPDATE = 'SERVER_UPDATE';

self.addEventListener('message', (event) => {
  console.log("got message: ", event)
  if (event.data && event.data.type == SERVER_UPDATE) {
    if ('server' in event.data) {
      SERVER = event.data.server;
    }
    const data = {
      type: SERVER_UPDATE,
      server: SERVER,
    };
    self.clients.matchAll().then((clients) => {
      for (const client of clients) {
        console.log("broadcasting new server: ", data);
        client.postMessage(data);
      }
    });
  }
});

// POST /entries: Proxy to the same path on the current "server".
// All other requests: Proxy without translation.
const ENTRIES_PATH = "/entries";
self.addEventListener('fetch', (event) => {
  const req = event.request;
  console.log('Service Worker Fetch', req);
  if (req.method == "POST" && req.url.endsWith(ENTRIES_PATH)) {
    let url = new URL(req.url);
    url.host = SERVER;
    event.respondWith(req.formData().then((data) => {
      console.log("forwarding request with body: ", data);
      return fetch(new Request(url, {
          method: req.method,
          body: data,
          credentials: "omit",
      }));
    }));
  } else {
    event.respondWith(fetch(req));
  }
});