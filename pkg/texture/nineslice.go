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

	if left > 0 && top > 0 {
		subImages[0] = ebiten.NewImageFromImage(t.SubImage(image.Rect(0, 0, left, top)))
	}

	if w-right > left && top > 0 {
		subImages[1] = ebiten.NewImageFromImage(t.SubImage(image.Rect(left, 0, w-right, top)))
	}

	if w > w-right && top > 0 {
		subImages[2] = ebiten.NewImageFromImage(t.SubImage(image.Rect(w-right, 0, w, top)))
	}

	if left > 0 && h-bottom > top {
		subImages[3] = ebiten.NewImageFromImage(t.SubImage(image.Rect(0, top, left, h-bottom)))
	}

	if w-right > left && h-bottom > top {
		subImages[4] = ebiten.NewImageFromImage(t.SubImage(image.Rect(left, top, w-right, h-bottom)))
	}

	if w > w-right && h-bottom > top {
		subImages[5] = ebiten.NewImageFromImage(t.SubImage(image.Rect(w-right, top, w, h-bottom)))
	}

	if left > 0 && h > h-bottom {
		subImages[6] = ebiten.NewImageFromImage(t.SubImage(image.Rect(0, h-bottom, left, h)))
	}

	if w-right > left && h > h-bottom {
		subImages[7] = ebiten.NewImageFromImage(t.SubImage(image.Rect(left, h-bottom, w-right, h)))
	}

	if w > w-right && h > h-bottom {
		subImages[8] = ebiten.NewImageFromImage(t.SubImage(image.Rect(w-right, h-bottom, w, h)))
	}

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
	sx := 1.0
	sy := 1.0

	if frame.Dx() <= n.Left+n.Right {
		sx = float64(frame.Dx()) / float64(n.Left+n.Right)
	}

	if frame.Dy() <= n.Top+n.Bottom {
		sy = float64(frame.Dy()) / float64(n.Top+n.Bottom)
	}

	if sx < 0.25 || sy < 0.25 {
		return
	}

	left := float64(n.Left) * sx
	right := float64(n.Right) * sx
	top := float64(n.Top) * sy
	bottom := float64(n.Bottom) * sy

	if n.Top == 0 && n.Bottom == 0 && n.Left == 0 && n.Right == 0 {
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(float64(frame.Dx())/float64(w), float64(frame.Dy())/float64(h))
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
		screen.DrawImage(n.Texture, op)

		return
	}

	if n.SubImages[1] != nil {
		// then draw the top
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Scale(max(0, float64(frame.Dx()-n.Left-n.Right)/float64(w-n.Left-n.Right)), float64(n.Top)/float64(n.Top))
		op.GeoM.Translate(float64(frame.Min.X)+left, float64(frame.Min.Y))
		screen.DrawImage(n.SubImages[1], op)
	}

	if n.SubImages[5] != nil {
		// then draw the right
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Scale(float64(n.Right)/float64(n.Right), max(0, float64(frame.Dy()-n.Top-n.Bottom)/float64(h-n.Top-n.Bottom)))
		op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Min.Y)+top)
		screen.DrawImage(n.SubImages[5], op)
	}

	if n.SubImages[4] != nil {
		// first draw the center
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

	if n.SubImages[0] != nil {
		// then draw the top left
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
		screen.DrawImage(n.SubImages[0], op)
	}

	if n.SubImages[2] != nil {
		// then draw the top right
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Min.Y))
		screen.DrawImage(n.SubImages[2], op)
	}

	if n.SubImages[3] != nil {
		// then draw the left
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Scale(float64(n.Left)/float64(n.Left), max(0, float64(frame.Dy()-n.Top-n.Bottom)/float64(h-n.Top-n.Bottom)))
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y)+top)
		screen.DrawImage(n.SubImages[3], op)
	}

	if n.SubImages[7] != nil {
		// then draw the bottom
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

	if n.SubImages[6] != nil {
		// then draw the bottom left
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Max.Y)-bottom)
		screen.DrawImage(n.SubImages[6], op)
	}

	if n.SubImages[8] != nil {
		// then draw the bottom right
		op := &ebiten.DrawImageOptions{}
		if colorScale != nil {
			op.ColorScale.ScaleWithColorScale(*colorScale)
		}

		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(float64(frame.Max.X)-right, float64(frame.Max.Y)-bottom)
		screen.DrawImage(n.SubImages[8], op)
	}
}
