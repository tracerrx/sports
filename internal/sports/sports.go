package sports

import (
	"context"

	sportspb "github.com/robbydyer/sports/internal/proto/sports"
	"github.com/robbydyer/sports/pkg/nhl"
)

type Server struct {
	nhlAPI *nhl.Nhl
}

func New(ctx context.Context) (*Server, error) {
	s := &Server{}

	n, err := nhl.New(ctx)
	if err != nil {
		return nil, err
	}

	s.nhlAPI = n

	return s, nil
}

func (s *Server) GetTeams(ctx context.Context, req *sportspb.GetTeamsRequest) (*sportspb.GetTeamsResponse, error) {
	return &sportspb.GetTeamsResponse{}, nil
}

func (s *Server) GetTeamByName(ctx context.Context, req *sportspb.GetTeamByNameRequest) (*sportspb.GetTeamByNameResponse, error) {
	return &sportspb.GetTeamByNameResponse{}, nil
}
func (s *Server) GetGameScoreboards(ctx context.Context, req *sportspb.GetGameScoreboardsRequest) (*sportspb.GetGameScoreboardsResponse, error) {
	return &sportspb.GetGameScoreboardsResponse{}, nil
}
