package texture

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type NineSlice struct {
	Texture                  *ebiten.Image
	Width                    int
	Height                   int
	Top, Bottom, Left, Right int
	SubImages                [9]*ebiten.Image
}

func NewNineSlice(t *ebiten.Image, top, bottom, left, right int) *NineSlice {
	subImages := [9]*ebiten.Image{}
	bounds := t.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	setSubImage := func(index int, x0, y0, x1, y1 int) {
		if x1 > x0 && y1 > y0 {
			subImages[index] = ebiten.NewImageFromImage(t.SubImage(image.Rect(x0, y0, x1, y1)))
		}
	}

	setSubImage(0, 0, 0, left, top)
	setSubImage(1, left, 0, w-right, top)
	setSubImage(2, w-right, 0, w, top)
	setSubImage(3, 0, top, left, h-bottom)
	setSubImage(4, left, top, w-right, h-bottom)
	setSubImage(5, w-right, top, w, h-bottom)
	setSubImage(6, 0, h-bottom, left, h)
	setSubImage(7, left, h-bottom, w-right, h)
	setSubImage(8, w-right, h-bottom, w, h)

	return &NineSlice{
		Texture:   t,
		Width:     t.Bounds().Dx(),
		Height:    t.Bounds().Dy(),
		Top:       top,
		Bottom:    bottom,
		Left:      left,
		Right:     right,
		SubImages: subImages,
	}
}

func (n *NineSlice) Draw(screen *ebiten.Image, frame image.Rectangle, colorScale *ebiten.ColorScale) {
	if frame.Dx() == 0 || frame.Dy() == 0 {
		return
	}

	bounds := n.Texture.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	sx, sy := n.calculateScale(frame)

	if sx < 0.25 || sy < 0.25 {
		return
	}

	left, right, top, bottom := n.calculateDimensions(sx, sy)

	if n.Top == 0 && n.Bottom == 0 && n.Left == 0 && n.Right == 0 {
		n.drawFullImage(screen, frame, colorScale, w, h)

		return
	}

	n.drawSubImages(screen, frame, colorScale, sx, sy, left, right, top, bottom, w, h)
}

func (n *NineSlice) calculateScale(frame image.Rectangle) (float64, float64) {
	sx := 1.0
	sy := 1.0

	if frame.Dx() <= n.Left+n.Right {
		sx = float64(frame.Dx()) / float64(n.Left+n.Right)
	}

	if frame.Dy() <= n.Top+n.Bottom {
		sy = float64(frame.Dy()) / float64(n.Top+n.Bottom)
	}

	return sx, sy
}

func (n *NineSlice) calculateDimensions(sx, sy float64) (float64, float64, float64, float64) {
	left := float64(n.Left) * sx
	right := float64(n.Right) * sx
	top := float64(n.Top) * sy
	bottom := float64(n.Bottom) * sy

	return left, right, top, bottom
}

func (n *NineSlice) drawFullImage(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	w, h int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(float64(frame.Dx())/float64(w), float64(frame.Dy())/float64(h))
	op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
	screen.DrawImage(n.Texture, op)
}

func (n *NineSlice) drawSubImages(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, left, right, top, bottom float64,
	w, h int,
) {
	if n.SubImages[1] != nil {
		n.drawTop(screen, frame, colorScale, sx, sy, left, w)
	}

	if n.SubImages[5] != nil {
		n.drawRight(screen, frame, colorScale, sx, sy, right, top, h)
	}

	if n.SubImages[4] != nil {
		n.drawCenter(screen, frame, colorScale, sx, sy, left, top, w, h)
	}

	if n.SubImages[0] != nil {
		n.drawTopLeft(screen, frame, colorScale, sx, sy)
	}

	if n.SubImages[2] != nil {
		n.drawTopRight(screen, frame, colorScale, sx, sy, right)
	}

	if n.SubImages[3] != nil {
		n.drawLeft(screen, frame, colorScale, sx, sy, top, h)
	}

	if n.SubImages[7] != nil {
		n.drawBottom(screen, frame, colorScale, sx, sy, left, bottom, w)
	}

	if n.SubImages[6] != nil {
		n.drawBottomLeft(screen, frame, colorScale, sx, sy, bottom)
	}

	if n.SubImages[8] != nil {
		n.drawBottomRight(screen, frame, colorScale, sx, sy, right, bottom)
	}
}

func (n *NineSlice) drawTop(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, left float64, w int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Scale(max(0, float64(frame.Dx()-n.Left-n.Right)/float64(w-n.Left-n.Right)), float64(n.Top)/float64(n.Top))
	op.GeoM.Translate(float64(frame.Min.X)+left, float64(frame.Min.Y))
	screen.DrawImage(n.SubImages[1], op)
}

func (n *NineSlice) drawRight(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, right, top float64, h int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Scale(float64(n.Right)/float64(n.Right), max(0, float64(frame.Dy()-n.Top-n.Bottom)/float64(h-n.Top-n.Bottom)))
	op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Min.Y)+top)
	screen.DrawImage(n.SubImages[5], op)
}

func (n *NineSlice) drawCenter(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, left, top float64,
	w, h int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Scale(
		max(
			0, float64(frame.Dx()-n.Left-n.Right)/float64(w-n.Left-n.Right)),
		max(0, float64(frame.Dy()-n.Top-n.Bottom)/float64(h-n.Top-n.Bottom)),
	)
	op.GeoM.Translate(float64(frame.Min.X)+left, float64(frame.Min.Y)+top)
	screen.DrawImage(n.SubImages[4], op)
}

func (n *NineSlice) drawTopLeft(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy float64,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
	screen.DrawImage(n.SubImages[0], op)
}

func (n *NineSlice) drawTopRight(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, right float64,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Min.Y))
	screen.DrawImage(n.SubImages[2], op)
}

func (n *NineSlice) drawLeft(screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy,
	top float64, h int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Scale(float64(n.Left)/float64(n.Left), max(0, float64(frame.Dy()-n.Top-n.Bottom)/float64(h-n.Top-n.Bottom)))
	op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y)+top)
	screen.DrawImage(n.SubImages[3], op)
}

func (n *NineSlice) drawBottom(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, left, bottom float64,
	w int,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Scale(
		max(0, float64(frame.Dx()-n.Left-n.Right)/float64(w-n.Left-n.Right)),
		float64(n.Bottom)/float64(n.Bottom),
	)
	op.GeoM.Translate(float64(frame.Min.X)+left, float64(frame.Max.Y)-bottom)
	screen.DrawImage(n.SubImages[7], op)
}

func (n *NineSlice) drawBottomLeft(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, bottom float64,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(float64(frame.Min.X), float64(frame.Max.Y)-bottom)
	screen.DrawImage(n.SubImages[6], op)
}

func (n *NineSlice) drawBottomRight(
	screen *ebiten.Image,
	frame image.Rectangle,
	colorScale *ebiten.ColorScale,
	sx, sy, right, bottom float64,
) {
	op := &ebiten.DrawImageOptions{}
	if colorScale != nil {
		op.ColorScale.ScaleWithColorScale(*colorScale)
	}

	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Max.Y)-bottom)
	screen.DrawImage(n.SubImages[8], op)
}
