package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

var checkboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/checkbox.xml")
var multicheckboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/multicheckbox.xml")

type CheckBoxHandler struct {
	Checked      bool
	ClickHandler func(bool)
}

var _ furex.Updater = (*CheckBoxHandler)(nil)
var _ furex.MouseLeftButtonHandler = (*CheckBoxHandler)(nil)
var _ furex.MouseEnterLeaveHandler = (*CheckBoxHandler)(nil)

func (c *CheckBoxHandler) Update(v *furex.View) {
	if c.Checked {
		v.MustGetByID("checked").Hidden = false
	} else {
		v.MustGetByID("checked").Hidden = true
	}
}

func (c *CheckBoxHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool {
	c.Checked = !c.Checked
	if c.ClickHandler != nil {
		c.ClickHandler(c.Checked)
	}
	return true
}

func (c *CheckBoxHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int) {

}

// HandleMouseEnter implements furex.MouseEnterLeaveHandler.
func (c *CheckBoxHandler) HandleMouseEnter(x int, y int) bool {
	ebiten.SetCursorShape(ebiten.CursorShapePointer)
	return true
}

// HandleMouseLeave implements furex.MouseEnterLeaveHandler.
func (c *CheckBoxHandler) HandleMouseLeave() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}

func CheckBoxView(size int, multiple bool, initial bool, clickHandler func(bool)) *furex.View {
	view := &furex.View{
		Width:  size,
		Height: size,
		Handler: &CheckBoxHandler{
			Checked:      initial,
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
	return view
}
