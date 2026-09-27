package main

import (
	"io/fs"
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

// goArchLint is run with `go run` so neither CI nor a local machine needs it
// installed.
const goArchLint = "github.com/fe3dback/go-arch-lint@v1.19.0"

// TestEndToEnd generates each project variant from the embedded skeleton and
// checks it: no leftover template names or dropped paths, then `go mod tidy`,
// `go build`, `go vet`, `gofmt -l` and go-arch-lint, and `go test` when Docker
// is available. Requires the Go toolchain and GOPROXY access for the generated
// project's dependencies. Skipped under -short.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build e2e test in -short mode")
	}

	t.Run("platform only", func(t *testing.T) {
		out := generate(t, scaffold.Options{
			Module:      "github.com/example/e2eapp",
			ProjectName: "e2eapp",
		}, template.Skeleton())

		assertMissing(t, out, "internal/features")
		checkProject(t, out)
	})
}

// generate writes the project for opts and layers into a temp dir and returns
// its path.
func generate(t *testing.T, opts scaffold.Options, layers ...fs.FS) string {
	t.Helper()
	files, err := scaffold.Generate(opts, layers...)
	require.NoError(t, err, "generate")

	out := filepath.Join(t.TempDir(), opts.ProjectName)
	require.NoError(t, scaffold.Write(out, files, false), "write")
	return out
}

// checkProject runs the checks every generated variant must pass.
func checkProject(t *testing.T, dir string) {
	t.Helper()
	assertExists(t, dir, "go.mod")
	for _, p := range []string{
		"go.mod.tmpl",
		"cmd/worker",
		"cmd/mockgateway",
		"internal/worker",
		"internal/money",
		"internal/apperror",
		"test/e2e",
		"AGENTS.md",
		"ARCHITECTURE.md",
		".air.worker.toml",
	} {
		assertMissing(t, dir, p)
	}
	assertNoLeftoverNames(t, dir)

	run(t, dir, "go", "mod", "tidy")
	run(t, dir, "go", "build", "./...")
	// go vet type-checks test files too, which go build skips.
	run(t, dir, "go", "vet", "./...")
	gofmtClean(t, dir)
	run(t, dir, "go", "run", goArchLint, "check")
	run(t, dir, "make", "-n", "build")

	if exec.Command("docker", "info").Run() != nil {
		t.Log("docker unavailable: skipping go test in the generated project")
		return
	}
	run(t, dir, "go", "test", "-count=1", "./...")
}

func assertExists(t *testing.T, root, rel string) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, rel))
	assert.NoErrorf(t, err, "%s should exist", rel)
}

func assertMissing(t *testing.T, root, rel string) {
	t.Helper()
	_, err := os.Stat(filepath.Join(root, rel))
	assert.Truef(t, os.IsNotExist(err), "%s should not be generated", rel)
}

// assertNoLeftoverNames fails for any generated file that still names the
// template's domain, the template module, a dropped binary, or an unreplaced
// project-name token.
func assertNoLeftoverNames(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		s := string(data)
		assert.Falsef(t, strings.Contains(strings.ToLower(s), "ecommerce"), "%s mentions ecommerce", rel)
		assert.Falsef(t, strings.Contains(s, "go-api-project-template"), "%s mentions go-api-project-template", rel)
		assert.Falsef(t, strings.Contains(s, "__PROJECT_NAME__"), "%s has an unreplaced __PROJECT_NAME__", rel)
		for _, dropped := range []string{"mockgateway", "cmd/worker"} {
			assert.Falsef(t, strings.Contains(s, dropped), "%s mentions dropped %s", rel, dropped)
		}
		return nil
	})
	require.NoError(t, err)
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "%s %s failed in the generated project:\n%s", name, strings.Join(args, " "), out)
}

func gofmtClean(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "gofmt")
	assert.Emptyf(t, strings.TrimSpace(string(out)), "gofmt-dirty files in the generated project:\n%s", out)
}
