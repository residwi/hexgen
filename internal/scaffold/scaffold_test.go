package scaffold

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	src := fstest.MapFS{
		"go.mod":                            {Data: []byte("module " + TemplateModule + "\n")},
		"internal/features/auth/service.go": {Data: []byte("package auth // " + TemplateModule)},
		"internal/features/cart/service.go": {Data: []byte("package cart")}, // dropped
		"internal/server/router.go":         {Data: []byte("package server // template original")},
		"bin/run_test":                      {Data: []byte("binary")}, // dropped
	}
	overrides := fstest.MapFS{
		"internal/server/router.go.tmpl": {Data: []byte("package server // OVERRIDE " + TemplateModule)},
		"README.md":                      {Data: []byte("# __PROJECT_NAME__")},
	}

	out, err := Generate(src, Options{Module: "github.com/me/myapp", ProjectName: "myapp"}, overrides)
	require.NoError(t, err)

	// dropped paths absent
	assert.NotContains(t, out, "internal/features/cart/service.go")
	assert.NotContains(t, out, "bin/run_test")
	// kept + rewritten
	assert.Equal(t, "module github.com/me/myapp\n", string(out["go.mod"].Data))
	// .go files are re-formatted (go/format), so assert on content rather than exact bytes.
	assert.Contains(t, string(out["internal/features/auth/service.go"].Data), "package auth // github.com/me/myapp")
	// override replaces source, .tmpl stripped, rewritten
	assert.Contains(t, string(out["internal/server/router.go"].Data), "package server // OVERRIDE github.com/me/myapp")
	assert.NotContains(t, out, "internal/server/router.go.tmpl")
	// project-name token replaced
	assert.Equal(t, "# myapp", string(out["README.md"].Data))
}

func TestGenerate_WorkerTemplating(t *testing.T) {
	src := fstest.MapFS{"go.mod": {Data: []byte("module x")}}
	overrides := fstest.MapFS{
		"cmd/worker/main.go.tmpl":        {Data: []byte("{{if .Worker}}package main\n{{- end}}\n")},
		"internal/config/config.go.tmpl": {Data: []byte("package config\n{{if .Worker}}// worker{{end}}\n")},
	}

	off, err := Generate(src, Options{Module: "m", ProjectName: "p"}, overrides)
	require.NoError(t, err)
	assert.NotContains(t, off, "cmd/worker/main.go", "worker-only file omitted without --worker")
	assert.NotContains(t, string(off["internal/config/config.go"].Data), "// worker")

	on, err := Generate(src, Options{Module: "m", ProjectName: "p", Worker: true}, overrides)
	require.NoError(t, err)
	assert.Contains(t, on, "cmd/worker/main.go", "worker-only file present with --worker")
	assert.Contains(t, string(on["internal/config/config.go"].Data), "// worker")
}
