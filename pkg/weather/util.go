package weather

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"

	"github.com/robbydyer/sports/pkg/rgbrender"
)

func (w *Weather) temperatureWriter(bounds image.Rectangle) (*rgbrender.TextWriter, error) {
	k := fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy())
	if writer, ok := w.temperatureWriters[k]; ok {
		return writer, nil
	}

	fnt, err := rgbrender.GetFont("score.ttf")
	if err != nil {
		return nil, err
	}

	size := 0.25 * float64(bounds.Dx())

	writer := rgbrender.NewTextWriter(fnt, size)

	w.Lock()
	defer w.Unlock()
	w.temperatureWriters[k] = writer

	return writer, nil
}

func getIcon(ctx context.Context, base string) (image.Image, error) {
	uri := fmt.Sprintf("http://openweathermap.org/img/wn/%s@4x.png", base)

	client := http.DefaultClient

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return png.Decode(resp.Body)
}
