package countdownboard

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/rgbrender"
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

// Render ...
func (s *CountdownBoard) render(ctx context.Context, canvas board.Canvas) (board.Canvas, error) {
	s.boardCtx, s.boardCancel = context.WithCancel(ctx)
	zeroed := rgbrender.ZeroedBounds(canvas.Bounds())
	writer, err := s.getWriter(zeroed)
	if err != nil {
		return nil, err
	}

	for _, event := range s.config.Events {
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
			return nil, err
		}
	}

	return nil, nil
}
