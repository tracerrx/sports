package weather

import (
	"bytes"
	"context"

	// embed
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"
	"time"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/logo"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"go.uber.org/zap"
)

//go:embed assets/degree.png
var degree []byte

func (w *Weather) renderCurrent(ctx context.Context, canvas board.Canvas) error {
	tempWriter, err := w.temperatureWriter(canvas.Bounds())
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		icon, err := w.currentIcon(ctx, canvas.Bounds())
		if err != nil {
			w.log.Error("failed to get current weather icon", zap.Error(err))
			return
		}

		draw.Draw(canvas, canvas.Bounds(), icon, image.Point{}, draw.Over)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		icon, err := w.degreeIcon(canvas.Bounds())
		if err != nil {
			w.log.Error("failed to get degree icon", zap.Error(err))
			return
		}

		draw.Draw(canvas, canvas.Bounds(), icon, image.Point{}, draw.Over)
	}()

	tempStr := fmt.Sprintf("%.0f", w.data.Current.Temp)

	w.log.Debug("current weather",
		zap.String("temperature", tempStr),
	)

	tempWriter.WriteAligned(
		rgbrender.LeftTop,
		canvas,
		canvas.Bounds(),
		[]string{
			tempStr,
		},
		color.White,
	)

	waiter := make(chan struct{})

	go func() {
		defer close(waiter)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return context.Canceled
	case <-waiter:
	case <-time.After(60 * time.Second):
		return fmt.Errorf("timed out waiting for weather")
	}

	return canvas.Render()
}

func (w *Weather) currentIcon(ctx context.Context, bounds image.Rectangle) (image.Image, error) {
	w.Lock()
	defer w.Unlock()

	for _, i := range w.data.Current.Weather {
		key := fmt.Sprintf("%s_%dx%d", i.Icon, bounds.Dx(), bounds.Dy())

		if img, ok := w.icons[key]; ok {
			w.log.Debug("using cached icon", zap.String("key", key))
			return img, nil
		}

		w.log.Info("fetching weather icon", zap.String("key", key), zap.String("icon", i.Icon))
		img, err := getIcon(ctx, i.Icon)
		if err != nil {
			return nil, err
		}
		c := logo.DefaultConfig()
		c.Abbrev = key
		c.Pt.Zoom = 1
		l := logo.New(key, img, "/tmp", bounds, c)

		rendered, err := l.RenderRightAligned(bounds, bounds.Dy()/2)
		if err != nil {
			return nil, err
		}

		w.icons[key] = rendered

		return rendered, nil
	}

	return nil, fmt.Errorf("no icon for weather")
}

func (w *Weather) degreeIcon(bounds image.Rectangle) (image.Image, error) {
	key := fmt.Sprintf("degree_%dx%d", bounds.Dx(), bounds.Dy())

	w.Lock()
	defer w.Unlock()

	if l, ok := w.icons[key]; ok {
		return l, nil
	}

	buf := bytes.NewBuffer(degree)
	img, err := png.Decode(buf)
	if err != nil {
		return nil, err
	}

	sizedBounds := image.Rect(0, 0, bounds.Dx()/8, bounds.Dx()/8)

	sized := rgbrender.ResizeImage(img, sizedBounds, 1)

	w.icons[key] = sized

	return sized, nil
}
