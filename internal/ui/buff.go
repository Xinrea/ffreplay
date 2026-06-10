package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	BuffWidth          = 24
	BuffHeight         = 32
	BuffStackFontSize  = 13
	BuffStackTop       = 2
	BuffStackLeft      = 9
	BuffStackShadow    = 4
	EUIBuffStackShadow = 1
	BuffRemainTop      = -6
	BuffRemainFontSize = 12
)

func EUIBuffView(buff *UIBuff, scale float64) *widget.Container {
	if scale <= 0 {
		scale = 1
	}

	w := int(BuffWidth * scale)
	iconH := int(BuffHeight * scale)
	totalH := int(float64(BuffHeight+BuffRemainFontSize+BuffRemainTop) * scale)
	if totalH < iconH {
		totalH = iconH
	}

	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.MinSize(w, totalH),
		widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			MaxWidth:  w,
			MaxHeight: totalH,
		}),
	}
	if tip := newBuffToolTip(buff, scale); tip != nil {
		widgetOpts = append(widgetOpts, widget.WidgetOpts.ToolTip(tip))
	}

	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widgetOpts...),
	)

	view.AddChild(widget.NewGraphic(
		// Match the legacy Furex Sprite frame (24x32), otherwise icons look
		// vertically compressed compared with the original game UI.
		widget.GraphicOpts.Image(scaleImage(buff.Texture(), w, iconH)),
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
		})),
	))

	if buff.Stacks > 1 {
		view.AddChild(newEUIShadowText(
			strconv.Itoa(buff.Stacks),
			BuffStackFontSize*scale,
			int(BuffStackFontSize*scale),
			int(BuffStackFontSize*scale),
			AlignEnd,
			color.White,
			EUIBuffStackShadow*scale,
			color.NRGBA{0, 0, 0, 200},
			widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding: &widget.Insets{
					Top:  int(BuffStackTop * scale),
					Left: int(BuffStackLeft * scale),
				},
			},
		))
	}

	remainText := newEUIShadowText(
		formatSeconds(buff.Remain),
		BuffRemainFontSize*scale,
		w,
		int(BuffRemainFontSize*scale),
		AlignCenter,
		color.White,
		1*scale,
		color.NRGBA{0, 0, 0, 128},
		widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
		},
	)
	view.AddChild(remainText)

	return view
}

type euiShadowText struct {
	widget       *widget.Widget
	content      any
	fontSize     float64
	width        int
	height       int
	align        TextAlign
	color        color.Color
	shadowColor  color.Color
	shadowOffset float64
}

func newEUIShadowText(
	content any,
	fontSize float64,
	width int,
	height int,
	align TextAlign,
	clr color.Color,
	shadowOffset float64,
	shadowColor color.Color,
	layoutData any,
) *euiShadowText {
	t := &euiShadowText{
		content:      content,
		fontSize:     fontSize,
		width:        width,
		height:       height,
		align:        align,
		color:        clr,
		shadowColor:  shadowColor,
		shadowOffset: shadowOffset,
	}
	t.widget = widget.NewWidget(widget.WidgetOpts.LayoutData(layoutData))
	return t
}

func (t *euiShadowText) GetWidget() *widget.Widget {
	return t.widget
}

func (t *euiShadowText) SetLocation(rect image.Rectangle) {
	t.widget.Rect = rect
}

func (t *euiShadowText) PreferredSize() (int, int) {
	return t.width, t.height
}

func (t *euiShadowText) Validate() {}

func (t *euiShadowText) Update(updObj *widget.UpdateObject) {
	t.widget.Update(updObj)
}

func (t *euiShadowText) resolveContent() string {
	switch v := t.content.(type) {
	case func() string:
		return v()
	case string:
		return v
	default:
		return ""
	}
}

func (t *euiShadowText) Render(screen *ebiten.Image) {
	t.widget.Render(screen)

	frame := t.widget.Rect
	x := float64(frame.Min.X)
	switch t.align {
	case AlignCenter:
		x += float64(frame.Dx()) / 2
	case AlignEnd:
		x += float64(frame.Dx())
	}

	DrawText(
		screen,
		t.resolveContent(),
		t.fontSize,
		x,
		float64(frame.Min.Y)+float64(frame.Dy())/2,
		t.color,
		t.align,
		&ShadowOpt{Color: t.shadowColor, Offset: t.shadowOffset},
	)
}

const secondsInMinute = 60

func formatSeconds(seconds int64) string {
	minutes := seconds / secondsInMinute
	hours := minutes / secondsInMinute

	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	if seconds > 0 {
		return fmt.Sprintf("%d", seconds)
	}

	return ""
}
