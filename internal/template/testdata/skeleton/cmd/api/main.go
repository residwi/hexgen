package main

import (
	"os"

	"__MODULE__/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		os.Exit(1)
	}
}
