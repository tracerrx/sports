package weather

import (
	"context"
	"image"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/twitchtv/twirp"

	"github.com/robbydyer/sports/internal/proto/basicboard"
	pb "github.com/robbydyer/sports/internal/proto/weatherboard"
	"github.com/robbydyer/sports/pkg/basicconfig"
	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/forecast"
	"github.com/robbydyer/sports/pkg/logo"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"github.com/robbydyer/sports/pkg/twirphelpers"
)

// Weather displays weather
type Weather struct {
	board               *basicboard.BasicBoard
	config              *Config
	api                 API
	log                 *zap.Logger
	enablerLock         sync.Mutex
	iconLock            sync.Mutex
	iconCache           map[string]*logo.Logo
	cancelBoard         chan struct{}
	bigWriter           *rgbrender.TextWriter
	smallWriter         *rgbrender.TextWriter
	rpcServer           pb.TwirpServer
	stateChangeNotifier board.StateChangeNotifier
	sync.Mutex
}

// Config for a Weather
type Config struct {
	basicconfig.Config
	ZipCode         string       `json:"zipCode"`
	Country         string       `json:"country"`
	APIKey          string       `json:"apiKey"`
	CurrentForecast *atomic.Bool `json:"currentForecast"`
	HourlyForecast  *atomic.Bool `json:"hourlyForecast"`
	DailyForecast   *atomic.Bool `json:"dailyForecast"`
	DailyNumber     int          `json:"dailyNumber"`
	HourlyNumber    int          `json:"hourlyNumber"`
	MetricUnits     *atomic.Bool `json:"metricUnits"`
	ShowBetween     *atomic.Bool `json:"showBetween"`
}

// API interface for getting weather data
type API interface {
	CurrentForecast(ctx context.Context, zipCode string, country string, bounds image.Rectangle, metricUnits bool) (*forecast.Forecast, error)
	DailyForecasts(ctx context.Context, zipCode string, country string, bounds image.Rectangle, metricUnits bool) ([]*forecast.Forecast, error)
	HourlyForecasts(ctx context.Context, zipCode string, country string, bounds image.Rectangle, metricUnits bool) ([]*forecast.Forecast, error)
	CacheClear()
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Config.SetDefaults()
	if c.CurrentForecast == nil {
		c.CurrentForecast = atomic.NewBool(false)
	}
	if c.HourlyForecast == nil {
		c.HourlyForecast = atomic.NewBool(false)
	}
	if c.DailyForecast == nil {
		c.DailyForecast = atomic.NewBool(false)
	}
	if c.MetricUnits == nil {
		c.MetricUnits = atomic.NewBool(false)
	}
	if c.ShowBetween == nil {
		c.ShowBetween = atomic.NewBool(false)
	}
	if c.DailyNumber == 0 {
		c.DailyNumber = 3
	}

	if c.HourlyNumber == 0 {
		c.HourlyNumber = 3
	}
}

// New ...
func New(api API, config *Config, log *zap.Logger) (*Weather, error) {
	w := &Weather{
		config:      config,
		api:         api,
		log:         log,
		cancelBoard: make(chan struct{}),
		iconCache:   make(map[string]*logo.Logo),
	}

	svr := &Server{
		board: w,
	}
	w.rpcServer = pb.NewWeatherBoardServer(svr,
		twirp.WithServerPathPrefix(""),
		twirp.ChainHooks(
			twirphelpers.GetDefaultHooks(w.HTTPPathPrefix(), w.log),
		),
	)

	return w, nil
}

func (w *Weather) cacheClear() {
	w.api.CacheClear()
}

// Config ...
func (w *Weather) Config() *basicconfig.Config {
	return &w.config.Config
}

// HTTPPathPrefix ...
func (w *Weather) HTTPPathPrefix() string {
	return "weather"
}

// SetBoard ...
func (w *Weather) SetBoard(b *basicboard.BasicBoard) {
	w.board = b
}

// InBetween ...
func (w *Weather) InBetween() bool {
	return w.config.ShowBetween.Load()
}

// Prepare ...
func (w *Weather) Prepare(ctx context.Context, canvas board.Canvas) ([]board.Canvas, error) {
	zeroed := rgbrender.ZeroedBounds(canvas.Bounds())
	forecasts := []*forecast.Forecast{}
	if w.config.CurrentForecast.Load() {
		f, err := w.api.CurrentForecast(ctx, w.config.ZipCode, w.config.Country, zeroed, w.config.MetricUnits.Load())
		if err != nil {
			return nil, err
		}
		forecasts = append(forecasts, f)
	}
	if w.config.HourlyForecast.Load() {
		fs, err := w.api.HourlyForecasts(ctx, w.config.ZipCode, w.config.Country, zeroed, w.config.MetricUnits.Load())
		if err != nil {
			return nil, err
		}
		// sortForecasts(fs)
		w.log.Debug("found hourly forecasts",
			zap.Int("num", len(fs)),
			zap.Int("max show", w.config.HourlyNumber),
		)
		if len(fs) > 0 {
		HOURLY:
			for i := 0; i < w.config.HourlyNumber; i++ {
				if len(fs) <= i {
					break HOURLY
				}
				forecasts = append(forecasts, fs[i])
			}
		}
	}

	if w.config.DailyForecast.Load() {
		fs, err := w.api.DailyForecasts(ctx, w.config.ZipCode, w.config.Country, zeroed, w.config.MetricUnits.Load())
		if err != nil {
			return nil, err
		}
		w.log.Debug("found daily forecasts",
			zap.Int("num", len(fs)),
			zap.Int("max show", w.config.DailyNumber),
		)

		// Drop today's forecast, as it's redundant
	TODAYCHECK:
		for i := range fs {
			if fs[i].Time.YearDay() == time.Now().Local().YearDay() {
				// delete this element
				fs = append(fs[:i], fs[i+1:]...)
				break TODAYCHECK
			}
		}
		if len(fs) > 0 {
		DAILY:
			for i := 0; i < w.config.DailyNumber; i++ {
				if len(fs) <= i {
					break DAILY
				}
				forecasts = append(forecasts, fs[i])
			}
		}
	}

	canvases := []board.Canvas{}

FORECASTS:
	for _, f := range forecasts {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
		}

		thisCanvas := canvas.Clone()

		if err := w.drawForecast(ctx, thisCanvas, f); err != nil {
			w.log.Error("failed to render forecast",
				zap.Error(err),
			)
			continue FORECASTS
		}

		canvases = append(canvases, thisCanvas)
	}

	return canvases, nil
}
