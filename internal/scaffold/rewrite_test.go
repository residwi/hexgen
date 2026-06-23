package scaffold

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteContent_ReplacesModuleAndName(t *testing.T) {
	in := []byte("import \"" + TemplateModule + "/internal/core\"\n// __PROJECT_NAME__\n")
	got := rewriteContent(in, Options{Module: "github.com/me/myapp", ProjectName: "myapp"})
	want := "import \"github.com/me/myapp/internal/core\"\n// myapp\n"
	assert.Equal(t, want, string(got))
}

func TestRewriteContent_SkipsNonUTF8(t *testing.T) {
	in := []byte{0xff, 0xfe, 0x00, 0x01}
	got := rewriteContent(in, Options{Module: "x", ProjectName: "y"})
	assert.Equal(t, in, got)
}
