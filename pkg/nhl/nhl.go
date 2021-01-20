package nhl

import (
	"fmt"
	"io"
	"time"
)

const (
	BaseURL = "http://statsapi.web.nhl.com/api/v1/"
)

type Nhl struct {
	Teams map[int]*Team
	Games map[string][]*Game
}

func New() (*Nhl, error) {
	n := &Nhl{
		Games: make(map[string][]*Game),
		Teams: make(map[int]*Team),
	}

	if err := n.UpdateTeams(); err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")

	if err := n.UpdateGames(today); err != nil {
		return nil, fmt.Errorf("failed to get today's games: %w", err)
	}

	return n, nil
}

func (n *Nhl) UpdateTeams() error {
	teamList, err := GetTeams()
	if err != nil {
		return err
	}

	n.Teams = teamList

	return nil
}

func (n *Nhl) UpdateGames(dateStr string) error {
	games, err := getGames(dateStr)
	if err != nil {
		return err
	}

	n.Games[dateStr] = games

	return nil
}

func (n *Nhl) nameFromID(id int) (string, error) {
	t, ok := n.Teams[id]
	if !ok {
		if err := n.UpdateTeams(); err != nil {
			return "", err
		}
	}

	return t.Name, nil
}

func (n *Nhl) PrintTodaySchedule(out io.Writer) error {
	return n.PrintSchedule(time.Now().Format("2006-01-02"), out)
}

func (n *Nhl) PrintSchedule(dateStr string, out io.Writer) error {
	if err := validateDateStr(dateStr); err != nil {
		return err
	}

	games, ok := n.Games[dateStr]
	if !ok {
		if err := n.UpdateGames(dateStr); err != nil {
			return err
		}
	}

	for _, game := range games {
		away, err := n.nameFromID(game.Teams.Away.Team.Id)
		if err != nil {
			return err
		}
		home, err := n.nameFromID(game.Teams.Home.Team.Id)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Home: %s\nAway:%s\n%s\n\n", home, away, game.GameTime.Local().Format("07:05PM"))
	}

	return nil
}

func validateDateStr(dateStr string) error {
	// TODO: Do this
	return nil
}
