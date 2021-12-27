
redo-ifchange modules.dir *.ts
# Ignore worker.js; it's a separate, unbundled file
# (the service worker entry point.)

exec >&2

# Type-check with TSC;
# esbuild just compiles and bundles.
npm run tsc -- --noemit

# This produces a .js.map file as a side-effect as well.
# We want to capture that separately, so use an output dir.
npm run esbuild -- \
  --bundle \
  --sourcemap \
  --minify \
  app.ts \
  --outdir=js/

cp js/app.js "$3"