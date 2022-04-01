package forecast

import (
	"time"

	"github.com/robbydyer/sports/pkg/logo"
)

// Forecast ...
type Forecast struct {
	Time         time.Time
	Temperature  *float64
	HighTemp     *float64
	LowTemp      *float64
	Humidity     int
	TempUnit     string
	Icon         *logo.Logo
	IconCode     string
	IsHourly     bool
	PrecipChance *int
}
