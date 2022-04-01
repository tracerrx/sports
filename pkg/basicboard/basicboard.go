package basicboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/robbydyer/sports/pkg/basicconfig"
	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/util"
)

// BasicBoard implements board.Board
type BasicBoard struct {
	config              *basicconfig.Config
	renderer            Renderer
	log                 *zap.Logger
	boardCtx            context.Context
	boardCancel         context.CancelFunc
	stateChangeNotifier board.StateChangeNotifier
}

// Renderer ...
type Renderer interface {
	Prepare(ctx context.Context, canvas board.Canvas) ([]board.Canvas, error)
	Config() *basicconfig.Config
	HTTPPathPrefix() string
	RPCHandler() (string, http.Handler)
	SetBoard(*BasicBoard)
	InBetween() bool
}

// New ...
func New(renderer Renderer, logger *zap.Logger) (*BasicBoard, error) {
	b := &BasicBoard{
		renderer: renderer,
		log:      logger,
	}

	renderer.SetBoard(b)

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
func (b *BasicBoard) InBetween() bool {
	return b.renderer.InBetween()
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

// Cancel ...
func (b *BasicBoard) Cancel() {
	if b.boardCancel != nil {
		b.log.Info("canceling board",
			zap.String("board", b.Name()),
		)
		b.boardCancel()
		if b.stateChangeNotifier != nil {
			b.stateChangeNotifier()
		}
	}
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
