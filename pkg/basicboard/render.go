package basicboard

import (
	"context"
	"fmt"
	"time"

	"github.com/robbydyer/sports/pkg/board"
	"github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
)

// ScrollRender ...
func (b *BasicBoard) ScrollRender(ctx context.Context, canvas board.Canvas, padding int) (board.Canvas, error) {
	origScrollMode := b.config.ScrollMode.Load()
	origPad := b.config.TightScrollPadding
	defer func() {
		b.config.ScrollMode.Store(origScrollMode)
		b.config.TightScrollPadding = origPad
	}()

	b.config.ScrollMode.Store(true)
	b.config.TightScrollPadding = padding

	var scrollCanvas *rgbmatrix.ScrollCanvas
	base, ok := canvas.(*rgbmatrix.ScrollCanvas)
	if !ok {
		return nil, fmt.Errorf("invalid scroll canvas")
	}

	var err error
	scrollCanvas, err = rgbmatrix.NewScrollCanvas(base.Matrix, b.log)
	if err != nil {
		return nil, err
	}
	scrollCanvas.SetScrollSpeed(b.config.scrollDelay)
	scrollCanvas.SetScrollDirection(rgbmatrix.RightToLeft)

	renderedCanvases, err := b.renderer.Prepare(ctx, canvas)
	if err != nil {
		return nil, err
	}

	for _, c := range renderedCanvases {
		scrollCanvas.AddCanvas(c)
	}

	scrollCanvas.Merge(b.config.TightScrollPadding)

	return scrollCanvas, nil
}

// Render ...
func (b *BasicBoard) Render(ctx context.Context, canvas board.Canvas) error {
	b.boardCtx, b.boardCancel = context.WithCancel(ctx)
	if canvas.Scrollable() && b.config.ScrollMode.Load() {
		scrollCanvas, err := b.ScrollRender(b.boardCtx, canvas, b.config.TightScrollPadding)
		if err != nil {
			return err
		}
		return scrollCanvas.Render(b.boardCtx)
	}

	renderedCanvases, err := b.renderer.Prepare(b.boardCtx, canvas)
	if err != nil {
		return err
	}

	for _, c := range renderedCanvases {
		select {
		case <-b.boardCtx.Done():
			return context.Canceled
		default:
		}

		if err := c.Render(b.boardCtx); err != nil {
			return err
		}

		select {
		case <-b.boardCtx.Done():
			return context.Canceled
		case <-time.After(b.config.boardDelay):
		}
	}

	return nil
}
