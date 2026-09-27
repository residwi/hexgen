package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type newOptions struct {
	Name   string
	Module string
	Ref    string
	Output string
	Force  bool
	Git    bool
	Check  bool
	Worker bool
}

func parseNewArgs(args []string) (newOptions, error) {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	var o newOptions
	fs.StringVar(&o.Module, "module", "", "Go module path (required), e.g. github.com/me/myapp")
	fs.StringVar(&o.Ref, "ref", "main", "template git ref (branch, tag, or sha)")
	fs.StringVar(&o.Output, "output", "", "output directory (default ./<name>)")
	fs.BoolVar(&o.Force, "force", false, "write into a non-empty directory")
	fs.BoolVar(&o.Git, "git", false, "run 'git init' in the generated project")
	fs.BoolVar(&o.Check, "check", false, "run 'go build ./...' in the output after generating")
	fs.BoolVar(&o.Worker, "worker", false, "include a background worker (cmd/worker on the jobs runner + worker config/tooling)")

	// Extract the leading positional name argument before flag parsing, so that
	// flags appearing after the name are handled correctly by flag.FlagSet.
	flagArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		o.Name = args[0]
		flagArgs = args[1:]
	}

	if err := fs.Parse(flagArgs); err != nil {
		return o, err
	}
	return o, nil
}

func validateNewOptions(o newOptions) error {
	if o.Name == "" {
		return errors.New("project name is required (usage: hexgen new <name> --module <path>)")
	}
	if strings.ContainsAny(o.Name, `/\`) || o.Name == "." || o.Name == ".." || strings.HasPrefix(o.Name, "-") {
		return errors.New(`project name must be a single path segment (no '/', '\', '.', '..', or leading '-')`)
	}
	if o.Module == "" {
		return errors.New("--module is required (e.g. --module github.com/me/myapp)")
	}
	if strings.ContainsAny(o.Module, " \t\n") {
		return errors.New("--module must not contain whitespace")
	}
	if strings.HasPrefix(o.Module, "/") || strings.HasSuffix(o.Module, "/") || strings.Contains(o.Module, "//") {
		return errors.New("--module is not a valid module path (no leading/trailing or doubled '/')")
	}
	return nil
}

func nextSteps(o newOptions) string {
	dir := o.Output
	var b strings.Builder
	fmt.Fprintf(&b, "Created %s at %s\n\n", o.Name, dir)
	b.WriteString("Next steps:\n")
	fmt.Fprintf(&b, "  cd %s\n", dir)
	b.WriteString("  go mod tidy\n")
	b.WriteString("  make setup            # install dev tooling (mockery, goose, air, golangci-lint)\n")
	b.WriteString("  cp .env.example .env  # then edit DATABASE_URL etc.\n")
	b.WriteString("  make migrate-up\n")
	b.WriteString("  make seed             # seeds admin user: admin@example.com / admin123  (change this!)\n")
	b.WriteString("  make run\n")
	if o.Worker {
		b.WriteString("  make run-worker       # run the background worker (separate process)\n")
	}
	b.WriteString("\n")
	b.WriteString("Run tests:  make test    (integration tests require Docker)\n")
	return b.String()
}
