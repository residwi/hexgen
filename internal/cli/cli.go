// Package cli implements the gen command-line interface.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/residwi/go-project-generator/internal/assets"
	"github.com/residwi/go-project-generator/internal/fetch"
	"github.com/residwi/go-project-generator/internal/prompt"
	"github.com/residwi/go-project-generator/internal/scaffold"
)

const version = "0.1.0-dev"

const defaultRepo = "residwi/go-api-project-template"

const usage = `gen - bootstrap a Go API project from go-api-project-template

Usage:
  gen new <name> --module <path> [--ref main] [--output dir] [--force] [--git] [--check]
  gen version

Flags for "new":
  --module   Go module path (required), e.g. github.com/me/myapp
  --ref      template ref to fetch (default "main")
  --output   output directory (default ./<name>)
  --force    write into a non-empty directory
  --git      run 'git init' in the generated project
  --check    run 'go build ./...' in the output after generating
  --worker   include a background worker (cmd/worker + jobs runner + worker config/tooling)`

// Run dispatches a gen invocation and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "new":
		return runNew(args[1:])
	case "version":
		fmt.Println(version)
		return 0
	case "-h", "--help", "help":
		fmt.Println(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s\n", args[0], usage)
		return 2
	}
}

func runNew(args []string) int {
	o, err := parseNewArgs(args)
	if err != nil {
		return 2 // flag package already printed the error
	}

	if (o.Name == "" || o.Module == "") && prompt.IsTTY(os.Stdin) {
		stdin := bufio.NewReader(os.Stdin)
		if o.Name == "" {
			o.Name, err = prompt.String(stdin, os.Stdout, "Project name", "")
			if err != nil {
				fmt.Fprintln(os.Stderr, "error reading input:", err)
				return 1
			}
		}
		if o.Module == "" {
			def := ""
			if o.Name != "" {
				def = "github.com/me/" + o.Name
			}
			o.Module, err = prompt.String(stdin, os.Stdout, "Go module path", def)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error reading input:", err)
				return 1
			}
		}
	}

	if err := validateNewOptions(o); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	if o.Output == "" {
		o.Output = "./" + o.Name
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	src := &fetch.GitHub{Repo: defaultRepo}
	fsys, cleanup, err := src.Fetch(ctx, o.Ref)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer cleanup()

	files, err := scaffold.Generate(fsys, scaffold.Options{
		Module:      o.Module,
		ProjectName: o.Name,
		Worker:      o.Worker,
	}, assets.Overrides())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if err := scaffold.Write(o.Output, files, o.Force); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if o.Git {
		if err := runCmd(o.Output, "git", "init"); err != nil {
			fmt.Fprintln(os.Stderr, "warning: git init failed:", err)
		}
	}
	if o.Check {
		if err := runCmd(o.Output, "go", "build", "./..."); err != nil {
			fmt.Fprintln(os.Stderr, "warning: 'go build ./...' failed in generated project:", err)
		}
	}

	fmt.Print(nextSteps(o))
	return 0
}

func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
