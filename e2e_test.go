package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/residwi/go-project-generator/internal/scaffold"
	"github.com/residwi/go-project-generator/internal/template"
)

// TestEndToEnd generates a project from the embedded skeleton and runs
// `go build ./...`, `go vet ./...` and `gofmt -l` against it. Requires the Go
// toolchain and GOPROXY access for the generated project's dependencies.
// Skipped under -short.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build e2e test in -short mode")
	}

	t.Run("platform only", func(t *testing.T) {
		files, err := scaffold.Generate(scaffold.Options{
			Module:      "github.com/example/e2eapp",
			ProjectName: "e2eapp",
		}, template.Skeleton())
		require.NoError(t, err, "generate")

		out := filepath.Join(t.TempDir(), "e2eapp")
		require.NoError(t, scaffold.Write(out, files, false), "write")

		assertExists(t, out, "go.mod")
		assertExists(t, out, "internal/platform/web/router.go")
		assertMissing(t, out, "go.mod.tmpl")
		assertMissing(t, out, "internal/features")
		assertMissing(t, out, "cmd/worker")

		goBuild(t, out)
		goVet(t, out)
		gofmtClean(t, out)
	})
}

func assertExists(t *testing.T, root, rel string) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, rel))
	assert.NoErrorf(t, err, "%s should exist", rel)
}

func assertMissing(t *testing.T, root, rel string) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, rel))
	assert.Truef(t, os.IsNotExist(err), "%s should have been dropped", rel)
}

func goBuild(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "generated project failed to build:\n%s", out)
}

// goVet type-checks the generated project including its test files (which
// `go build` skips), catching e.g. a kept test referencing a trimmed symbol.
func goVet(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "go vet failed in generated project:\n%s", out)
}

func gofmtClean(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "gofmt")
	assert.Emptyf(t, strings.TrimSpace(string(out)), "gofmt-dirty files in generated project:\n%s", out)
}
