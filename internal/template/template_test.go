package template

import (
	"io/fs"
	"strings"
	"testing"
	texttemplate "text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkeleton_ContainsRequiredFiles(t *testing.T) {
	for _, p := range []string{
		"go.mod.tmpl",
		"go.sum",
		"cmd/api/main.go",
		"internal/config/config.go",
		"internal/server/server.go",
		"internal/testutil/testutil.go",
		"internal/platform/web/router.go",
		".github/workflows/ci.yml",
		".golangci.yml",
		".gitignore",
		"bin/.keep",
	} {
		_, err := fs.Stat(Skeleton(), p)
		assert.NoErrorf(t, err, "skeleton should contain %s", p)
	}
}

func TestAuth_ContainsRequiredFiles(t *testing.T) {
	for _, p := range []string{
		"internal/features/auth/service.go",
		"internal/features/user/service.go",
		"internal/testutil/fixtures.go",
		"db/seeds/data.sql",
	} {
		_, err := fs.Stat(Auth(), p)
		assert.NoErrorf(t, err, "auth overlay should contain %s", p)
	}
}

func TestTrees_ExcludeECommerceAndWorkerPaths(t *testing.T) {
	for name, tree := range map[string]fs.FS{"skeleton": Skeleton(), "auth": Auth()} {
		for _, p := range []string{
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
			_, err := fs.Stat(tree, p)
			assert.ErrorIsf(t, err, fs.ErrNotExist, "%s should not contain %s", name, p)
		}
	}
}

func TestTrees_ContainNoECommerceName(t *testing.T) {
	for name, tree := range map[string]fs.FS{"skeleton": Skeleton(), "auth": Auth()} {
		err := fs.WalkDir(tree, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := fs.ReadFile(tree, p)
			require.NoError(t, err)
			assert.Falsef(t, strings.Contains(strings.ToLower(string(data)), "ecommerce"),
				"%s: %s should not mention ecommerce", name, p)
			return nil
		})
		require.NoError(t, err)
	}
}

func TestSkeleton_HasNoFeatures(t *testing.T) {
	_, err := fs.Stat(Skeleton(), "internal/features")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestAuth_HasOnlyAuthAndUserFeatures(t *testing.T) {
	entries, err := fs.ReadDir(Auth(), "internal/features")
	require.NoError(t, err)

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	assert.ElementsMatch(t, []string{"auth", "user"}, names)
}

func TestTemplates_Parse(t *testing.T) {
	for name, tree := range map[string]fs.FS{"skeleton": Skeleton(), "auth": Auth()} {
		err := fs.WalkDir(tree, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".tmpl") {
				return err
			}
			data, err := fs.ReadFile(tree, p)
			require.NoError(t, err)
			_, err = texttemplate.New(p).Parse(string(data))
			assert.NoErrorf(t, err, "%s: %s should parse as text/template", name, p)
			return nil
		})
		require.NoError(t, err)
	}
}
