// Package fetch downloads and extracts the project template tarball.
package fetch

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/residwi/go-project-generator/internal/pathutil"
)

// TarballURL builds the codeload tarball URL for a repo at a ref. If baseURL is
// empty, https://codeload.github.com is used.
func TarballURL(baseURL, repo, ref string) string {
	if baseURL == "" {
		baseURL = "https://codeload.github.com"
	}
	return fmt.Sprintf("%s/%s/tar.gz/%s", baseURL, repo, ref)
}

const (
	maxFileSize  = 100 << 20 // 100 MiB per extracted file
	maxTotalSize = 500 << 20 // 500 MiB total across the archive
	maxEntries   = 20000     // guard against archives with absurd entry counts
)

// ExtractTarGz extracts a gzipped tar stream into destDir and returns the name
// of the single top-level directory the archive entries share. It rejects path
// traversal, link entries, and archives that exceed size/entry limits.
func ExtractTarGz(r io.Reader, destDir string) (string, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	root := ""
	cleanDest := filepath.Clean(destDir)
	var total int64
	entries := 0

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar: %w", err)
		}

		// Skip PAX global/extended headers; they are metadata, not file entries.
		if hdr.Typeflag == tar.TypeXGlobalHeader || hdr.Typeflag == tar.TypeXHeader {
			continue
		}

		if entries++; entries > maxEntries {
			return "", fmt.Errorf("archive has too many entries (>%d)", maxEntries)
		}

		name := path.Clean(hdr.Name)
		if name == "." {
			continue
		}
		if root == "" {
			root = strings.SplitN(name, "/", 2)[0]
		}

		// Reject any entry that resolves outside destDir (zip-slip).
		target := filepath.Join(cleanDest, filepath.FromSlash(name))
		if !pathutil.WithinDir(cleanDest, target) {
			return "", fmt.Errorf("unsafe path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			mode := os.FileMode(0o644)
			if hdr.FileInfo().Mode().Perm()&0o111 != 0 {
				mode = 0o755
			}
			n, err := writeRegular(target, tr, mode, maxFileSize)
			if err != nil {
				return "", err
			}
			total += n
			if total > maxTotalSize {
				return "", fmt.Errorf("archive exceeds %d bytes uncompressed", maxTotalSize)
			}
		case tar.TypeSymlink, tar.TypeLink:
			return "", fmt.Errorf("unsupported link entry in archive: %s", hdr.Name)
		}
	}

	if root == "" {
		return "", errors.New("empty archive")
	}
	if fi, err := os.Stat(filepath.Join(cleanDest, root)); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("archive has no single top-level directory (root %q)", root)
	}
	return root, nil
}

// writeRegular copies at most max bytes from r into a new file at target with
// the given mode, returning the bytes written. It errors if the entry exceeds
// max or if closing the file fails (e.g. a deferred-flush error on a network FS).
func writeRegular(target string, r io.Reader, mode os.FileMode, max int64) (int64, error) {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return 0, err
	}
	n, copyErr := io.Copy(f, io.LimitReader(r, max+1))
	closeErr := f.Close()
	switch {
	case copyErr != nil:
		return n, copyErr
	case n > max:
		return n, fmt.Errorf("file %s exceeds %d bytes", target, max)
	case closeErr != nil:
		return n, fmt.Errorf("closing %s: %w", target, closeErr)
	}
	return n, nil
}

// GitHub fetches the template as a tarball from GitHub's codeload host.
type GitHub struct {
	Repo    string       // owner/name
	BaseURL string       // default https://codeload.github.com
	Client  *http.Client // default http.DefaultClient
}

// Fetch downloads and extracts the template at ref, returning an fs.FS rooted at
// the extracted project plus a cleanup function that removes the temp dir.
func (g *GitHub) Fetch(ctx context.Context, ref string) (fs.FS, func(), error) {
	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TarballURL(g.BaseURL, g.Repo, ref), nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch template: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, fmt.Errorf("template ref %q not found at %s", ref, g.Repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("fetch template: unexpected status %s", resp.Status)
	}

	tmp, err := os.MkdirTemp("", "gen-template-*")
	if err != nil {
		return nil, nil, err
	}
	root, err := ExtractTarGz(resp.Body, tmp)
	if err != nil {
		_ = os.RemoveAll(tmp)
		return nil, nil, err
	}
	return os.DirFS(filepath.Join(tmp, root)), func() { _ = os.RemoveAll(tmp) }, nil
}
