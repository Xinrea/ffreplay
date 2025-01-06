package ui

import (
	"image"
	"log"
	"sync"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

var (
	limitbreakTexture = texture.NewTextureAtlasFromFile("asset/ui/limitbreak.xml")
	widthScale        = 0.8
)

const (
	SingleBarMaxValue = 10000
	SingleBarWidth    = 150
	SingleBarHeight   = 13
	SingleBarPadding  = 13
)

type LimitBreak struct {
	Value     *int
	BarNumber *int
	once      sync.Once

	handler furex.ViewHandler
}

var _ furex.HandlerProvider = (*LimitBreak)(nil)

func (l *LimitBreak) Handler() furex.ViewHandler {
	l.handler.Extra = l
	l.handler.Update = l.update
	l.handler.Draw = l.draw
	return l.handler
}

func (l *LimitBreak) update(v *furex.View) {
	l.once.Do(func() {
		v.SetWidth(SingleBarWidth * *l.BarNumber)
		v.SetHeight(SingleBarHeight)
	})
}

func (l *LimitBreak) draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	s := ebiten.Monitor().DeviceScaleFactor()
	x := float64(frame.Min.X)
	y := float64(frame.Min.Y)
	value := *l.Value

	for i := 0; i < *l.BarNumber; i++ {
		if value > SingleBarMaxValue {
			l.RenderLimitbreakSingleBar(screen, x+float64(i)*150*s, y, SingleBarMaxValue)
			value -= SingleBarMaxValue

			continue
		}
		l.RenderLimitbreakSingleBar(screen, x+float64(i)*150*s, y, value)
		value = 0
	}
}

func (l *LimitBreak) RenderLimitbreakSingleBar(canvas *ebiten.Image, x, y float64, value int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	bg := limitbreakTexture.GetNineSlice("limitbreak_bg.png")
	bg.Draw(canvas, image.Rect(int(x), int(y), int(x+float64(bg.Width)*s*widthScale), int(y+float64(bg.Height)*s)), nil)

	if value == SingleBarMaxValue {
		full := limitbreakTexture.GetNineSlice("limitbreak_full.png")
		full.Draw(canvas,
			image.Rect(int(x), int(y), int(x+float64(full.Width)*s*widthScale),
				int(y+float64(full.Height)*s)),
			nil,
		)
	} else {
		fg := limitbreakTexture.GetNineSlice("limitbreak_fg.png").Texture
		fgValidWidth := float64(fg.Bounds().Dx()) - SingleBarPadding*2*widthScale
		fgWidth := float64(value)/SingleBarMaxValue*fgValidWidth + SingleBarPadding*widthScale
		subImg := fg.SubImage(image.Rect(0, 0, int(fgWidth*widthScale), fg.Bounds().Dy()))
		subFG, ok := subImg.(*ebiten.Image)
		if !ok {
			log.Fatal("failed to convert sub image to ebiten.Image")
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s, s*widthScale)
		op.GeoM.Translate(x, y)
		canvas.DrawImage(subFG, op)
	}
	frame := limitbreakTexture.GetNineSlice("limitbreak_frame.png")
	frame.Draw(canvas,
		image.Rect(int(x), int(y), int(x+float64(frame.Width)*s*widthScale), int(y+float64(frame.Height)*s)),
		nil,
	)
}
