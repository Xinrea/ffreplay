package renderer

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

var LIMITBREAK_BG = color.NRGBA{37, 51, 73, 255}
var LIMITBREAK_FG = color.NRGBA{17, 166, 255, 255}

// Limit break value [0, 30000], each bar stands for 10000
func (r *Renderer) RenderLimitbreakBar(canvas *ebiten.Image, x, y float64, bar int, value int) {
	for i := 0; i < bar; i++ {
		if value > 10000 {
			r.RenderLimitbreakSingleBar(canvas, x+float64(i)*165, y, 10000)
			value -= 10000
			continue
		}
		r.RenderLimitbreakSingleBar(canvas, x+float64(i)*165, y, value)
		value = 0
	}
}

func (r *Renderer) RenderLimitbreakSingleBar(canvas *ebiten.Image, x, y float64, value int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	if value == 10000 {
		t := texture.NewTextureFromFile("asset/limitbreak_full.png")
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		op.GeoM.Scale(s, s)
		canvas.DrawImage(t.Img(), op)
		return
	}
	t := texture.NewTextureFromFile("asset/limitbreak.png")
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(s, s)
	DrawFilledRect(canvas, x+19, y+7, 130, 8, LIMITBREAK_BG)
	DrawFilledRect(canvas, x+19, y+7, 130*float64(value)/10000, 8, LIMITBREAK_FG)
	canvas.DrawImage(t.Img(), op)
}
