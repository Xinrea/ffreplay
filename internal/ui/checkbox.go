package ui

import (
	"image"
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/yohamta/furex/v2"
)

var checkboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/checkbox.xml")
var multicheckboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/multicheckbox.xml")

type CheckBoxHandler struct {
	Size         int
	Checked      *bool
	ClickHandler func(bool)
}

var _ furex.Updater = (*CheckBoxHandler)(nil)
var _ furex.MouseLeftButtonHandler = (*CheckBoxHandler)(nil)

func (c *CheckBoxHandler) Update(v *furex.View) {
	v.SetWidth(v.MustGetByID("label").Width + 5 + c.Size)
	if *c.Checked {
		v.MustGetByID("checked").Hidden = false
	} else {
		v.MustGetByID("checked").Hidden = true
	}
}

func (c *CheckBoxHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool {
	*c.Checked = !*c.Checked
	if c.ClickHandler != nil {
		c.ClickHandler(*c.Checked)
	}
	return true
}

func (c *CheckBoxHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int) {

}

func CheckBoxView(size int, multiple bool, value *bool, label string, clickHandler func(bool)) *furex.View {
	view := &furex.View{
		Height:     size,
		AlignItems: furex.AlignItemCenter,
		Handler: &CheckBoxHandler{
			Size:         size,
			Checked:      value,
			ClickHandler: clickHandler,
		},
	}
	if multiple {
		view.AddChild(&furex.View{
			Position: furex.PositionAbsolute,
			Width:    size,
			Height:   size,
			Top:      0,
			Left:     0,
			Handler: &Sprite{
				NineSliceTexture: multicheckboxTextureAtlas.GetNineSlice("checkbox_bg.png"),
			},
		})
		view.AddChild(&furex.View{
			ID:       "checked",
			Hidden:   true,
			Position: furex.PositionAbsolute,
			Width:    size,
			Height:   size,
			Top:      0,
			Left:     0,
			Handler: &Sprite{
				NineSliceTexture: multicheckboxTextureAtlas.GetNineSlice("checkbox_checked.png"),
			},
		})
	} else {
		view.AddChild(&furex.View{
			Position: furex.PositionAbsolute,
			Width:    size,
			Height:   size,
			Top:      0,
			Left:     0,
			Handler: &Sprite{
				NineSliceTexture: checkboxTextureAtlas.GetNineSlice("checkbox_bg.png"),
			},
		})
		view.AddChild(&furex.View{
			ID:       "checked",
			Hidden:   true,
			Position: furex.PositionAbsolute,
			Width:    size,
			Height:   size,
			Top:      0,
			Left:     0,
			Handler: &Sprite{
				NineSliceTexture: checkboxTextureAtlas.GetNineSlice("checkbox_checked.png"),
			},
		})
	}
	view.AddChild(&furex.View{
		ID:         "label",
		MarginLeft: int(float64(size) * 1.2),
		Height:     int(float64(size) * 0.8),
		Handler: &Text{
			Align:        furex.AlignItemStart,
			Content:      label,
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		},
	})
	return view
}
