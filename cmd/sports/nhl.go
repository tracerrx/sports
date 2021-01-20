package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/robbydyer/sports/pkg/nhl"
)

type nhlCmd struct {
	rArgs *rootArgs
}

func newNhlCmd(args *rootArgs) *cobra.Command {
	c := nhlCmd{
		rArgs: args,
	}

	cmd := &cobra.Command{
		Use:   "nhl",
		Short: "nhl",
		RunE:  c.run,
	}

	return cmd
}

func (c *nhlCmd) run(cmd *cobra.Command, args []string) error {
	n, err := nhl.New()
	if err != nil {
		return err
	}

	n.PrintTodaySchedule(os.Stdout)
	return nil
}
