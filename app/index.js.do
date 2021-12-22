
redo-ifchange modules.dir js.dir
# Ignore worker.js; it's a separate, unbundled file (service worker entry point.)

exec >&2

npm run esbuild -- \
  --bundle \
  js/app.js \
  --outfile="$3"