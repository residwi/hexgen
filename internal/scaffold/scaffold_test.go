package scaffold

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_RendersTemplatesAndRewritesTokens(t *testing.T) {
	skeleton := fstest.MapFS{
		"go.mod.tmpl":               {Data: []byte("module {{.Module}}\n")},
		"internal/server/server.go": {Data: []byte("package server // " + ModulePlaceholder + "/internal/app\n")},
		"README.md":                 {Data: []byte("# __PROJECT_NAME__")},
	}

	out, err := Generate(Options{Module: "github.com/me/myapp", ProjectName: "myapp"}, skeleton)
	require.NoError(t, err)

	assert.Equal(t, "module github.com/me/myapp\n", string(out["go.mod"].Data))
	assert.NotContains(t, out, "go.mod.tmpl")
	assert.Equal(t, "package server // github.com/me/myapp/internal/app\n", string(out["internal/server/server.go"].Data))
	assert.Equal(t, "# myapp", string(out["README.md"].Data))
}

func TestGenerate_OmitsTemplateThatRendersToWhitespace(t *testing.T) {
	skeleton := fstest.MapFS{
		"gated.go.tmpl": {Data: []byte("{{if eq .ProjectName \"other\"}}package gated{{end}}\n")},
	}

	out, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton)
	require.NoError(t, err)

	assert.NotContains(t, out, "gated.go")
	assert.NotContains(t, out, "gated.go.tmpl")
}

func TestGenerate_GatesOnAuth(t *testing.T) {
	skeleton := fstest.MapFS{
		"router.go.tmpl": {Data: []byte("package server\n{{if .Auth}}// auth routes\n{{end}}")},
	}

	off, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton)
	require.NoError(t, err)
	assert.NotContains(t, string(off["router.go"].Data), "auth routes")

	on, err := Generate(Options{Module: "m", ProjectName: "p", Auth: true}, skeleton)
	require.NoError(t, err)
	assert.Contains(t, string(on["router.go"].Data), "auth routes")
}

func TestGenerate_LaterLayerWins(t *testing.T) {
	skeleton := fstest.MapFS{
		"shared.txt": {Data: []byte("skeleton")},
		"base.txt":   {Data: []byte("base")},
	}
	overlay := fstest.MapFS{
		"shared.txt": {Data: []byte("overlay")},
		"extra.txt":  {Data: []byte("extra")},
	}

	out, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton, overlay)
	require.NoError(t, err)

	assert.Equal(t, "overlay", string(out["shared.txt"].Data))
	assert.Equal(t, "base", string(out["base.txt"].Data))
	assert.Equal(t, "extra", string(out["extra.txt"].Data))
}

func TestGenerate_FormatsGoFiles(t *testing.T) {
	skeleton := fstest.MapFS{"x.go": {Data: []byte("package x\nfunc  f() {}\n")}}

	out, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton)
	require.NoError(t, err)

	assert.Equal(t, "package x\n\nfunc f() {}\n", string(out["x.go"].Data))
}

func TestGenerate_NormalizesModes(t *testing.T) {
	skeleton := fstest.MapFS{
		"script.sh": {Data: []byte("#!/bin/sh\n"), Mode: 0o700},
		"notes.txt": {Data: []byte("notes"), Mode: 0o600},
	}

	out, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton)
	require.NoError(t, err)

	assert.Equal(t, fs.FileMode(0o755), out["script.sh"].Mode)
	assert.Equal(t, fs.FileMode(0o644), out["notes.txt"].Mode)
}

func TestGenerate_RejectsInvalidTemplate(t *testing.T) {
	skeleton := fstest.MapFS{"broken.txt.tmpl": {Data: []byte("{{.Module")}}

	_, err := Generate(Options{Module: "m", ProjectName: "p"}, skeleton)

	assert.ErrorContains(t, err, "broken.txt.tmpl")
}
