// Package cli implements the hexgen command-line interface.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/residwi/go-project-generator/internal/prompt"
	"github.com/residwi/go-project-generator/internal/scaffold"
	"github.com/residwi/go-project-generator/internal/template"
)

const version = "0.1.0-dev"

const usage = `hexgen - bootstrap a platform-only Go API project

Usage:
  hexgen new <name> --module <path> [--output dir] [--force] [--git] [--check]
  hexgen version

Flags for "new":
  --module   Go module path (required), e.g. github.com/me/myapp
  --output   output directory (default ./<name>)
  --force    write into a non-empty directory
  --git      run 'git init' in the generated project
  --check    run 'go build ./...' in the output after generating`

// Run dispatches a hexgen invocation and returns a process exit code.
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

	files, err := scaffold.Generate(scaffold.Options{
		Module:      o.Module,
		ProjectName: o.Name,
	}, template.Skeleton())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if err := scaffold.Write(o.Output, files, o.Force); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Drops template dependencies the generated project no longer imports.
	if err := runCmd(o.Output, "go", "mod", "tidy"); err != nil {
		fmt.Fprintln(os.Stderr, "warning: 'go mod tidy' failed in generated project:", err)
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
