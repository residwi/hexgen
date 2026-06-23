package fetch

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func TestTarballURL(t *testing.T) {
	got := TarballURL("", "residwi/go-api-project-template", "main")
	assert.Equal(t, "https://codeload.github.com/residwi/go-api-project-template/tar.gz/main", got)
}

func TestExtractTarGz(t *testing.T) {
	data := makeTarGz(t, map[string]string{
		"repo-main/go.mod":          "module x",
		"repo-main/cmd/api/main.go": "package main",
	})
	dest := t.TempDir()
	root, err := ExtractTarGz(bytes.NewReader(data), dest)
	require.NoError(t, err)
	assert.Equal(t, "repo-main", root)
	got, err := os.ReadFile(filepath.Join(dest, "repo-main", "cmd", "api", "main.go"))
	require.NoError(t, err)
	assert.Equal(t, "package main", string(got))
}

func TestExtractTarGz_SkipsPaxGlobalHeader(t *testing.T) {
	// GitHub codeload tarballs begin with a pax_global_header entry; it must be
	// skipped so the real top-level directory is detected as the root.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeXGlobalHeader,
		Name:     "pax_global_header",
		Size:     0,
	}))
	const content = "module x"
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "repo-main/go.mod",
		Mode:     0o644,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}))
	_, err := tw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	dest := t.TempDir()
	root, err := ExtractTarGz(bytes.NewReader(buf.Bytes()), dest)
	require.NoError(t, err)
	assert.Equal(t, "repo-main", root)
}

func TestGitHubFetch(t *testing.T) {
	data := makeTarGz(t, map[string]string{"repo-main/go.mod": "module x"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/residwi/go-api-project-template/tar.gz/main" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	gh := &GitHub{Repo: "residwi/go-api-project-template", BaseURL: srv.URL, Client: srv.Client()}
	fsys, cleanup, err := gh.Fetch(context.Background(), "main")
	require.NoError(t, err)
	defer cleanup()

	got, err := fs.ReadFile(fsys, "go.mod")
	require.NoError(t, err)
	assert.Equal(t, "module x", string(got))
}

func TestGitHubFetch_RefNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()
	gh := &GitHub{Repo: "residwi/go-api-project-template", BaseURL: srv.URL, Client: srv.Client()}
	_, _, err := gh.Fetch(context.Background(), "nope")
	assert.Error(t, err)
}
