// Package webui embeds the compiled frontend assets from web/dist.
// This package exists solely to hold the //go:embed directive; the embedded
// path (dist/) is a subdirectory of this file's location (web/).
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// FS returns an fs.FS rooted at the web/dist directory.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// This can only fail if the embed directive didn't compile, which is
		// caught at build time.
		panic("webui: sub fs: " + err.Error())
	}
	return sub
}
