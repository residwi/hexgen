package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/residwi/hexgen/internal/pathutil"
)

// Write writes the generated files under dest. If dest exists and is non-empty
// and force is false, it returns an error without writing anything. Keys that
// would resolve outside dest are rejected.
func Write(dest string, files map[string]File, force bool) error {
	// Check the same cleaned path the files are written to: a raw dest like
	// missing/../existing fails the OS lookup yet cleans to an existing dir.
	cleanDest := filepath.Clean(dest)
	empty, err := dirEmptyOrAbsent(cleanDest)
	if err != nil {
		return err
	}
	if !empty && !force {
		return fmt.Errorf("destination %q is not empty (use --force to overwrite)", dest)
	}
	for p, f := range files {
		full := filepath.Join(cleanDest, filepath.FromSlash(p))
		if !pathutil.WithinDir(cleanDest, full) {
			return fmt.Errorf("refusing to write outside destination: %q", p)
		}
		mode := f.Mode
		if mode == 0 {
			mode = 0o644 // defensive: a File built without a Mode still gets a sane perm
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, f.Data, mode); err != nil {
			return err
		}
	}
	return nil
}

func dirEmptyOrAbsent(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
