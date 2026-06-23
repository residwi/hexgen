package prompt

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString_UsesInput(t *testing.T) {
	var out bytes.Buffer
	got, err := String(bufio.NewReader(strings.NewReader("myapp\n")), &out, "Project name", "")
	require.NoError(t, err)
	assert.Equal(t, "myapp", got)
	assert.Contains(t, out.String(), "Project name")
}

func TestString_FallsBackToDefault(t *testing.T) {
	var out bytes.Buffer
	got, err := String(bufio.NewReader(strings.NewReader("\n")), &out, "Module", "github.com/me/x")
	require.NoError(t, err)
	assert.Equal(t, "github.com/me/x", got)
}

func TestString_SharedReaderAcrossPrompts(t *testing.T) {
	var out bytes.Buffer
	r := bufio.NewReader(strings.NewReader("myapp\ngithub.com/me/myapp\n"))
	name, err := String(r, &out, "Project name", "")
	require.NoError(t, err)
	assert.Equal(t, "myapp", name)
	mod, err := String(r, &out, "Module", "")
	require.NoError(t, err)
	assert.Equal(t, "github.com/me/myapp", mod)
}
