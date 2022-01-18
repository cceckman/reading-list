// Includes the static-content portions of the webserver,
// i.e. the app.
package static

import (
	"embed"
)

// We do *not* include the .js.map file in the embedded contents; we only serve it when serving
// from the local filesystem.
//go:embed *.js *.png *.json *.css
var Files embed.FS

// TODO:
// The "related applications" feature, which you use to detect if your own web app is installed,
// requires the full manifest URL in the JSON.
// So, to use that, we have to patch manifest.json with the actual URL.
