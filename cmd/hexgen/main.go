package main

import (
	"os"

	"github.com/residwi/go-project-generator/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
