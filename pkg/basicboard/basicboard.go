package basicboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/robbydyer/sports/pkg/basicconfig"
	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/util"

	pb "github.com/robbydyer/sports/internal/proto/basicboard"
)

// BasicBoard implements board.Board
type BasicBoard struct {
	config              *basicconfig.Config
	renderer            Renderer
	log                 *zap.Logger
	rpcServer           pb.TwirpServer
	boardCtx            context.Context
	boardCancel         context.CancelFunc
	stateChangeNotifier board.StateChangeNotifier
}

// Todayer is a func that returns a string representing a date
// that will be used for determining "Today's" games.
// This is useful in testing what past days looked like
type Todayer func() []time.Time

// Renderer ...
type Renderer interface {
	Prepare(ctx context.Context, canvas board.Canvas) ([]board.Canvas, error)
	Config() *basicconfig.Config
	HTTPPathPrefix() string
	RPCHandler() (string, http.Handler)
}

// New ...
func New(renderer Renderer, logger *zap.Logger) (*BasicBoard, error) {
	b := &BasicBoard{
		renderer: renderer,
		log:      logger,
	}

	b.config = renderer.Config()

	b.log.Info("Register Basic Board",
		zap.String("board name", b.Name()),
	)

	if b.config.TodayFunc == nil {
		b.config.TodayFunc = util.Today
	}

	c := cron.New()

	for _, on := range b.config.OnTimes {
		b.log.Info("basicboard will be schedule to turn on",
			zap.String("turn on", on),
			zap.String("name", b.Name()),
		)
		_, err := c.AddFunc(on, func() {
			b.log.Info("basicboard turning on",
				zap.String("name", b.Name()),
			)
			b.Enable()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add cron for sportboard: %w", err)
		}
	}

	for _, off := range b.config.OffTimes {
		b.log.Info("basicboard will be schedule to turn off",
			zap.String("turn on", off),
			zap.String("name", b.Name()),
		)
		_, err := c.AddFunc(off, func() {
			b.log.Info("basicboard turning off",
				zap.String("name", b.Name()),
			)
			b.Disable()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add cron for sportboard: %w", err)
		}
	}

	if _, err := c.AddFunc("0 4 * * *", b.cacheClear); err != nil {
		return nil, err
	}

	c.Start()

	return b, nil
}

func (b *BasicBoard) cacheClear() {
}

// Name ...
func (b *BasicBoard) Name() string {
	return b.renderer.HTTPPathPrefix()
}

// Enabled ...
func (b *BasicBoard) Enabled() bool {
	return b.config.Enabled.Load()
}

// Enable ...
func (b *BasicBoard) Enable() bool {
	if b.config.Enabled.CAS(false, true) {
		if b.stateChangeNotifier != nil {
			b.stateChangeNotifier()
		}
		return true
	}
	return false
}

// InBetween ...
func (s *BasicBoard) InBetween() bool {
	return false
}

// Disable ...
func (b *BasicBoard) Disable() bool {
	if b.config.Enabled.CAS(true, false) {
		if b.stateChangeNotifier != nil {
			b.stateChangeNotifier()
		}
		return true
	}
	return false
}

// SetStateChangeNotifier ...
func (b *BasicBoard) SetStateChangeNotifier(st board.StateChangeNotifier) {
	b.stateChangeNotifier = st
}

// ScrollMode ...
func (b *BasicBoard) ScrollMode() bool {
	return b.config.ScrollMode.Load()
}

// HasPriority ...
func (b *BasicBoard) HasPriority() bool {
	return false
}

// GetHTTPHandlers ...
func (b *BasicBoard) GetHTTPHandlers() ([]*board.HTTPHandler, error) {
	return nil, nil
}

// GetRPCHandler ...
func (b *BasicBoard) GetRPCHandler() (string, http.Handler) {
	return b.renderer.RPCHandler()
}
