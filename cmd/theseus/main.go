package main

import (
	"os"

	"github.com/heptio/theseus/pkg/cmd/theseus"
)

func main() {
	if err := theseus.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
