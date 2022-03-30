package basicconfig

import (
	"time"

	"github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
	"go.uber.org/atomic"
)

type Todayer func() []time.Time

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
			*c.boardDelay = 10 * time.Second
		}
		*c.boardDelay = d
	} else {
		*c.boardDelay = 10 * time.Second
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
			*c.scrollDelay = rgbmatrix.DefaultScrollDelay
		}
		*c.scrollDelay = d
	} else {
		*c.scrollDelay = rgbmatrix.DefaultScrollDelay
	}

	return *c.scrollDelay
}
