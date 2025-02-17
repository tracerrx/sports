package sportboard

import (
	"context"
	"net/http"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/robbydyer/sports/internal/proto/sportboard"
)

// Server ...
type Server struct {
	board *SportBoard
}

// GetRPCHandler ...
func (s *SportBoard) GetRPCHandler() (string, http.Handler) {
	return s.rpcServer.PathPrefix(), s.rpcServer
}

// SetStatus ...
func (s *Server) SetStatus(ctx context.Context, req *pb.SetStatusReq) (*emptypb.Empty, error) {
	cancelBoard := false
	clearDrawCache := false

	if req.Status == nil {
		return &emptypb.Empty{}, twirp.NewError(twirp.InvalidArgument, "nil status sent")
	}

	if s.board.config.HideFavoriteScore.CompareAndSwap(!req.Status.FavoriteHidden, req.Status.FavoriteHidden) {
		cancelBoard = true
	}
	if s.board.Enabler().Store(req.Status.Enabled) {
		cancelBoard = true
	}
	if s.board.config.FavoriteSticky.CompareAndSwap(!req.Status.FavoriteSticky, req.Status.FavoriteSticky) {
		cancelBoard = true
	}
	if s.board.config.GamblingSpread.CompareAndSwap(!req.Status.OddsEnabled, req.Status.OddsEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.ScrollMode.CompareAndSwap(!req.Status.ScrollEnabled, req.Status.ScrollEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.TightScroll.CompareAndSwap(!req.Status.TightScrollEnabled, req.Status.TightScrollEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.ShowRecord.CompareAndSwap(!req.Status.RecordRankEnabled, req.Status.RecordRankEnabled) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.UseGradient.CompareAndSwap(!req.Status.UseGradient, req.Status.UseGradient) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.LiveOnly.CompareAndSwap(!req.Status.LiveOnly, req.Status.LiveOnly) {
		cancelBoard = true
	}
	if s.board.config.DetailedLive.CompareAndSwap(!req.Status.DetailedLive, req.Status.DetailedLive) {
		cancelBoard = true
		clearDrawCache = true
	}
	if s.board.config.ShowLeagueLogo.CompareAndSwap(!req.Status.ShowLeagueLogo, req.Status.ShowLeagueLogo) {
		clearDrawCache = true
	}

	if clearDrawCache {
		s.board.clearDrawCache()
	}

	if cancelBoard {
		s.board.callCancelBoard()
	}

	return &emptypb.Empty{}, nil
}

// GetStatus ...
func (s *Server) GetStatus(ctx context.Context, req *emptypb.Empty) (*pb.StatusResp, error) {
	return &pb.StatusResp{
		Status: &pb.Status{
			Enabled:            s.board.Enabler().Enabled(),
			FavoriteHidden:     s.board.config.HideFavoriteScore.Load(),
			FavoriteSticky:     s.board.config.FavoriteSticky.Load(),
			ScrollEnabled:      s.board.config.ScrollMode.Load(),
			TightScrollEnabled: s.board.config.TightScroll.Load(),
			RecordRankEnabled:  s.board.config.ShowRecord.Load(),
			OddsEnabled:        s.board.config.GamblingSpread.Load(),
			UseGradient:        s.board.config.UseGradient.Load(),
			LiveOnly:           s.board.config.LiveOnly.Load(),
			DetailedLive:       s.board.config.DetailedLive.Load(),
			ShowLeagueLogo:     s.board.config.ShowLeagueLogo.Load(),
		},
	}, nil
}
