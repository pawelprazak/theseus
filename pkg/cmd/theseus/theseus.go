package theseus

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/heptio/theseus/pkg/cmd/diff"
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use: "theseus",
	}

	c.AddCommand(diff.NewCommand())

	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return c
}
