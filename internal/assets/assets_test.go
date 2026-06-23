package assets

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverrides_ContainsExpectedFiles(t *testing.T) {
	want := []string{
		"internal/server/router.go.tmpl",
		"internal/server/router_test.go.tmpl",
		"internal/config/config.go.tmpl",
		"internal/config/config_test.go.tmpl",
		"internal/core/apperror.go.tmpl",
		"internal/core/response/error_mapper.go.tmpl",
		"cmd/worker/main.go.tmpl",
		".mockery.yml",
		"db/seeds/data.sql",
		"README.md",
		"Makefile.tmpl",
		"compose.yml.tmpl",
		"Dockerfile.tmpl",
		".env.example.tmpl",
		".air.worker.toml.tmpl",
	}
	o := Overrides()
	for _, p := range want {
		_, err := fs.Stat(o, p)
		assert.NoErrorf(t, err, "missing override %q", p)
	}
}
