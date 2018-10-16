package main

import (
	"os"

	"github.com/gshamov/lustrebeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
