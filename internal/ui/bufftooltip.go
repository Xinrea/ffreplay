package ui

import (
	"image"
	"image/color"
	"strconv"

	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var buffTooltipAreas []buffHitArea

type buffHitArea struct {
	buff *UIBuff
	rect image.Rectangle
}

func BeginBuffTooltipFrame() {
	buffTooltipAreas = buffTooltipAreas[:0]
}

func TrackBuffTooltip(buff *UIBuff, rect image.Rectangle) {
	if buff == nil || rect.Empty() {
		return
	}
	buffTooltipAreas = append(buffTooltipAreas, buffHitArea{buff: buff, rect: rect})
}

func BuffHitRect(x, y, scale float64) image.Rectangle {
	if scale <= 0 {
		scale = 1
	}
	w := BuffWidth * scale
	h := float64(BuffHeight+BuffRemainFontSize+BuffRemainTop) * scale
	if h < BuffHeight*scale {
		h = BuffHeight * scale
	}
	return image.Rect(int(x), int(y), int(x+w), int(y+h))
}

func buffTooltipLabel(buff *UIBuff) string {
	if buff == nil {
		return ""
	}
	if buff.Name == "" {
		return strconv.FormatInt(buff.ID, 10)
	}
	if buff.ID == 0 {
		return buff.Name
	}

	return buff.Name + "\n" + strconv.FormatInt(buff.ID, 10)
}

func newBuffToolTip(buff *UIBuff, scale float64) *widget.ToolTip {
	label := buffTooltipLabel(buff)
	if label == "" {
		return nil
	}
	if scale <= 0 {
		scale = 1
	}

	face := newEUIFace(12 * scale)
	bg := euiimage.NewNineSliceColor(color.NRGBA{16, 20, 28, 230})

	return widget.NewTextToolTip(
		label,
		&face,
		color.White,
		bg,
	)
}

func hoveredManualBuffTooltip() *UIBuff {
	mx, my := ebiten.CursorPosition()
	p := image.Pt(mx, my)
	for i := len(buffTooltipAreas) - 1; i >= 0; i-- {
		if rectContains(buffTooltipAreas[i].rect, p) {
			return buffTooltipAreas[i].buff
		}
	}

	return nil
}

func DrawBuffTooltip(screen *ebiten.Image, scale float64) {
	buff := hoveredManualBuffTooltip()
	if buff == nil {
		return
	}
	if scale <= 0 {
		scale = 1
	}

	label := buffTooltipLabel(buff)
	if label == "" {
		return
	}

	mx, my := ebiten.CursorPosition()
	fontSize := 12 * scale
	padding := 6.0 * scale
	lineH := fontSize * 1.2

	lines := splitTooltipLines(label)
	maxW := 0.0
	for _, line := range lines {
		w, _ := measureText(line, fontSize)
		if w > maxW {
			maxW = w
		}
	}

	boxW := maxW + padding*2
	boxH := float64(len(lines))*lineH + padding*2
	boxX := float64(mx) + 12*scale
	boxY := float64(my) + 16*scale

	bounds := screen.Bounds()
	if boxX+boxW > float64(bounds.Dx()) {
		boxX = float64(mx) - boxW - 12*scale
	}
	if boxY+boxH > float64(bounds.Dy()) {
		boxY = float64(my) - boxH - 8*scale
	}

	bg := euiimage.NewNineSliceColor(color.NRGBA{16, 20, 28, 230})
	bg.Draw(screen, int(boxW), int(boxH), func(opts *ebiten.DrawImageOptions) {
		opts.GeoM.Translate(boxX, boxY)
	})

	for i, line := range lines {
		DrawText(
			screen,
			line,
			fontSize,
			boxX+padding,
			boxY+padding+lineH*float64(i)+lineH/2,
			color.White,
			AlignStart,
			nil,
		)
	}
}

func splitTooltipLines(label string) []string {
	if label == "" {
		return nil
	}

	lines := []string{}
	start := 0
	for i := 0; i < len(label); i++ {
		if label[i] != '\n' {
			continue
		}
		lines = append(lines, label[start:i])
		start = i + 1
	}
	lines = append(lines, label[start:])

	return lines
}

func measureText(content string, fontSize float64) (float64, float64) {
	if content == "" {
		return 0, 0
	}

	face := newEUIFace(fontSize)
	return text.Measure(content, face, 0)
}

func rectContains(rect image.Rectangle, p image.Point) bool {
	return p.X >= rect.Min.X && p.X < rect.Max.X && p.Y >= rect.Min.Y && p.Y < rect.Max.Y
}
