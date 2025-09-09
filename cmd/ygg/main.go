package main

import (
	"os"

	"github.com/joch/ygg/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}