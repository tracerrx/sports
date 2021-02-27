package weather

import (
	"context"
	// embed
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

//go:embed assets/weather.json
var exampleData []byte

type DataGetter func(ctx context.Context, lat float64, lon float64, apiKey string) (*WeatherData, error)

type WeatherData struct {
	Lat      float64          `json:"lat"`
	Lon      float64          `json:"lon"`
	Timezone string           `json:"timezone"`
	Current  *weatherPeriod   `json:"current"`
	Hourly   []*weatherPeriod `json:"hourly"`
	Daily    []struct {
		Dt      int64             `json:"dt"`
		Sunrise int64             `json:"sunrise"`
		Sunset  int64             `json:"sunset"`
		Clouds  int               `json:"clouds"`
		Weather []*currentWeather `json:"weather"`
		Temp    struct {
			Day   float64 `json:"day"`
			Min   float64 `json:"min"`
			Max   float64 `json:"max"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"temp"`
		FeelsLike struct {
			Day   float64 `json:"day"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"feels_like"`
	}
	Alerts []struct {
		Event string `json:"event"`
		Start int64  `json:"start"`
		End   int64  `json:"end"`
	}
}

type weatherPeriod struct {
	Dt           int64             `json:"dt"`
	Sunrise      int64             `json:"sunrise,omitempty"`
	Sunset       int64             `json:"sunset,omitempty"`
	Temp         float64           `json:"temp"`
	FeelsLike    float64           `json:"feels_like"`
	Humidity     int               `json:"humidity"`
	Clouds       int               `json:"clouds"`
	Weather      []*currentWeather `json:"weather"`
	PrecipChance float64           `json:"pop"`
}

type currentWeather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func GetWeather(ctx context.Context, lat float64, lon float64, apiKey string) (*WeatherData, error) {
	uri, err := url.Parse("https://api.openweathermap.org/data/2.5/onecall")
	if err != nil {
		return nil, err
	}

	latStr := strconv.FormatFloat(lat, 'E', -1, 64)
	lonStr := strconv.FormatFloat(lon, 'E', -1, 64)

	v := uri.Query()
	v.Set("lat", latStr)
	v.Set("lon", lonStr)
	v.Set("appid", apiKey)
	v.Set("units", "imperial")
	v.Set("exclude", "minutely")

	uri.RawQuery = v.Encode()

	client := http.DefaultClient

	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dat *WeatherData

	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}

	return dat, nil
}

func GetWeatherFromAsset(ctx context.Context, lat float64, lon float64, apiKey string) (*WeatherData, error) {
	var dat *WeatherData

	if err := json.Unmarshal(exampleData, &dat); err != nil {
		return nil, err
	}

	return dat, nil
}
