package board

import (
	"context"
	"fmt"

	scrcnvs "github.com/robbydyer/sports/internal/scrollcanvas"
)

func Render(ctx context.Context, b Board, canvas Canvas) error {
	c, err := b.Render(ctx, canvas)
	if err != nil {
		return err
	}
	if c == nil {
		// Rendering was managed by the board itself
		return nil
	}

	scr, ok := c.(*scrcnvs.ScrollCanvas)
	if !ok {
		// non-scroll canvas
		return c.Render(ctx)
	}

	origSpd := scr.GetScrollSpeed()
	origDir := scr.GetScrollDirection()

	defer func() {
		b.SetScrollDelay(scr.GetScrollSpeed())
		scr.SetScrollSpeed(origSpd)
		scr.SetScrollDirection(origDir)
	}()

	scr.SetScrollDirection(b.ScrollDirection())
	scr.SetScrollSpeed(b.ScrollDelay())
	scr.MergePad = b.ScrollPad()

	if err := scr.Render(ctx); err != nil {
		return err
	}

	return nil
}

// GetScroll preps a Board in scroll mode, but just returns the prepared ScrollCanvas. It
// does not perform a canvas Render
func GetScroll(ctx context.Context, b Board, canvas Canvas) (*scrcnvs.ScrollCanvas, error) {
	origScrMode := b.ScrollMode()
	defer func() {
		b.SetScrollMode(origScrMode)
	}()

	b.SetScrollMode(true)

	c, err := b.Render(ctx, canvas)
	if err != nil {
		return nil, err
	}

	scr, ok := c.(*scrcnvs.ScrollCanvas)
	if !ok {
		return nil, fmt.Errorf("unexpected canvas type returned from board: %T", scr)
	}

	return scr, nil
}
