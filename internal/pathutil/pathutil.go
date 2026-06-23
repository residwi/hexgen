// Package pathutil holds small filesystem-path helpers shared across the generator.
package pathutil

import (
	"path/filepath"
	"strings"
)

// WithinDir reports whether target is base itself or a path nested under base
// (after cleaning), rejecting traversal such as base/../x. Using filepath.Rel
// handles edge cases like base == "/" correctly. base and target should both be
// absolute or both relative.
func WithinDir(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
