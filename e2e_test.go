package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/residwi/go-project-generator/internal/assets"
	"github.com/residwi/go-project-generator/internal/fetch"
	"github.com/residwi/go-project-generator/internal/scaffold"
)

// TestEndToEnd fetches the real template, generates a project (default skeleton
// and the --worker variant), and runs `go build ./...` + `gofmt -l` against each.
// Requires network + Go toolchain. Skipped under -short.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network/build e2e test in -short mode")
	}

	src := &fetch.GitHub{Repo: "residwi/go-api-project-template"}
	fsys, cleanup, err := src.Fetch(t.Context(), "main")
	require.NoError(t, err, "fetch")
	defer cleanup()

	t.Run("default skeleton", func(t *testing.T) {
		files, err := scaffold.Generate(fsys, scaffold.Options{
			Module:      "github.com/example/e2eapp",
			ProjectName: "e2eapp",
		}, assets.Overrides())
		require.NoError(t, err, "generate")

		out := filepath.Join(t.TempDir(), "e2eapp")
		require.NoError(t, scaffold.Write(out, files, false), "write")

		assertExists(t, out, "internal/features/auth/service.go")
		assertExists(t, out, "internal/platform/jobs/runner.go")
		assertMissing(t, out, "internal/features/cart")
		assertMissing(t, out, "internal/wiring")
		assertMissing(t, out, "internal/core/money.go")
		assertMissing(t, out, "internal/platform/email")
		assertMissing(t, out, "cmd/worker")

		goBuild(t, out)
		goVet(t, out)
		gofmtClean(t, out)
	})

	t.Run("with worker", func(t *testing.T) {
		files, err := scaffold.Generate(fsys, scaffold.Options{
			Module:      "github.com/example/e2eworker",
			ProjectName: "e2eworker",
			Worker:      true,
		}, assets.Overrides())
		require.NoError(t, err, "generate")

		out := filepath.Join(t.TempDir(), "e2eworker")
		require.NoError(t, scaffold.Write(out, files, false), "write")

		assertExists(t, out, "cmd/worker/main.go")
		assertExists(t, out, ".air.worker.toml")
		assertExists(t, out, "internal/platform/jobs/runner.go")

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
