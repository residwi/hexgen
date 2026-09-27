package scaffold

import (
	"bytes"
	"unicode/utf8"
)

// rewriteContent rewrites the module placeholder and the project-name token
// in a single file's bytes. Non-UTF-8 (binary) content is returned unchanged.
func rewriteContent(data []byte, opts Options) []byte {
	if !utf8.Valid(data) {
		return data
	}
	out := bytes.ReplaceAll(data, []byte(ModulePlaceholder), []byte(opts.Module))
	out = bytes.ReplaceAll(out, []byte("__PROJECT_NAME__"), []byte(opts.ProjectName))
	return out
}
