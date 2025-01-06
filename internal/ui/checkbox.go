package ui

import (
	"image"
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/yohamta/furex/v2"
)

var (
	checkboxTextureAtlas      = texture.NewTextureAtlasFromFile("asset/ui/checkbox.xml")
	multicheckboxTextureAtlas = texture.NewTextureAtlasFromFile("asset/ui/multicheckbox.xml")
)

type CheckBoxHandler struct {
	Size         int
	Checked      *bool
	ClickHandler func(bool)

	handler furex.ViewHandler
}

func (c *CheckBoxHandler) Handler() furex.ViewHandler {
	c.handler.Extra = c
	c.handler.Update = c.Update
	c.handler.JustPressedMouseButtonLeft = c.HandleJustPressedMouseButtonLeft
	c.handler.JustReleasedMouseButtonLeft = c.HandleJustReleasedMouseButtonLeft

	return c.handler
}

func (c *CheckBoxHandler) Update(v *furex.View) {
	v.SetWidth(v.Last().Attrs.Width + 5 + c.Size)

	if *c.Checked {
		v.NthChild(1).Attrs.Hidden = false
	} else {
		v.NthChild(1).Attrs.Hidden = true
	}
}

func (c *CheckBoxHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool {
	*c.Checked = !*c.Checked
	if c.ClickHandler != nil {
		c.ClickHandler(*c.Checked)
	}

	return true
}

func (c *CheckBoxHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int) {}

func CheckBoxView(size int, multiple bool, value *bool, label string, clickHandler func(bool)) *furex.View {
	view := furex.NewView(furex.Height(size), furex.AlignItems(furex.AlignItemCenter), furex.Handler(&CheckBoxHandler{
		Size:         size,
		Checked:      value,
		ClickHandler: clickHandler,
	}))

	addCheckBoxSprites(multiple, size, view)

	view.AddChild(
		furex.NewView(
			furex.ID("label"),
			furex.MarginLeft(int(float64(size))),
			furex.Height(int(float64(size))),
			furex.Handler(&Text{
				Align:        furex.AlignItemStart,
				Content:      label,
				Color:        color.White,
				Shadow:       true,
				ShadowOffset: 2,
				ShadowColor:  color.NRGBA{0, 0, 0, 128},
			})))

	return view
}

func addCheckBoxSprites(multiple bool, size int, view *furex.View) {
	atlas := checkboxTextureAtlas
	if multiple {
		atlas = multicheckboxTextureAtlas
	}

	view.AddChild(
		furex.NewView(
			furex.Position(furex.PositionAbsolute),
			furex.Width(size),
			furex.Height(size),
			furex.Top(0),
			furex.Left(0),
			furex.Handler(&Sprite{
				NineSliceTexture: atlas.GetNineSlice("checkbox_bg.png"),
			})))
	view.AddChild(
		furex.NewView(furex.ID("checked"),
			furex.Hidden(true),
			furex.Position(furex.PositionAbsolute),
			furex.Width(size),
			furex.Height(size),
			furex.Top(0),
			furex.Left(0),
			furex.Handler(&Sprite{
				NineSliceTexture: atlas.GetNineSlice("checkbox_checked.png"),
			})))
}
