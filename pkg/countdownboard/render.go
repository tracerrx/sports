package countdownboard

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
	"github.com/robbydyer/sports/pkg/rgbrender"
	"go.uber.org/zap"
)

// ScrollRender ...
func (s *CountdownBoard) ScrollRender(ctx context.Context, canvas board.Canvas, padding int) (board.Canvas, error) {
	origScrollMode := s.config.ScrollMode.Load()
	origPad := s.config.TightScrollPadding
	defer func() {
		s.config.ScrollMode.Store(origScrollMode)
		s.config.TightScrollPadding = origPad
	}()

	s.config.ScrollMode.Store(true)
	s.config.TightScrollPadding = padding

	return s.render(ctx, canvas)
}

// Render ...
func (s *CountdownBoard) Render(ctx context.Context, canvas board.Canvas) error {
	c, err := s.render(ctx, canvas)
	if err != nil {
		return err
	}
	if c != nil {
		return c.Render(ctx)
	}

	return nil
}

func (s *CountdownBoard) render(ctx context.Context, canvas board.Canvas) (board.Canvas, error) {
	s.boardCtx, s.boardCancel = context.WithCancel(ctx)
	zeroed := rgbrender.ZeroedBounds(canvas.Bounds())
	writer, err := s.getWriter(zeroed)
	if err != nil {
		return nil, err
	}

	var scrollCanvas *rgbmatrix.ScrollCanvas
	if canvas.Scrollable() && s.config.ScrollMode.Load() {
		base, ok := canvas.(*rgbmatrix.ScrollCanvas)
		if !ok {
			return nil, fmt.Errorf("invalid scroll canvas")
		}

		var err error
		scrollCanvas, err = rgbmatrix.NewScrollCanvas(base.Matrix, s.log)
		if err != nil {
			return nil, err
		}
		scrollCanvas.SetScrollSpeed(s.config.scrollDelay)
		scrollCanvas.SetScrollDirection(rgbmatrix.RightToLeft)
	}

EVENTS:
	for _, event := range s.config.Events {
		select {
		case <-s.boardCtx.Done():
			return nil, context.Canceled
		default:
		}
		if err := s.renderEvent(s.boardCtx, canvas, event, writer); err != nil {
			continue EVENTS
		}

		if scrollCanvas != nil && s.config.ScrollMode.Load() {
			scrollCanvas.AddCanvas(canvas)
			draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Over)
			continue EVENTS
		}

		if err := canvas.Render(s.boardCtx); err != nil {
			s.log.Error("failed to render countdown board",
				zap.Error(err),
			)
			continue EVENTS
		}

		if !s.config.ScrollMode.Load() {
			select {
			case <-ctx.Done():
				return nil, context.Canceled
			case <-time.After(s.config.boardDelay):
			}
		}
	}

	if canvas.Scrollable() && scrollCanvas != nil {
		scrollCanvas.Merge(s.config.TightScrollPadding)
		return scrollCanvas, nil
	}

	return nil, nil
}

func (s *CountdownBoard) renderEvent(ctx context.Context, canvas board.Canvas, event *Event, writer *rgbrender.TextWriter) error {
	zeroed := rgbrender.ZeroedBounds(canvas.Bounds())
	daysLeft := int(math.Ceil(time.Until(event.date).Hours() / 24.0))
	if err := writer.WriteAligned(
		rgbrender.CenterCenter,
		canvas,
		zeroed,
		[]string{
			fmt.Sprintf("Days until %s", event.Title),
			fmt.Sprintf("%d", daysLeft),
		},
		color.White,
	); err != nil {
		return err
	}

	return nil
}
