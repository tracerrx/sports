package racing

import (
	"context"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/robbydyer/sports/internal/proto/racingboard"
)

// Server ...
type Server struct {
	board *Racing
}

// SetStatus ...
func (s *Server) SetStatus(ctx context.Context, req *pb.SetStatusReq) (*emptypb.Empty, error) {
	cancelBoard := false
	if req.Status == nil {
		return &emptypb.Empty{}, twirp.NewError(twirp.InvalidArgument, "nil status sent")
	}

	if req.Status.Enabled {
		if s.board.board.Enable() {
			cancelBoard = true
		}
	} else {
		if s.board.board.Disable() {
			cancelBoard = true
		}
	}
	if s.board.config.ScrollMode.CAS(!req.Status.ScrollEnabled, req.Status.ScrollEnabled) {
		cancelBoard = true
	}

	if cancelBoard {
		s.board.board.Cancel()
	}

	return &emptypb.Empty{}, nil
}

// GetStatus ...
func (s *Server) GetStatus(ctx context.Context, req *emptypb.Empty) (*pb.StatusResp, error) {
	return &pb.StatusResp{
		Status: &pb.Status{
			Enabled:       s.board.config.Enabled.Load(),
			ScrollEnabled: s.board.config.ScrollMode.Load(),
		},
	}, nil
}
