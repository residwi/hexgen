// Package scaffold turns the embedded project tree into a generated project.
package scaffold

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"strings"
	"text/template"
)

// ModulePlaceholder is the module path the embedded Go files import. Every
// occurrence is rewritten to Options.Module during generation.
const ModulePlaceholder = "__MODULE__"

// Options configures a generation run. It is also the data context for .tmpl
// files.
type Options struct {
	Module      string // new module path, e.g. github.com/me/myapp
	ProjectName string // e.g. myapp
	Auth        bool   // include the auth and user features
}

// File is a generated file's content and permission mode.
type File struct {
	Data []byte
	Mode fs.FileMode
}

// Generate merges layers in order, so a later layer's file replaces an earlier
// one at the same path, then rewrites the module path and project-name token
// across every resulting file. Files ending in ".tmpl" are rendered with
// text/template using opts as the data context (and the ".tmpl" suffix
// stripped); a template that renders to only whitespace is omitted, which lets
// a whole file be gated behind `{{if}}`.
func Generate(opts Options, layers ...fs.FS) (map[string]File, error) {
	out := map[string]File{}
	for _, layer := range layers {
		if err := collect(layer, out, opts); err != nil {
			return nil, err
		}
	}
	for p, f := range out {
		data := rewriteContent(f.Data, opts)
		// The module rename can disturb import ordering in files that group the
		// local import with third-party ones (e.g. generated mocks). Re-canonicalize
		// Go files; best-effort, since `go build` is the real validator.
		if strings.HasSuffix(p, ".go") {
			if formatted, err := format.Source(data); err == nil {
				data = formatted
			}
		}
		out[p] = File{Data: data, Mode: f.Mode}
	}
	return out, nil
}

func collect(fsys fs.FS, out map[string]File, opts Options) error {
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		key := p
		if name, ok := strings.CutSuffix(p, ".tmpl"); ok {
			rendered, err := renderTemplate(p, data, opts)
			if err != nil {
				return err
			}
			if len(bytes.TrimSpace(rendered)) == 0 {
				return nil // entirely gated out by {{if}}
			}
			data, key = rendered, name
		}
		out[key] = File{Data: data, Mode: normalizeMode(info.Mode())}
		return nil
	})
}

func renderTemplate(name string, data []byte, opts Options) ([]byte, error) {
	t, err := template.New(name).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing override template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, opts); err != nil {
		return nil, fmt.Errorf("executing override template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// normalizeMode collapses a file mode to 0o755 for executables, 0o644 otherwise.
func normalizeMode(m fs.FileMode) fs.FileMode {
	if m.Perm()&0o111 != 0 {
		return 0o755
	}
	return 0o644
}
