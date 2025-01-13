package ui

import (
	"bytes"
	_ "embed"
	"fmt"
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

var (
	fontSource *text.GoTextFaceSource
	fontFace   *text.GoTextFace
)

func InitializeFont() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fontTTF))
	if err != nil {
		log.Panic(err)
	}

	fontSource = s
	fontFace = &text.GoTextFace{
		Source:    fontSource,
		Direction: text.DirectionLeftToRight,
		Size:      0,
		Language:  language.SimplifiedChinese,
	}
}

// Content must be string or func() string.
type Text struct {
	Align        furex.FlexAlignItem
	Content      any
	Color        color.Color
	Shadow       bool
	ShadowOffset float64
	ShadowColor  color.Color

	handler furex.ViewHandler
}

func (t *Text) Handler() furex.ViewHandler {
	t.handler.Extra = t
	t.handler.Draw = t.draw
	t.handler.Update = t.update

	return t.handler
}

func (t *Text) Measure(fontSize float64) (float64, float64) {
	fontFace.Size = fontSize
	content := ""

	if v, ok := t.Content.(string); ok {
		content = v
	}

	if v, ok := t.Content.(func() string); ok {
		content = v()
	}

	return text.Measure(content, fontFace, 0)
}

func (t *Text) update(v *furex.View) {
	if v.Attrs.Width == 0 {
		w, _ := t.Measure(float64(v.Attrs.Height))
		v.SetWidth(int(w))
	}
}

func (t *Text) draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
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

var textCache = make(map[string]*ebiten.Image)

func DrawText(
	screen *ebiten.Image,
	content string,
	fontSize float64,
	x, y float64,
	clr color.Color,
	align furex.FlexAlignItem,
	opt *ShadowOpt,
) {
	if content == "" {
		return
	}

	cacheKey := fmt.Sprintf("%s_%f_%v", content, fontSize, clr)

	if opt != nil {
		cacheKey += fmt.Sprintf("_%v", opt.Color)
	}

	// Try to draw from cache
	if drawFromCache(cacheKey, align, x, y, screen) {
		return
	}

	// Measure the text bounds
	fontFace.Size = fontSize
	w, h := text.Measure(content, fontFace, 0)

	// Calculate the total size including shadow
	offset := 0
	if opt != nil {
		offset = int(opt.Offset)
	}

	const shadowMultiplier = 2

	totalWidth := int(w) + shadowMultiplier*offset
	totalHeight := int(h) + shadowMultiplier*offset

	// Create a new image to draw the text and shadow
	img := ebiten.NewImage(totalWidth, totalHeight)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(offset), float64(offset))

	drawShadow(img, content, op, opt)

	op.ColorScale.ScaleWithColor(clr)
	text.Draw(img, content, fontFace, op)

	// Cache the image
	textCache[cacheKey] = img

	// Draw the cached image
	dop := &ebiten.DrawImageOptions{}

	switch align {
	case furex.AlignItemStart:
		dop.GeoM.Translate(x, y-float64(totalHeight)/2)
	case furex.AlignItemEnd:
		dop.GeoM.Translate(x-float64(totalWidth), y-float64(totalHeight)/2)
	case furex.AlignItemCenter:
		dop.GeoM.Translate(x-float64(totalWidth)/2, y-float64(totalHeight)/2)
	}

	screen.DrawImage(img, dop)
}

func drawFromCache(key string, align furex.FlexAlignItem, x, y float64, screen *ebiten.Image) bool {
	if img, ok := textCache[key]; ok {
		op := &ebiten.DrawImageOptions{}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()

		switch align {
		case furex.AlignItemStart:
			op.GeoM.Translate(x, y-float64(h)/2)
		case furex.AlignItemEnd:
			op.GeoM.Translate(x-float64(w), y-float64(h)/2)
		case furex.AlignItemCenter:
			op.GeoM.Translate(x-float64(w)/2, y-float64(h)/2)
		}

		screen.DrawImage(img, op)

		return true
	}

	return false
}

func drawShadow(img *ebiten.Image, content string, op *text.DrawOptions, opt *ShadowOpt) {
	if opt != nil {
		offset := opt.Offset
		op.ColorScale.ScaleWithColor(opt.Color)

		shadowOffsets := []struct{ dx, dy float64 }{
			{float64(offset), float64(offset)},
			{float64(offset), -float64(offset)},
			{-float64(offset), float64(offset)},
			{-float64(offset), -float64(offset)},
			{float64(offset), 0},
			{-float64(offset), 0},
			{0, float64(offset)},
			{0, -float64(offset)},
		}

		for _, o := range shadowOffsets {
			op.GeoM.Translate(o.dx, o.dy)
			text.Draw(img, content, fontFace, op)
			op.GeoM.Translate(-o.dx, -o.dy)
		}

		op.ColorScale.Reset()
	}
}
