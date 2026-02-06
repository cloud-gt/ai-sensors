package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS returns the filesystem containing the built SPA files.
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
