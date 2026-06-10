package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
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

func drawLimitbreakSingleBar(canvas *ebiten.Image, x, y float64, value int, s float64) {
	if s <= 0 {
		s = 1
	}
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
		if subFG, ok := subImg.(*ebiten.Image); ok {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s, s*widthScale)
			op.GeoM.Translate(x, y)
			canvas.DrawImage(subFG, op)
		}
	}

	frame := limitbreakTexture.GetNineSlice("limitbreak_frame.png")
	frame.Draw(canvas,
		image.Rect(int(x), int(y), int(x+float64(frame.Width)*s*widthScale), int(y+float64(frame.Height)*s)),
		nil,
	)
}

type euiLimitBreak struct {
	widget    *widget.Widget
	value     *int
	barNumber *int
	scale     float64
}

func NewEUILimitBreak(value *int, barNumber *int, scale float64) *euiLimitBreak {
	if scale <= 0 {
		scale = 1
	}
	lb := &euiLimitBreak{
		value:     value,
		barNumber: barNumber,
		scale:     scale,
	}
	lb.widget = widget.NewWidget(
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding: &widget.Insets{
				Top:  int(float64(UIPadding) * scale),
				Left: int(float64(UIPadding) * scale),
			},
		}),
	)
	return lb
}

func (l *euiLimitBreak) GetWidget() *widget.Widget {
	return l.widget
}

func (l *euiLimitBreak) SetLocation(rect image.Rectangle) {
	l.widget.Rect = rect
}

func (l *euiLimitBreak) PreferredSize() (int, int) {
	barNumber := 0
	if l.barNumber != nil {
		barNumber = *l.barNumber
	}
	return int(float64(SingleBarWidth*barNumber) * l.scale), int(float64(SingleBarHeight) * l.scale)
}

func (l *euiLimitBreak) Validate() {}

func (l *euiLimitBreak) Update(updObj *widget.UpdateObject) {
	l.widget.Update(updObj)
}

func (l *euiLimitBreak) Render(screen *ebiten.Image) {
	l.widget.Render(screen)
	x := float64(l.widget.Rect.Min.X)
	y := float64(l.widget.Rect.Min.Y)
	value := 0
	barNumber := 0
	if l.value != nil {
		value = *l.value
	}
	if l.barNumber != nil {
		barNumber = *l.barNumber
	}

	for i := 0; i < barNumber; i++ {
		barValue := value
		if barValue > SingleBarMaxValue {
			barValue = SingleBarMaxValue
		}
		drawLimitbreakSingleBar(screen, x+float64(i)*float64(SingleBarWidth)*l.scale, y, barValue, l.scale)
		value -= SingleBarMaxValue
		if value < 0 {
			value = 0
		}
	}
}
