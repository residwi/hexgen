package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNewArgs(t *testing.T) {
	o, err := parseNewArgs([]string{"myapp", "--module", "github.com/me/myapp", "--ref", "v1.0.0", "--force"})
	require.NoError(t, err)
	assert.Equal(t, "myapp", o.Name)
	assert.Equal(t, "github.com/me/myapp", o.Module)
	assert.Equal(t, "v1.0.0", o.Ref)
	assert.True(t, o.Force)
}

func TestParseNewArgs_DefaultRef(t *testing.T) {
	o, err := parseNewArgs([]string{"myapp", "--module", "github.com/me/myapp"})
	require.NoError(t, err)
	assert.Equal(t, "main", o.Ref)
}

func TestValidateNewOptions(t *testing.T) {
	assert.NoError(t, validateNewOptions(newOptions{Name: "x", Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "x", Module: "bad module"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "../evil", Module: "github.com/me/x"}))
	assert.Error(t, validateNewOptions(newOptions{Name: "a/b", Module: "github.com/me/x"}))
}

func TestNextSteps(t *testing.T) {
	got := nextSteps(newOptions{Name: "myapp", Output: "./myapp"})
	for _, want := range []string{"cd ./myapp", "make migrate-up", "make seed", "admin@example.com / admin123", "make test"} {
		assert.Contains(t, got, want)
	}
}

func TestParseNewArgs_Worker(t *testing.T) {
	o, err := parseNewArgs([]string{"myapp", "--module", "github.com/me/myapp", "--worker"})
	require.NoError(t, err)
	assert.True(t, o.Worker)

	def, err := parseNewArgs([]string{"myapp", "--module", "github.com/me/myapp"})
	require.NoError(t, err)
	assert.False(t, def.Worker)
}

func TestNextSteps_Worker(t *testing.T) {
	assert.Contains(t, nextSteps(newOptions{Name: "myapp", Output: "./myapp", Worker: true}), "make run-worker")
	assert.NotContains(t, nextSteps(newOptions{Name: "myapp", Output: "./myapp"}), "make run-worker")
}
