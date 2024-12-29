package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/furex/v2"
	"golang.org/x/text/language"
)

//go:embed OPPOSans-Regular.ttf
var fontTTF []byte

var fontSource *text.GoTextFaceSource

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fontTTF))
	if err != nil {
		log.Fatal(err)
	}
	fontSource = s

}

// Content must be string or func() string
type Text struct {
	Align        furex.AlignItem
	Content      any
	Color        color.Color
	Shadow       bool
	ShadowOffset float64
	ShadowColor  color.Color
}

func (t *Text) Update(v *furex.View) {
	f := &text.GoTextFace{
		Source:    fontSource,
		Direction: text.DirectionLeftToRight,
		Size:      float64(v.Height),
		Language:  language.SimplifiedChinese,
	}
	content := ""
	if v, ok := t.Content.(string); ok {
		content = v
	}
	if v, ok := t.Content.(func() string); ok {
		content = v()
	}
	w, _ := text.Measure(content, f, 0)
	v.SetWidth(int(w))
}

func (t *Text) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	content := ""
	if v, ok := t.Content.(string); ok {
		content = v
	}
	if v, ok := t.Content.(func() string); ok {
		content = v()
	}
	x := float64(frame.Min.X)
	y := float64(frame.Min.Y) + float64(frame.Dy())/2
	switch t.Align {
	case furex.AlignItemEnd:
		x += float64(frame.Dx())
	case furex.AlignItemCenter:
		x += float64(frame.Dx()) / 2
	}
	var opt *ShadowOpt = nil
	if t.Shadow {
		opt = &ShadowOpt{
			Color:  t.ShadowColor,
			Offset: t.ShadowOffset,
		}
	}
	DrawText(screen, content, float64(frame.Dy()), x, y, t.Color, t.Align, opt)
}

type ShadowOpt struct {
	Color  color.Color
	Offset float64
}

func DrawText(screen *ebiten.Image, content string, fontSize float64, x, y float64, clr color.Color, align furex.AlignItem, opt *ShadowOpt) {
	f := &text.GoTextFace{
		Source:    fontSource,
		Direction: text.DirectionLeftToRight,
		Size:      fontSize,
		Language:  language.SimplifiedChinese,
	}
	op := &text.DrawOptions{}
	w, h := text.Measure(content, f, 0)
	switch align {
	case furex.AlignItemStart:
		op.GeoM.Translate(x, y-h/2)
	case furex.AlignItemEnd:
		op.GeoM.Translate(x-w, y-h/2)
	case furex.AlignItemCenter:
		op.GeoM.Translate(x-w/2, y-h/2)
	}

	if opt != nil {
		offset := opt.Offset
		op.ColorScale.ScaleWithColor(opt.Color)
		op.GeoM.Translate(offset, offset)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(0, -offset)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(0, -offset)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(-offset, 0)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(-offset, 0)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(0, offset)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(0, offset)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(offset, 0)
		text.Draw(screen, content, f, op)
		op.GeoM.Translate(0, -offset)
		op.ColorScale.Reset()
	}
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, content, f, op)
}
