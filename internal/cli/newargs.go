package cli

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

type newOptions struct {
	Name   string
	Module string
	Output string
	Force  bool
	Git    bool
	Check  bool
	Auth   bool
}

// projectName keeps the name valid as a Postgres database name (at most 63
// bytes) and as Docker container and image names (lowercase, no leading or
// trailing separator), all of which the generated project derives from it.
var projectName = regexp.MustCompile(`^[a-z]([a-z0-9_-]*[a-z0-9])?$`)

const maxProjectName = 63

func parseNewArgs(args []string) (newOptions, error) {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	var o newOptions
	fs.StringVar(&o.Module, "module", "", "Go module path (required), e.g. github.com/me/myapp")
	fs.StringVar(&o.Output, "output", "", "output directory (default ./<name>)")
	fs.BoolVar(&o.Force, "force", false, "write into a non-empty directory")
	fs.BoolVar(&o.Git, "git", false, "run 'git init' in the generated project")
	fs.BoolVar(&o.Check, "check", false, "run 'go build ./...' in the output after generating")
	fs.BoolVar(&o.Auth, "auth", false, "include the auth and user features (register/login/refresh, /users/me, admin user routes)")

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
	if !projectName.MatchString(o.Name) || len(o.Name) > maxProjectName {
		return fmt.Errorf("project name %q must be lowercase letters, digits, '-' or '_', start with a letter, "+
			"end with a letter or digit, and be at most %d characters: it names the database and Docker containers",
			o.Name, maxProjectName)
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
	b.WriteString("  make setup            # install dev tooling (goose, air, golangci-lint)\n")
	b.WriteString("  cp .env.example .env  # then edit DATABASE_URL etc.\n")
	b.WriteString("  make docker-up        # start postgres and redis\n")
	b.WriteString("  make migrate-up\n")
	if o.Auth {
		b.WriteString("  make seed             # seeds admin user: admin@example.com / admin123456  (change this!)\n")
	}
	b.WriteString("  make run\n")
	b.WriteString("\n")
	b.WriteString("Run tests:  make test    (integration tests require Docker)\n")
	return b.String()
}
