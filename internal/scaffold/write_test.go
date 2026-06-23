package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out")
	files := map[string]File{
		"go.mod":                    {Data: []byte("module x")},
		"internal/server/router.go": {Data: []byte("package server")},
	}
	require.NoError(t, Write(dest, files, false))
	got, err := os.ReadFile(filepath.Join(dest, "internal", "server", "router.go"))
	require.NoError(t, err)
	assert.Equal(t, "package server", string(got))
}

func TestWrite_RefusesNonEmptyDir(t *testing.T) {
	dest := t.TempDir() // t.TempDir is empty, so a first write is allowed
	require.NoError(t, os.WriteFile(filepath.Join(dest, "existing.txt"), []byte("x"), 0o644))
	assert.Error(t, Write(dest, map[string]File{"a": {Data: []byte("b")}}, false))
	assert.NoError(t, Write(dest, map[string]File{"a": {Data: []byte("b")}}, true))
}

func TestWrite_RejectsEscapingKey(t *testing.T) {
	dest := t.TempDir()
	err := Write(dest, map[string]File{"../escape.txt": {Data: []byte("x")}}, true)
	assert.Error(t, err)
}
