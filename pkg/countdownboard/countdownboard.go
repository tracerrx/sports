package countdownboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/twitchtv/twirp"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/logo"
	"github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"github.com/robbydyer/sports/pkg/twirphelpers"

	pb "github.com/robbydyer/sports/internal/proto/basicboard"
)

// CountdownBoard implements board.Board
type CountdownBoard struct {
	config              *Config
	log                 *zap.Logger
	scheduleWriter      *rgbrender.TextWriter
	leagueLogo          *logo.Logo
	rpcServer           pb.TwirpServer
	boardCtx            context.Context
	boardCancel         context.CancelFunc
	stateChangeNotifier board.StateChangeNotifier
	icon                *logo.Logo
}

// Config ...
type Config struct {
	boardDelay         time.Duration
	scrollDelay        time.Duration
	Enabled            *atomic.Bool `json:"enabled"`
	BoardDelay         string       `json:"boardDelay"`
	ScrollMode         *atomic.Bool `json:"scrollMode"`
	ScrollDelay        string       `json:"scrollDelay"`
	OnTimes            []string     `json:"onTimes"`
	OffTimes           []string     `json:"offTimes"`
	TightScrollPadding int          `json:"tightScrollPadding"`
	Events             []*Event     `json:"events"`
}

// Event ...
type Event struct {
	// EventDate is the date of the event in the format YYYY-MM-DD
	date    time.Time
	Date    string `json:"date"`
	Title   string `json:"title"`
	IconURL string `json:"iconURL"`
}

// SetDefaults sets config defaults
func (c *Config) SetDefaults() {
	if c.BoardDelay != "" {
		d, err := time.ParseDuration(c.BoardDelay)
		if err != nil {
			c.boardDelay = 10 * time.Second
		}
		c.boardDelay = d
	} else {
		c.boardDelay = 10 * time.Second
	}

	if c.Enabled == nil {
		c.Enabled = atomic.NewBool(false)
	}
	if c.ScrollMode == nil {
		c.ScrollMode = atomic.NewBool(false)
	}
	if c.ScrollDelay != "" {
		d, err := time.ParseDuration(c.ScrollDelay)
		if err != nil {
			c.scrollDelay = rgbmatrix.DefaultScrollDelay
		}
		c.scrollDelay = d
	} else {
		c.scrollDelay = rgbmatrix.DefaultScrollDelay
	}
}

// New ...
func New(icon *logo.Logo, logger *zap.Logger, config *Config) (*CountdownBoard, error) {
	s := &CountdownBoard{
		config: config,
		log:    logger,
		icon:   icon,
	}

	s.log.Info("Register Countdown Board",
		zap.String("board name", s.Name()),
	)

	c := cron.New()

	for _, on := range config.OnTimes {
		s.log.Info("countdownboard will be schedule to turn on",
			zap.String("turn on", on),
		)
		_, err := c.AddFunc(on, func() {
			s.log.Info("sportboard turning on")
			s.Enable()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add cron for sportboard: %w", err)
		}
	}

	for _, off := range config.OffTimes {
		s.log.Info("countdownboard will be schedule to turn off",
			zap.String("turn on", off),
		)
		_, err := c.AddFunc(off, func() {
			s.log.Info("countdownboard turning off")
			s.Disable()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add cron for sportboard: %w", err)
		}
	}

	if _, err := c.AddFunc("0 4 * * *", s.cacheClear); err != nil {
		return nil, err
	}

	c.Start()

	svr := &Server{
		board: s,
	}
	prfx := "/countdown"
	s.rpcServer = pb.NewBasicBoardServer(svr,
		twirp.WithServerPathPrefix(prfx),
		twirp.ChainHooks(
			twirphelpers.GetDefaultHooks(s, s.log),
		),
	)

	return s, nil
}

func (s *CountdownBoard) cacheClear() {
}

// Name ...
func (s *CountdownBoard) Name() string {
	return "countdown"
}

// Enabled ...
func (s *CountdownBoard) Enabled() bool {
	return s.config.Enabled.Load()
}

// Enable ...
func (s *CountdownBoard) Enable() bool {
	if s.config.Enabled.CAS(false, true) {
		if s.stateChangeNotifier != nil {
			s.stateChangeNotifier()
		}
		return true
	}
	return false
}

// InBetween ...
func (s *CountdownBoard) InBetween() bool {
	return false
}

// Disable ...
func (s *CountdownBoard) Disable() bool {
	if s.config.Enabled.CAS(true, false) {
		if s.stateChangeNotifier != nil {
			s.stateChangeNotifier()
		}
		return true
	}
	return false
}

// SetStateChangeNotifier ...
func (s *CountdownBoard) SetStateChangeNotifier(st board.StateChangeNotifier) {
	s.stateChangeNotifier = st
}

// ScrollMode ...
func (s *CountdownBoard) ScrollMode() bool {
	return s.config.ScrollMode.Load()
}

// HasPriority ...
func (s *CountdownBoard) HasPriority() bool {
	return false
}

// GetHTTPHandlers ...
func (s *CountdownBoard) GetHTTPHandlers() ([]*board.HTTPHandler, error) {
	return nil, nil
}

// GetRPCHandler ...
func (s *CountdownBoard) GetRPCHandler() (string, http.Handler) {
	return s.rpcServer.PathPrefix(), s.rpcServer
}
