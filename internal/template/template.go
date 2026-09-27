// Package template embeds the project tree that `hexgen new` writes.
//
// The tree lives under testdata/ so the hexgen build never compiles it: its Go
// files import the placeholder module path that scaffold rewrites. go.mod is
// stored as go.mod.tmpl because a real go.mod would make the directory a
// separate module, which go:embed refuses to embed.
package template

import (
	"embed"
	"io/fs"
)

//go:embed all:testdata
var files embed.FS

// Skeleton returns the platform-only project tree.
func Skeleton() fs.FS { return sub("testdata/skeleton") }

// Auth returns the overlay written over Skeleton with --auth.
func Auth() fs.FS { return sub("testdata/auth") }

func sub(dir string) fs.FS {
	f, err := fs.Sub(files, dir)
	if err != nil {
		panic(err) // dir is a constant path inside the embedded tree
	}
	return f
}
