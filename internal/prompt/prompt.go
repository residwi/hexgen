// Package prompt provides minimal interactive input helpers.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// IsTTY reports whether f is an interactive terminal.
func IsTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// String prints "label [def]: ", reads one line from r, and returns it trimmed.
// An empty line returns def. Callers should pass a single shared *bufio.Reader
// across prompts so buffered input is not lost between calls.
func String(r *bufio.Reader, out io.Writer, label, def string) (string, error) {
	if def != "" {
		_, _ = fmt.Fprintf(out, "%s [%s]: ", label, def)
	} else {
		_, _ = fmt.Fprintf(out, "%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}
