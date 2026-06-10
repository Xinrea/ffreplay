package ui

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

// toEUINineSlice converts our custom texture.NineSlice to ebitenui's NineSlice,
// preserving the nine-slice border dimensions so game textures stretch identically.
func toEUINineSlice(ns *texture.NineSlice) *euiimage.NineSlice {
	w := ns.Width
	h := ns.Height
	return euiimage.NewNineSlice(
		ns.Texture,
		[3]int{ns.Left, w - ns.Left - ns.Right, ns.Right},
		[3]int{ns.Top, h - ns.Top - ns.Bottom, ns.Bottom},
	)
}

// scaleImage returns a new *ebiten.Image with src drawn at the target dimensions.
// Used to produce DPI-correct versions of fixed-size game sprites.
func scaleImage(src *ebiten.Image, w, h int) *ebiten.Image {
	if w <= 0 || h <= 0 {
		return src
	}
	dst := ebiten.NewImage(w, h)
	sb := src.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(sb.Dx()), float64(h)/float64(sb.Dy()))
	dst.DrawImage(src, op)
	return dst
}

// overlayImage draws src scaled to fill the target dst image (top-left origin).
func overlayImage(dst *ebiten.Image, src *ebiten.Image) {
	db := dst.Bounds()
	sb := src.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(db.Dx())/float64(sb.Dx()), float64(db.Dy())/float64(sb.Dy()))
	dst.DrawImage(src, op)
}

// imageWithAlpha returns a copy of src with alpha multiplied by alpha.
func imageWithAlpha(src *ebiten.Image, alpha float32) *ebiten.Image {
	dst := ebiten.NewImage(src.Bounds().Dx(), src.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	dst.DrawImage(src, op)
	return dst
}

// nineSliceWithAlpha converts the custom nine-slice and applies a uniform alpha
// to match legacy Sprite.BlendAlpha rendering.
func nineSliceWithAlpha(ns *texture.NineSlice, alpha float32) *euiimage.NineSlice {
	img := imageWithAlpha(ns.Texture, alpha)
	w := ns.Width
	h := ns.Height
	return euiimage.NewNineSlice(
		img,
		[3]int{ns.Left, w - ns.Left - ns.Right, ns.Right},
		[3]int{ns.Top, h - ns.Top - ns.Bottom, ns.Bottom},
	)
}

func transparentNineSlice() *euiimage.NineSlice {
	return euiimage.NewNineSliceColor(color.NRGBA{0, 0, 0, 0})
}
