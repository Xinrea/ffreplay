package ui

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	euiinput "github.com/ebitenui/ebitenui/input"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	propEUISliderH       = 10
	propEUISectionHeader = 24
	propEUIButtonSize    = 24
	propEUIScrubSensePos = 0.25
	propEUIScrubSenseRot = 0.5
	propEUIScrubSenseHP  = 50
	// Fixed slider ceiling for MaxHP; precise values use text input / steppers.
	propMaxHPSliderMax = 10_000_000
)

// propFieldSpec describes one editable scalar in the inspector.
type propFieldSpec struct {
	Label      string
	Get        func() float64
	Set        func(float64)
	Step       float64
	Format     string
	ScrubSense float64
	SliderMin     float64
	SliderMax     float64
	SliderMaxFunc func() float64
	HasSlider     bool
}

type propFieldBinding struct {
	input          *widget.TextInput
	slider         *widget.Slider
	sliderMin      float64
	sliderMax      float64
	sliderMaxFunc  func() float64
	get            func() float64
	set            func(float64)
	format         string
	sliderDragging bool
}

type propInspectorStyle struct {
	scale      float64
	panelW     int
	padding    int
	rowH       int
	rowSpacing int
	labelW     int
	btnW       int
	sliderH    int
	fontSize   float64
	face       text.Face
	tiPad      *widget.Insets
}

func newPropInspectorStyle(scale float64) propInspectorStyle {
	if scale <= 0 {
		scale = 1
	}
	panelW := int(float64(propEUIPanelWidth) * scale)
	padding := int(float64(propEUIPadding) * scale)
	rowH := int(float64(propEUIRowHeight) * scale)
	fontSize := propEUIFontSize * scale
	return propInspectorStyle{
		scale:      scale,
		panelW:     panelW,
		padding:    padding,
		rowH:       rowH,
		rowSpacing: int(float64(propEUIRowSpacing) * scale),
		labelW:     int(float64(propEUILabelW) * scale),
		btnW:       int(float64(propEUIButtonSize) * scale),
		sliderH:    int(float64(propEUISliderH) * scale),
		fontSize:   fontSize,
		face:       newEUIFace(fontSize),
		tiPad: &widget.Insets{
			Left:   int(6 * scale),
			Right:  int(6 * scale),
			Top:    int(4 * scale),
			Bottom: int(4 * scale),
		},
	}
}

func propButtonImage() *widget.ButtonImage {
	idle := color.NRGBA{48, 50, 68, 255}
	hover := color.NRGBA{58, 62, 82, 255}
	pressed := color.NRGBA{38, 40, 56, 255}
	return &widget.ButtonImage{
		Idle:         euiimage.NewNineSliceColor(idle),
		Hover:        euiimage.NewNineSliceColor(hover),
		Pressed:      euiimage.NewNineSliceColor(pressed),
		PressedHover: euiimage.NewNineSliceColor(pressed),
		Disabled:     euiimage.NewNineSliceColor(color.NRGBA{32, 34, 46, 200}),
	}
}

func propTextInputImage() *widget.TextInputImage {
	return &widget.TextInputImage{
		Idle:     euiimage.NewBorderedNineSliceColor(color.NRGBA{38, 40, 58, 235}, color.NRGBA{70, 72, 100, 180}, 1),
		Disabled: euiimage.NewNineSliceColor(color.NRGBA{28, 30, 42, 200}),
	}
}

func propTextInputColor() *widget.TextInputColor {
	return &widget.TextInputColor{
		Idle:          color.NRGBA{220, 222, 235, 255},
		Disabled:      color.NRGBA{120, 122, 140, 255},
		Caret:         color.NRGBA{180, 200, 255, 255},
		DisabledCaret: color.NRGBA{80, 80, 100, 255},
	}
}

func propSliderImages() (*widget.SliderTrackImage, *widget.ButtonImage) {
	track := &widget.SliderTrackImage{
		Idle:  euiimage.NewNineSliceColor(color.NRGBA{32, 34, 48, 255}),
		Hover: euiimage.NewNineSliceColor(color.NRGBA{40, 42, 58, 255}),
	}
	handle := &widget.ButtonImage{
		Idle:    euiimage.NewNineSliceColor(color.NRGBA{120, 150, 220, 255}),
		Hover:   euiimage.NewNineSliceColor(color.NRGBA{140, 170, 235, 255}),
		Pressed: euiimage.NewNineSliceColor(color.NRGBA{100, 130, 200, 255}),
	}
	return track, handle
}

func newPropSection(title string, st propInspectorStyle) (*widget.Container, *widget.Container) {
	headerH := int(float64(propEUISectionHeader) * st.scale)
	section := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(color.NRGBA{34, 36, 52, 255})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.panelW-st.padding*2, headerH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	headerFace := newEUIFace(st.fontSize)
	header.AddChild(widget.NewText(
		widget.TextOpts.Text(title, &headerFace, color.NRGBA{170, 175, 200, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding: &widget.Insets{
				Left: int(8 * st.scale),
			},
		})),
	))
	section.AddChild(header)

	body := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    int(4 * st.scale),
				Bottom: int(4 * st.scale),
			}),
			widget.RowLayoutOpts.Spacing(st.rowSpacing),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	section.AddChild(body)

	return section, body
}

func buildPropFieldRow(
	st propInspectorStyle,
	spec propFieldSpec,
	onScrubStart func(propFieldSpec),
) (*widget.Container, propFieldBinding) {
	if spec.Format == "" {
		spec.Format = "%.2f"
	}
	if spec.Step <= 0 {
		spec.Step = 1
	}
	if spec.ScrubSense <= 0 {
		spec.ScrubSense = 0.1
	}

	rowSpacingPx := int(6 * st.scale)
	controlSpacing := int(4 * st.scale)
	inputMinW := st.panelW - st.padding*2 - st.labelW - st.btnW*2 - rowSpacingPx - controlSpacing*3
	if inputMinW < int(72*st.scale) {
		inputMinW = int(72 * st.scale)
	}

	binding := propFieldBinding{
		get:       spec.Get,
		set:       spec.Set,
		format:    spec.Format,
		sliderMin:     spec.SliderMin,
		sliderMax:     spec.SliderMax,
		sliderMaxFunc: spec.SliderMaxFunc,
	}

	applyValue := func(v float64) {
		spec.Set(v)
		binding.syncText()
		binding.syncSlider()
	}

	sp := spec
	ti := widget.NewTextInput(
		widget.TextInputOpts.Face(&st.face),
		widget.TextInputOpts.Image(propTextInputImage()),
		widget.TextInputOpts.Color(propTextInputColor()),
		widget.TextInputOpts.Padding(st.tiPad),
		widget.TextInputOpts.SubmitOnEnter(true),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			v, err := parsePropNumber(args.InputText)
			if err == nil {
				applyValue(v)
			} else {
				binding.syncText()
			}
		}),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(inputMinW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true, MaxHeight: st.rowH}),
		),
	)
	binding.input = ti
	binding.syncText()

	labelFace := newEUIFace(st.fontSize)
	label := widget.NewText(
		widget.TextOpts.Text(spec.Label, &labelFace, color.NRGBA{150, 155, 180, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.labelW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.labelW}),
			widget.WidgetOpts.CursorHovered(euiinput.CURSOR_EWRESIZE),
			widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft && onScrubStart != nil {
					onScrubStart(sp)
				}
			}),
		),
	)

	btnFace := newEUIFace(st.fontSize * 0.95)
	minusBtn := widget.NewButton(
		widget.ButtonOpts.Text("−", &btnFace, &widget.ButtonTextColor{
			Idle: color.NRGBA{210, 212, 225, 255},
		}),
		widget.ButtonOpts.Image(propButtonImage()),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.btnW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.btnW, MaxHeight: st.rowH}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			applyValue(spec.Get() - spec.Step)
		}),
	)
	plusBtn := widget.NewButton(
		widget.ButtonOpts.Text("+", &btnFace, &widget.ButtonTextColor{
			Idle: color.NRGBA{210, 212, 225, 255},
		}),
		widget.ButtonOpts.Image(propButtonImage()),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.btnW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.btnW, MaxHeight: st.rowH}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			applyValue(spec.Get() + spec.Step)
		}),
	)

	fieldRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(controlSpacing),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	fieldRow.AddChild(label)
	fieldRow.AddChild(ti)
	fieldRow.AddChild(minusBtn)
	fieldRow.AddChild(plusBtn)

	block := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(int(4 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	block.AddChild(fieldRow)

	if spec.HasSlider && (spec.SliderMax > spec.SliderMin || spec.SliderMaxFunc != nil) {
		sliderW := st.panelW - st.padding*2 - st.labelW
		track, handle := propSliderImages()
		slider := widget.NewSlider(
			widget.SliderOpts.MinMax(0, 1000),
			widget.SliderOpts.InitialCurrent(binding.valueToSlider(spec.Get())),
			widget.SliderOpts.Images(track, handle),
			widget.SliderOpts.FixedHandleSize(int(8 * st.scale)),
			widget.SliderOpts.TrackOffset(0),
			widget.SliderOpts.PageSizeFunc(func() int { return 10 }),
			widget.SliderOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(sliderW, st.sliderH),
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch:   true,
					MaxHeight: st.sliderH,
					Position:  widget.RowLayoutPositionEnd,
				}),
			),
			widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
				binding.sliderDragging = args.Dragging
				spec.Set(binding.sliderToValue(args.Current))
				if binding.input != nil && !binding.input.IsFocused() {
					binding.syncText()
				}
			}),
		)
		binding.slider = slider

		sliderRow := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		)
		spacer := widget.NewContainer(
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(st.labelW, 1),
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.labelW}),
			),
		)
		sliderRow.AddChild(spacer)
		sliderRow.AddChild(slider)
		block.AddChild(sliderRow)
	}

	return block, binding
}

func (b *propFieldBinding) syncText() {
	if b.input == nil {
		return
	}
	b.input.SetText(formatPropNumber(b.get(), b.format))
}

// formatPropNumber renders a field value with thousands separators.
func formatPropNumber(v float64, format string) string {
	if format == "" {
		format = "%.2f"
	}
	return addThousandsSeparator(fmt.Sprintf(format, v))
}

// parsePropNumber accepts plain or comma-separated numeric input.
func parsePropNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}

func addThousandsSeparator(s string) string {
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	}

	intPart, fracPart, hasFrac := s, "", false
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		intPart = s[:dot]
		fracPart = s[dot+1:]
		hasFrac = true
	}

	if len(intPart) <= 3 {
		if hasFrac {
			return sign + intPart + "." + fracPart
		}
		return sign + intPart
	}

	var builder strings.Builder
	builder.Grow(len(s) + len(s)/3)
	builder.WriteString(sign)
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			builder.WriteByte(',')
		}
		builder.WriteRune(digit)
	}
	if hasFrac {
		builder.WriteByte('.')
		builder.WriteString(fracPart)
	}
	return builder.String()
}

func (b *propFieldBinding) syncSlider() {
	if b.slider == nil || b.sliderDragging {
		return
	}
	b.slider.Current = b.valueToSlider(b.get())
}

func (b *propFieldBinding) effectiveSliderMax() float64 {
	if b.sliderMaxFunc != nil {
		if v := b.sliderMaxFunc(); v > b.sliderMin {
			return v
		}
	}
	return b.sliderMax
}

func (b *propFieldBinding) valueToSlider(v float64) int {
	max := b.effectiveSliderMax()
	if max <= b.sliderMin {
		return 0
	}
	ratio := (v - b.sliderMin) / (max - b.sliderMin)
	ratio = math.Max(0, math.Min(1, ratio))
	return int(math.Round(ratio * 1000))
}

func (b *propFieldBinding) sliderToValue(current int) float64 {
	max := b.effectiveSliderMax()
	if max <= b.sliderMin {
		return b.sliderMin
	}
	ratio := float64(current) / 1000
	return b.sliderMin + ratio*(max-b.sliderMin)
}