package weather

import (
	"context"
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var defaultInterval = 1 * time.Hour

type Option func(*Weather) error

type Weather struct {
	config             *Config
	log                *zap.Logger
	data               *WeatherData
	dataGetter         DataGetter
	lastUpdate         time.Time
	temperatureWriters map[string]*rgbrender.TextWriter
	icons              map[string]image.Image
	sync.Mutex
}

type Config struct {
	updateInterval time.Duration
	boardDelay     time.Duration
	Enabled        *atomic.Bool `json:"enabled"`
	BoardDelay     string       `json:"boardDelay"`
	UpdateInterval string       `json:"updateInterval"`
	Latitude       float64      `json:"latitude"`
	Longitude      float64      `json:"longitutde"`
	APIKey         string       `json:"apiKey"`
}

func (c *Config) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = atomic.NewBool(false)
	}

	if c.UpdateInterval != "" {
		var err error
		c.updateInterval, err = time.ParseDuration(c.UpdateInterval)
		if err != nil {
			c.updateInterval = defaultInterval
		}
	} else {
		c.updateInterval = defaultInterval
	}

	if c.BoardDelay != "" {
		var err error
		c.boardDelay, err = time.ParseDuration(c.BoardDelay)
		if err != nil {
			c.boardDelay = 10 * time.Second
		}
	} else {
		c.boardDelay = 10 * time.Second
	}
}

// New ...
func New(cfg *Config, logger *zap.Logger, opts ...Option) (*Weather, error) {
	w := &Weather{
		config:             cfg,
		log:                logger,
		temperatureWriters: make(map[string]*rgbrender.TextWriter),
		icons:              make(map[string]image.Image),
	}

	for _, f := range opts {
		if err := f(w); err != nil {
			return nil, err
		}
	}

	if w.dataGetter == nil {
		w.dataGetter = GetWeather
	}

	return w, nil
}

func (w *Weather) Name() string {
	return "Weather"
}
func (w *Weather) Render(ctx context.Context, canvas board.Canvas) error {
	if !w.config.Enabled.Load() {
		return nil
	}

	if w.dataGetter == nil {
		return fmt.Errorf("dataGetter not set")
	}

	if time.Since(w.lastUpdate) > w.config.updateInterval || w.data == nil {
		w.log.Info("updating weather")
		dat, err := w.dataGetter(ctx, w.config.Latitude, w.config.Longitude, w.config.APIKey)
		if err != nil {
			return fmt.Errorf("failed to get weather update: %w", err)
		}
		w.data = dat
		w.lastUpdate = time.Now()
	}

	if err := w.renderCurrent(ctx, canvas); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case <-time.After(w.config.boardDelay):
	}

	return nil
}
func (w *Weather) GetHTTPHandlers() ([]*board.HTTPHandler, error) {
	return []*board.HTTPHandler{}, nil
}
func (w *Weather) Enabled() bool {
	return w.config.Enabled.Load()
}
func (w *Weather) Enable() {
	w.config.Enabled.Store(true)
}
func (w *Weather) Disable() {
	w.config.Enabled.Store(false)
}

func WithWeatherDataGetter(d DataGetter) Option {
	return func(w *Weather) error {
		w.dataGetter = d
		return nil
	}
}
