// Package assets embeds the override files written over the fetched template.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:files
var embedded embed.FS

// Overrides returns the override tree rooted at the files/ directory. Files ending
// in ".tmpl" are rendered with text/template by the scaffold engine (so worker-only
// content can be gated behind `{{if .Worker}}`); the suffix is then stripped.
func Overrides() fs.FS {
	sub, err := fs.Sub(embedded, "files")
	if err != nil {
		panic(err)
	}
	return sub
}
