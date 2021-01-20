package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	sportspb "github.com/robbydyer/sports/internal/proto/sports"
	"github.com/robbydyer/sports/internal/sports"
)

type serverCmd struct {
	rArgs *rootArgs
	port  int
}

func newServerCmd(args *rootArgs) *cobra.Command {
	c := serverCmd{
		rArgs: args,
	}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Runs the gRPC server",
		RunE:  c.run,
	}

	return cmd
}

func (s *serverCmd) run(cmd *cobra.Command, args []string) error {
	s.port = 10000

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start net listener: %w", err)
	}

	srv, err := sports.New()
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	sportspb.RegisterSportsServer(grpcServer, srv)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		grpcServer.GracefulStop()
	}()

	fmt.Println("Starting server")
	if err := grpcServer.Serve(l); err != nil {
		return err
	}

	return nil
}
