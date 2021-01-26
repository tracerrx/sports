package nhl

import (
	"context"
	"fmt"
	"io"
	"time"
)

const (
	BaseURL  = "http://statsapi.web.nhl.com/api/v1/"
	LinkBase = "http://statsapi.web.nhl.com"
)

type Nhl struct {
	teams map[int]*Team
	games map[string][]*Game
}

func New(ctx context.Context) (*Nhl, error) {
	n := &Nhl{
		games: make(map[string][]*Game),
		teams: make(map[int]*Team),
	}

	if err := n.UpdateTeams(ctx); err != nil {
		return nil, err
	}

	fmt.Printf("Getting games for %s\n", Today())
	if err := n.UpdateGames(ctx, Today()); err != nil {
		return nil, fmt.Errorf("failed to get today's games: %w", err)
	}

	return n, nil
}

func (n *Nhl) UpdateTeams(ctx context.Context) error {
	teamList, err := GetTeams(ctx)
	if err != nil {
		return err
	}

	n.teams = teamList

	return nil
}

func (n *Nhl) UpdateGames(ctx context.Context, dateStr string) error {
	games, err := getGames(ctx, dateStr)
	if err != nil {
		return err
	}

	n.games[dateStr] = games

	return nil
}

func (n *Nhl) TeamFromAbbreviation(abbrev string) (*Team, error) {
	for _, t := range n.teams {
		if t.Abbreviation == abbrev {
			return t, nil
		}
	}

	return nil, fmt.Errorf("could not find team with abbreviation '%s'", abbrev)
}

func (n *Nhl) nameFromID(ctx context.Context, id int) (string, error) {
	t, ok := n.teams[id]
	if !ok {
		if err := n.UpdateTeams(ctx); err != nil {
			return "", err
		}
	}

	return t.Name, nil
}

func (n *Nhl) PrintTodaySchedule(ctx context.Context, out io.Writer) error {
	return n.PrintSchedule(ctx, Today(), out)
}

func (n *Nhl) PrintSchedule(ctx context.Context, dateStr string, out io.Writer) error {
	if err := validateDateStr(dateStr); err != nil {
		return err
	}

	games, ok := n.games[dateStr]
	if !ok {
		if err := n.UpdateGames(ctx, dateStr); err != nil {
			return err
		}
	}

	for _, game := range games {
		away, err := n.nameFromID(ctx, game.Teams.Away.Team.ID)
		if err != nil {
			return err
		}
		home, err := n.nameFromID(ctx, game.Teams.Home.Team.ID)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Home: %s\nAway:%s\n%s\n\n", home, away, game.GameTime.Local().Format("07:05PM"))
	}

	return nil
}

// Today is sometimes actually yesterday
func Today() string {
	// Don't update until the morning, because games might go past midnight
	if time.Now().Local().Hour() < 4 {
		return time.Now().AddDate(0, 0, -1).Local().Format("2006-01-02")
	}
	return time.Now().Local().Format("2006-01-02")
}

func validateDateStr(dateStr string) error {
	// TODO: Do this
	return nil
}
