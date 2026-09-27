package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNewArgs(t *testing.T) {
	o, err := parseNewArgs([]string{"myapp", "--module", "github.com/me/myapp", "--force"})
	require.NoError(t, err)
	assert.Equal(t, "myapp", o.Name)
	assert.Equal(t, "github.com/me/myapp", o.Module)
	assert.True(t, o.Force)
}

func TestRun_RejectsRemovedFlags(t *testing.T) {
	for _, flag := range [][]string{{"--ref", "main"}, {"--worker"}} {
		// --output keeps a regression (flag accepted, project generated) out of the repo.
		args := append([]string{"new", "myapp", "--module", "github.com/me/myapp", "--output", t.TempDir()}, flag...)
		assert.Equalf(t, 2, Run(args), "%v should be an unknown flag", flag)
	}
}

func TestValidateNewOptions(t *testing.T) {
	assert.NoError(t, validateNewOptions(newOptions{Name: "x", Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "x", Module: "bad module"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "../evil", Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "a/b", Module: "github.com/me/x"}))
}

func TestUsageNamesHexgen(t *testing.T) {
	assert.Contains(t, usage, "hexgen new <name> --module <path>")
	assert.ErrorContains(t, validateNewOptions(newOptions{Module: "github.com/me/x"}), "usage: hexgen new")
}

func TestNextSteps(t *testing.T) {
	got := nextSteps(newOptions{Name: "myapp", Output: "./myapp"})
	for _, want := range []string{"cd ./myapp", "make migrate-up", "make seed", "admin@example.com / admin123", "make test"} {
		assert.Contains(t, got, want)
	}
	assert.NotContains(t, got, "go mod tidy", "hexgen new runs go mod tidy itself")
}
