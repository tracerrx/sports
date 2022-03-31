package basicconfig

import (
	"time"

	"go.uber.org/atomic"

	"github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
)

var defaultBoardDelay = 10 * time.Second

// Todayer is a func that returns a string representing a date
// that will be used for determining "Today's" games.
// This is useful in testing what past days looked like
type Todayer func() []time.Time

// Config ...
type Config struct {
	boardDelay         *time.Duration
	scrollDelay        *time.Duration
	TodayFunc          Todayer
	Enabled            *atomic.Bool `json:"enabled"`
	BoardDelay         string       `json:"boardDelay"`
	ScrollMode         *atomic.Bool `json:"scrollMode"`
	ScrollDelay        string       `json:"scrollDelay"`
	OnTimes            []string     `json:"onTimes"`
	OffTimes           []string     `json:"offTimes"`
	TightScrollPadding int          `json:"tightScrollPadding"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = atomic.NewBool(false)
	}
	if c.ScrollMode == nil {
		c.ScrollMode = atomic.NewBool(false)
	}
}

// GetBoardDelay ...
func (c *Config) GetBoardDelay() time.Duration {
	if c.boardDelay != nil {
		return *c.boardDelay
	}

	if c.BoardDelay != "" {
		d, err := time.ParseDuration(c.BoardDelay)
		if err != nil {
			c.boardDelay = &defaultBoardDelay
		}
		c.boardDelay = &d
	} else {
		c.boardDelay = &defaultBoardDelay
	}

	return *c.boardDelay
}

// GetScrollDelay ...
func (c *Config) GetScrollDelay() time.Duration {
	if c.scrollDelay != nil {
		return *c.scrollDelay
	}

	if c.ScrollDelay != "" {
		d, err := time.ParseDuration(c.ScrollDelay)
		if err != nil {
			c.scrollDelay = &rgbmatrix.DefaultScrollDelay
		}
		c.scrollDelay = &d
	} else {
		c.scrollDelay = &rgbmatrix.DefaultScrollDelay
	}

	return *c.scrollDelay
}
