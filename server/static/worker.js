// Meet the letter but not the spirit of "installable":
// register a fetch handler that "just" forwards to the network.
self.addEventListener('fetch', (event) => {
  console.log('Service Worker Fetch', event.request.url);
  event.respondWith(fetch(event.request));
});