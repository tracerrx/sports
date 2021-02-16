package tv

import (
	"image"
	"image/color"

	"github.com/jphsd/glui"
)

// TV implements image.Image and draw.Image
type TV struct {
	width  int
	height int
	pixels []uint32
}

func New(width int, height int) *TV {
	return &TV{
		width:  width,
		height: height,
		pixels: make([]uint32, (width * height)),
	}
}
func (t *TV) position(x, y int) int {
	return x + (y * t.width)
}

func (t *TV) ColorModel() color.Model {
	return color.RGBAModel
}
func (t *TV) Bounds() image.Rectangle {
	return image.Rect(0, 0, t.width, t.height)
}
func (t *TV) At(x, y int) color.Color {
	return uint32ToColor(t.pixels[t.position(x, y)])
}

func (t *TV) Set(x, y int, c color.Color) {
	t.pixels[t.position(x, y)] = colorToUint32(c)
}

func (t *TV) Clear() error {
	return nil
}

func (t *TV) Render() error {
	_ = glui.NewGLWin(t.width, t.height, "", t, true)

	glui.Loop(nil)
	return nil
}

func colorToUint32(c color.Color) uint32 {
	if c == nil {
		return 0
	}

	// A color's RGBA method returns values in the range [0, 65535]
	red, green, blue, _ := c.RGBA()
	return (red>>8)<<16 | (green>>8)<<8 | blue>>8
}

func uint32ToColor(u uint32) color.Color {
	return color.RGBA{
		uint8(u>>16) & 255,
		uint8(u>>8) & 255,
		uint8(u>>0) & 255,
		0,
	}
}
