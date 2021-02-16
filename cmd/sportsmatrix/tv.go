package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/robbydyer/sports/pkg/sportsmatrix"
	"github.com/robbydyer/sports/pkg/tv"
)

type TVCmd struct {
	rArgs *rootArgs
}

func newTVCmd(args *rootArgs) *cobra.Command {
	c := TVCmd{
		rArgs: args,
	}

	cmd := &cobra.Command{
		Use:   "tv",
		Short: "Runs in TV display mode",
		RunE:  c.run,
	}

	return cmd
}

func (s *TVCmd) run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		fmt.Println("Got OS interrupt signal, Shutting down")
		cancel()
	}()

	logger := getLogger(s.rArgs.logLevel)

	boards, err := s.rArgs.getBoards(ctx, logger)
	if err != nil {
		return err
	}

	canvas := tv.New(s.rArgs.config.SportsMatrixConfig.HardwareConfig.Cols, s.rArgs.config.SportsMatrixConfig.HardwareConfig.Rows)

	mtrx, err := sportsmatrix.New(ctx, logger, s.rArgs.config.SportsMatrixConfig, canvas, boards...)
	if err != nil {
		return err
	}
	defer mtrx.Close()

	fmt.Println("Starting matrix service")
	if err := mtrx.Serve(ctx); err != nil {
		fmt.Printf("Matrix returned an error: %s", err.Error())
		return err
	}

	return nil
}
