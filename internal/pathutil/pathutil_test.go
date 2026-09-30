package pathutil_test

import (
	"path/filepath"
	"testing"

	"github.com/residwi/hexgen/internal/pathutil"
	"github.com/stretchr/testify/assert"
)

func TestWithinDir(t *testing.T) {
	base := filepath.FromSlash("/home/user/proj")

	cases := []struct {
		name   string
		target string
		want   bool
	}{
		{"base itself", "/home/user/proj", true},
		{"direct child", "/home/user/proj/main.go", true},
		{"nested child", "/home/user/proj/internal/server/router.go", true},
		{"parent", "/home/user", false},
		{"sibling", "/home/user/other", false},
		{"traversal out", "/home/user/proj/../escape", false},
		{"absolute escape", "/etc/passwd", false},
		// Prefix that is not actually a child: "/home/user/proj" vs
		// "/home/user/projX" — a naive HasPrefix check would wrongly accept this.
		{"shared prefix not nested", "/home/user/projX", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pathutil.WithinDir(base, filepath.FromSlash(tc.target))
			assert.Equal(t, tc.want, got)
		})
	}
}

// A target that climbs out and back in is still within the base after cleaning.
func TestWithinDir_NormalizesBeforeComparing(t *testing.T) {
	base := filepath.FromSlash("/home/user/proj")
	target := filepath.FromSlash("/home/user/proj/sub/../main.go")
	assert.True(t, pathutil.WithinDir(base, target))
}

// Rel returning an error (mismatched absolute/relative) is treated as "not within".
func TestWithinDir_RelError(t *testing.T) {
	assert.False(t, pathutil.WithinDir(filepath.FromSlash("/abs/base"), "relative/target"))
}
