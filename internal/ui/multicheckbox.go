package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

var multicheckboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/multicheckbox.xml")

type MultiCheckBoxHandler struct {
	checked *bool
}

var _ furex.Updater = (*MultiCheckBoxHandler)(nil)
var _ furex.MouseLeftButtonHandler = (*MultiCheckBoxHandler)(nil)
var _ furex.MouseEnterLeaveHandler = (*MultiCheckBoxHandler)(nil)

func (c *MultiCheckBoxHandler) Update(v *furex.View) {
	if *c.checked {
		v.MustGetByID("checked").Hidden = false
	} else {
		v.MustGetByID("checked").Hidden = true
	}
}

func (c *MultiCheckBoxHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool {
	*c.checked = !*c.checked
	return true
}

func (c *MultiCheckBoxHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int) {

}

// HandleMouseEnter implements furex.MouseEnterLeaveHandler.
func (c *MultiCheckBoxHandler) HandleMouseEnter(x int, y int) bool {
	ebiten.SetCursorShape(ebiten.CursorShapePointer)
	return true
}

// HandleMouseLeave implements furex.MouseEnterLeaveHandler.
func (c *MultiCheckBoxHandler) HandleMouseLeave() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}

func MultiCheckBoxView(size int, checked *bool) *furex.View {
	view := &furex.View{
		Width:  size,
		Height: size,
		Handler: &MultiCheckBoxHandler{
			checked: checked,
		},
	}
	view.AddChild(&furex.View{
		Position: furex.PositionAbsolute,
		Width:    size,
		Height:   size,
		Top:      0,
		Left:     0,
		Handler: &Sprite{
			NineSliceTexture: multicheckboxTextureAtlas.GetNineSlice("multicheckbox_bg.png"),
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
			NineSliceTexture: multicheckboxTextureAtlas.GetNineSlice("multicheckbox_checked.png"),
		},
	})
	return view
}
