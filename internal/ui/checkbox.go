package ui

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	checkboxTextureAtlas      = texture.NewTextureAtlasFromFile("asset/ui/checkbox.xml")
	multicheckboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/multicheckbox.xml")
)

// NewEUICheckbox creates an ebitenui Checkbox using the game's checkbox sprites,
// scaled to match the device pixel ratio. value is read at creation time for the
// initial state and written back on every state change.
func NewEUICheckbox(size int, multiple bool, value *bool, label string, clickHandler func(bool), scale float64) *widget.Checkbox {
	atlas := checkboxTextureAtlas
	if multiple {
		atlas = multicheckboxTextureAtlas
	}

	boxPx := int(float64(size) * scale)

	uncheckedImg := scaleImage(atlas.GetNineSlice("checkbox_bg.png").Texture, boxPx, boxPx)
	checkedImg := scaleImage(atlas.GetNineSlice("checkbox_checked.png").Texture, boxPx, boxPx)

	uncheckedSlice := euiimage.NewFixedNineSlice(uncheckedImg)
	checkedSlice := euiimage.NewFixedNineSlice(checkedImg)

	var face text.Face = newEUIFace(float64(size) * scale)

	initState := widget.WidgetUnchecked
	if *value {
		initState = widget.WidgetChecked
	}

	return widget.NewCheckbox(
		widget.CheckboxOpts.Image(&widget.CheckboxImage{
			Unchecked:        uncheckedSlice,
			Checked:          checkedSlice,
			UncheckedHovered: uncheckedSlice,
			CheckedHovered:   checkedSlice,
		}),
		widget.CheckboxOpts.Text(label, &face, &widget.LabelColor{
			Idle:     color.White,
			Disabled: color.NRGBA{255, 255, 255, 128},
		}),
		widget.CheckboxOpts.Spacing(int(float64(size)*scale*0.5+4*scale)),
		widget.CheckboxOpts.InitialState(initState),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			*value = args.State == widget.WidgetChecked
			if clickHandler != nil {
				clickHandler(*value)
			}
		}),
	)
}
