package racing

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	pb "github.com/robbydyer/sports/internal/proto/racingboard"
	"github.com/robbydyer/sports/pkg/basicboard"
	"github.com/robbydyer/sports/pkg/basicconfig"
	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/logo"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"github.com/robbydyer/sports/pkg/twirphelpers"
	"github.com/twitchtv/twirp"
)

// Racing ...
type Racing struct {
	api            API
	log            *zap.Logger
	rpcServer      pb.TwirpServer
	config         *Config
	leagueLogo     *logo.Logo
	events         []*Event
	scheduleWriter *rgbrender.TextWriter
	board          *basicboard.BasicBoard
}

// Config ...
type Config struct {
	basicconfig.Config
}

// API ...
type API interface {
	LeagueShortName() string
	GetLogo(ctx context.Context, bounds image.Rectangle) (*logo.Logo, error)
	GetScheduledEvents(ctx context.Context) ([]*Event, error)
	HTTPPathPrefix() string
}

// Event ...
type Event struct {
	Date time.Time
	Name string
}

// New ...
func New(api API, config *Config, logger *zap.Logger) (*Racing, error) {
	r := &Racing{
		api:    api,
		log:    logger,
		config: config,
	}

	svr := &Server{
		board: r,
	}
	prfx := r.api.HTTPPathPrefix()
	if !strings.HasPrefix(prfx, "/") {
		prfx = fmt.Sprintf("/%s", prfx)
	}
	r.rpcServer = pb.NewRacingServer(svr,
		twirp.WithServerPathPrefix(prfx),
		twirp.ChainHooks(
			twirphelpers.GetDefaultHooks(r.api.HTTPPathPrefix(), r.log),
		),
	)

	return r, nil
}

// SetBoard ...
func (r *Racing) SetBoard(b *basicboard.BasicBoard) {
	r.board = b
}

// Config ...
func (r *Racing) Config() *basicconfig.Config {
	return &r.config.Config
}

// HTTPPathPrefix ...
func (r *Racing) HTTPPathPrefix() string {
	return r.api.LeagueShortName()
}

// RPCHandler ...
func (r *Racing) RPCHandler() (string, http.Handler) {
	return r.rpcServer.PathPrefix(), r.rpcServer
}

// Prepare ...
func (r *Racing) Prepare(ctx context.Context, canvas board.Canvas) ([]board.Canvas, error) {
	if r.leagueLogo == nil {
		var err error
		r.leagueLogo, err = r.api.GetLogo(ctx, canvas.Bounds())
		if err != nil {
			return nil, err
		}
	}

	if len(r.events) < 1 {
		var err error
		r.events, err = r.api.GetScheduledEvents(ctx)
		if err != nil {
			return nil, err
		}
	}

	r.log.Debug("preparing racing events",
		zap.String("league", r.api.LeagueShortName()),
		zap.Int("num events", len(r.events)),
	)

	scheduleWriter, err := r.getScheduleWriter(rgbrender.ZeroedBounds(canvas.Bounds()))
	if err != nil {
		return nil, err
	}

	canvases := []board.Canvas{}

EVENTS:
	for _, event := range r.events {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
		}

		thisCanvas := canvas.Clone()

		if err := r.renderEvent(ctx, thisCanvas, event, r.leagueLogo, scheduleWriter); err != nil {
			r.log.Error("failed to render racing event",
				zap.Error(err),
			)
			continue EVENTS
		}

		canvases = append(canvases, thisCanvas)
	}

	return canvases, nil
}

func (r *Racing) renderEvent(ctx context.Context, canvas board.Canvas, event *Event, leagueLogo *logo.Logo, scheduleWriter *rgbrender.TextWriter) error {
	canvasBounds := rgbrender.ZeroedBounds(canvas.Bounds())

	logoImg, err := leagueLogo.RenderRightAlignedWithEnd(ctx, canvasBounds, (canvasBounds.Max.X-canvasBounds.Min.X)/2)
	if err != nil {
		return err
	}

	pt := image.Pt(logoImg.Bounds().Min.X, logoImg.Bounds().Min.Y)
	draw.Draw(canvas, logoImg.Bounds(), logoImg, pt, draw.Over)

	gradient := rgbrender.GradientXRectangle(
		canvasBounds,
		0.1,
		color.Black,
		r.log,
	)
	pt = image.Pt(gradient.Bounds().Min.X, gradient.Bounds().Min.Y)
	draw.Draw(canvas, gradient.Bounds(), gradient, pt, draw.Over)

	event.Date = event.Date.Local()

	tz, _ := event.Date.Zone()
	txt := []string{
		event.Name,
		event.Date.Format("01/02/2006"),
		fmt.Sprintf("%s %s", event.Date.Format("3:04PM"), tz),
	}

	lengths, err := scheduleWriter.MeasureStrings(canvas, txt)
	if err != nil {
		return err
	}
	max := canvasBounds.Dx() / 2

	for _, length := range lengths {
		if length > max {
			max = length
		}
	}

	r.log.Debug("max racing schedule text length",
		zap.Int("max", max),
		zap.Int("half bounds", canvasBounds.Dy()/2),
	)

	scheduleBounds := image.Rect(
		canvasBounds.Max.X/2,
		canvasBounds.Min.Y,
		(canvasBounds.Max.X/2)+max,
		canvasBounds.Max.Y,
	)

	if err := scheduleWriter.WriteAligned(
		rgbrender.LeftCenter,
		canvas,
		scheduleBounds,
		txt,
		color.White,
	); err != nil {
		return fmt.Errorf("failed to write schedule: %w", err)
	}

	return nil
}

func (r *Racing) getScheduleWriter(bounds image.Rectangle) (*rgbrender.TextWriter, error) {
	if r.scheduleWriter != nil {
		return r.scheduleWriter, nil
	}

	var err error
	r.scheduleWriter, err = rgbrender.DefaultTextWriter()
	if err != nil {
		return nil, err
	}

	if bounds.Dy() <= 256 {
		r.scheduleWriter.FontSize = 8.0
	} else {
		r.scheduleWriter.FontSize = 0.25 * float64(bounds.Dy())
	}

	if bounds.Dy() <= 256 {
		r.scheduleWriter.YStartCorrection = -2
	} else {
		r.scheduleWriter.YStartCorrection = -1 * ((bounds.Dy() / 32) + 1)
	}

	return r.scheduleWriter, nil
}
