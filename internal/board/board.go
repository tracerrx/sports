package board

import (
	"context"
	"image"
	"image/draw"
	"net/http"
	"time"

	scrcnvs "github.com/robbydyer/sports/internal/scrollcanvas"
)

// HTTPHandler is the type returned to the sportsmatrix for HTTP endpoints
type HTTPHandler struct {
	Handler func(http.ResponseWriter, *http.Request)
	Path    string
}

// Enabler is an interface for basic Enable/Disable functions
type Enabler interface {
	Enabled() bool
	Enable() bool
	Disable() bool
	Store(bool) bool
	SetStateChangeCallback(func())
}

// Board is the interface to implement for displaying on the matrix
type Board interface {
	Name() string
	Render(ctx context.Context, canvas Canvas) (Canvas, error)
	// ScrollRender(ctx context.Context, canvas Canvas, padding int) (Canvas, error)
	GetHTTPHandlers() ([]*HTTPHandler, error)
	ScrollMode() bool
	SetScrollMode(bool)
	GetRPCHandler() (string, http.Handler)
	InBetween() bool
	Enabler() Enabler
	SetScrollDelay(time.Duration)
	ScrollDelay() time.Duration
	ScrollDirection() scrcnvs.ScrollDirection
	ScrollPad() int
}

// Canvas ...
type Canvas interface {
	image.Image
	draw.Image
	Enabler
	Name() string
	Clear() error
	Render(ctx context.Context) error
	Close() error
	Scrollable() bool
	AlwaysRender() bool
	SetWidth(int)
	GetWidth() int
}

// StateChangeNotifier is a func that an Enabler uses to notify when its
// enabled/disabled state changes
type StateChangeNotifier func()
