package ui

import (
	"image"
	"sync"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

var limitbreakTexture = texture.NewTextureAtlasFromFile("asset/ui/limitbreak.xml")
var widthScale = 0.8

type LimitBreak struct {
	Value     *int
	BarNumber *int
	once      sync.Once
}

func (l *LimitBreak) Update(v *furex.View) {
	l.once.Do(func() {
		v.SetWidth(150 * *l.BarNumber)
		v.SetHeight(13)
	})
}

func (l *LimitBreak) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	s := ebiten.Monitor().DeviceScaleFactor()
	x := float64(frame.Min.X)
	y := float64(frame.Min.Y)
	value := *l.Value
	for i := 0; i < *l.BarNumber; i++ {
		if value > 10000 {
			l.RenderLimitbreakSingleBar(screen, x+float64(i)*150*s, y, 10000)
			value -= 10000
			continue
		}
		l.RenderLimitbreakSingleBar(screen, x+float64(i)*150*s, y, value)
		value = 0
	}
}

func (l *LimitBreak) RenderLimitbreakSingleBar(canvas *ebiten.Image, x, y float64, value int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	bg := limitbreakTexture.GetNineSlice("limitbreak_bg.png")
	bg.Draw(canvas, image.Rect(int(x), int(y), int(x+float64(bg.Width)*s*widthScale), int(y+float64(bg.Height)*s)))
	if value == 10000 {
		full := limitbreakTexture.GetNineSlice("limitbreak_full.png")
		full.Draw(canvas, image.Rect(int(x), int(y), int(x+float64(full.Width)*s*widthScale), int(y+float64(full.Height)*s)))
	} else {
		fg := limitbreakTexture.GetNineSlice("limitbreak_fg.png").Texture
		fgWidth := float64(value)/10000*(float64(fg.Bounds().Dx())-26*widthScale) + 13.0*widthScale
		subFG := fg.SubImage(image.Rect(0, 0, int(fgWidth*widthScale), fg.Bounds().Dy())).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s, s*widthScale)
		op.GeoM.Translate(x, y)
		canvas.DrawImage(subFG, op)
	}
	frame := limitbreakTexture.GetNineSlice("limitbreak_frame.png")
	frame.Draw(canvas, image.Rect(int(x), int(y), int(x+float64(frame.Width)*s*widthScale), int(y+float64(frame.Height)*s)))
}
