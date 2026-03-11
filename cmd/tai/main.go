package main

import (
	"os"

	"github.com/NitorCreations/tai/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
