package ui

import (
	"image"
	"image/color"
	"sync"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type LimitBreak struct {
	Value     *int
	BarNumber *int
	once      sync.Once
}

func (l *LimitBreak) Update(v *furex.View) {
	l.once.Do(func() {
		v.SetWidth(165 * *l.BarNumber)
		v.SetHeight(22)
	})
}

func (l *LimitBreak) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	s := ebiten.Monitor().DeviceScaleFactor()
	x := float64(frame.Min.X)
	y := float64(frame.Min.Y)
	value := *l.Value
	for i := 0; i < *l.BarNumber; i++ {
		if value > 10000 {
			l.RenderLimitbreakSingleBar(screen, x+float64(i)*165*s, y, 10000)
			value -= 10000
			continue
		}
		l.RenderLimitbreakSingleBar(screen, x+float64(i)*165*s, y, value)
		value = 0
	}
}

var LIMITBREAK_BG = ebiten.NewImage(1, 1)
var LIMITBREAK_FG = ebiten.NewImage(1, 1)

func init() {
	LIMITBREAK_BG.Fill(color.NRGBA{37, 51, 73, 255})
	LIMITBREAK_FG.Fill(color.NRGBA{17, 166, 255, 255})
}

func (l *LimitBreak) RenderLimitbreakSingleBar(canvas *ebiten.Image, x, y float64, value int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	ox := x + 19*s
	oy := y + 7*s
	if value == 10000 {
		t := texture.NewTextureFromFile("asset/limitbreak_full.png")
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(x, y)
		canvas.DrawImage(t, op)
		return
	}
	t := texture.NewTextureFromFile("asset/limitbreak.png")

	{
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(130, 8)
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(ox, oy)
		canvas.DrawImage(LIMITBREAK_BG, op)
	}

	{
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(130*float64(value)/10000, 8)
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(ox, oy)
		canvas.DrawImage(LIMITBREAK_FG, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x, y)
	canvas.DrawImage(t, op)
}
