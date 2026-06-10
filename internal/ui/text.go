package ui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

type TextAlign int

const (
	AlignStart TextAlign = iota
	AlignEnd
	AlignCenter
)

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
	align any,
	opt *ShadowOpt,
) {
	if content == "" {
		return
	}
	textAlign := normalizeTextAlign(align)

	cacheKey := fmt.Sprintf("%s_%f_%v", content, fontSize, clr)

	if opt != nil {
		cacheKey += fmt.Sprintf("_%v", opt.Color)
	}

	// Try to draw from cache
	if drawFromCache(cacheKey, textAlign, x, y, screen) {
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

	switch textAlign {
	case AlignStart:
		dop.GeoM.Translate(x, y-float64(totalHeight)/2)
	case AlignEnd:
		dop.GeoM.Translate(x-float64(totalWidth), y-float64(totalHeight)/2)
	case AlignCenter:
		dop.GeoM.Translate(x-float64(totalWidth)/2, y-float64(totalHeight)/2)
	}

	screen.DrawImage(img, dop)
}

func drawFromCache(key string, align TextAlign, x, y float64, screen *ebiten.Image) bool {
	if img, ok := textCache[key]; ok {
		op := &ebiten.DrawImageOptions{}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()

		switch align {
		case AlignStart:
			op.GeoM.Translate(x, y-float64(h)/2)
		case AlignEnd:
			op.GeoM.Translate(x-float64(w), y-float64(h)/2)
		case AlignCenter:
			op.GeoM.Translate(x-float64(w)/2, y-float64(h)/2)
		}

		screen.DrawImage(img, op)

		return true
	}

	return false
}

func normalizeTextAlign(align any) TextAlign {
	switch v := align.(type) {
	case TextAlign:
		return v
	}

	rv := reflect.ValueOf(align)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TextAlign(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return TextAlign(rv.Uint())
	default:
		return AlignStart
	}
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
