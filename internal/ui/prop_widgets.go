package ui

import (
	"image"
	"image/color"

	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/utilities/constantutil"
	"github.com/ebitenui/ebitenui/widget"
)

func propListButtonParams(st propInspectorStyle) *widget.ButtonParams {
	face := newEUIFace(st.fontSize)
	return &widget.ButtonParams{
		Image:       propButtonImage(),
		TextPadding: widget.NewInsetsSimple(int(6 * st.scale)),
		TextColor: &widget.ButtonTextColor{
			Idle:     color.NRGBA{220, 222, 235, 255},
			Disabled: color.NRGBA{120, 122, 140, 255},
		},
		TextFace: &face,
		MinSize:  &image.Point{int(120 * st.scale), st.rowH},
	}
}

func propListParams(st propInspectorStyle) *widget.ListParams {
	face := newEUIFace(st.fontSize)
	btnImg := propButtonImage()
	return &widget.ListParams{
		ScrollContainerImage: &widget.ScrollContainerImage{
			Idle:     euiimage.NewNineSliceColor(color.NRGBA{28, 30, 42, 255}),
			Disabled: euiimage.NewNineSliceColor(color.NRGBA{24, 26, 36, 255}),
			Mask:     euiimage.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
		},
		Slider: &widget.SliderParams{
			TrackImage: &widget.SliderTrackImage{
				Idle:  euiimage.NewNineSliceColor(color.NRGBA{40, 42, 58, 255}),
				Hover: euiimage.NewNineSliceColor(color.NRGBA{48, 50, 66, 255}),
			},
			HandleImage:   btnImg,
			MinHandleSize: constantutil.ConstantToPointer(int(6 * st.scale)),
			TrackPadding:  widget.NewInsetsSimple(int(2 * st.scale)),
		},
		EntryFace: &face,
		EntryColor: &widget.ListEntryColor{
			Selected:                   color.NRGBA{235, 238, 250, 255},
			Unselected:                 color.NRGBA{200, 204, 220, 255},
			SelectedBackground:         color.NRGBA{70, 90, 150, 255},
			SelectedFocusedBackground:  color.NRGBA{90, 110, 170, 255},
			FocusedBackground:          color.NRGBA{50, 54, 72, 255},
			DisabledUnselected:         color.NRGBA{100, 100, 110, 255},
			DisabledSelected:           color.NRGBA{100, 100, 110, 255},
			DisabledSelectedBackground: color.NRGBA{40, 42, 52, 255},
		},
		EntryTextPadding: widget.NewInsetsSimple(int(6 * st.scale)),
		MinSize:          &image.Point{int(120 * st.scale), st.rowH},
	}
}

func propLabeledControlRow(st propInspectorStyle, label string, control widget.PreferredSizeLocateableWidget) *widget.Container {
	labelFace := newEUIFace(st.fontSize)
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.MinSize(st.panelW-st.padding*2, st.rowH),
		),
	)
	row.AddChild(widget.NewText(
		widget.TextOpts.Text(label, &labelFace, color.NRGBA{150, 155, 180, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.labelW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.labelW}),
		),
	))
	control.GetWidget().LayoutData = widget.RowLayoutData{Stretch: true}
	row.AddChild(control)
	return row
}

func propFieldLabel(st propInspectorStyle, label string) *widget.Text {
	labelFace := newEUIFace(st.fontSize)
	return widget.NewText(
		widget.TextOpts.Text(label, &labelFace, color.NRGBA{150, 155, 180, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.labelW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: st.labelW}),
		),
	)
}

func propCompactTextInput(st propInspectorStyle, width int, initial string) *widget.TextInput {
	face := newEUIFace(st.fontSize)
	ti := widget.NewTextInput(
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Image(propTextInputImage()),
		widget.TextInputOpts.Color(propTextInputColor()),
		widget.TextInputOpts.Padding(st.tiPad),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(width, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: width, MaxHeight: st.rowH}),
		),
	)
	if initial != "" {
		ti.SetText(initial)
	}
	return ti
}

func propActionButton(st propInspectorStyle, label string, onClick func()) *widget.Button {
	btnFace := newEUIFace(st.fontSize * 0.95)
	btnW := int(52 * st.scale)
	return widget.NewButton(
		widget.ButtonOpts.Text(label, &btnFace, &widget.ButtonTextColor{
			Idle: color.NRGBA{210, 212, 225, 255},
		}),
		widget.ButtonOpts.Image(propButtonImage()),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(btnW, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: btnW, MaxHeight: st.rowH}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			onClick()
		}),
	)
}
