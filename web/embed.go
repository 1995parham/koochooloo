// Package web embeds the built admin-panel single-page application so the
// server ships as a single self-contained binary. Rebuild the assets with
// `just web` (or `pnpm --dir web build`) after changing anything under src/.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built SPA assets rooted at the dist directory. The embed is
// resolved at build time, so a malformed layout is a compile/startup error.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}

	return sub
}
