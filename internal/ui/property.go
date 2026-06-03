package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

// PropertyPanel shows adjustable parameters for the selected playground object.
// It is hidden when nothing is selected.
type PropertyPanel struct {
	view    *furex.View
	body    *furex.View
	title   *Text
	handler furex.ViewHandler

	// boundEntry tracks which entry the body rows are currently built for, so
	// we only rebuild the rows when the selection changes.
	boundEntry *donburi.Entry
	boundInst  int
}

const (
	propPanelWidth = 220
	propRowHeight  = 24
)

// NewPropertyPanel builds the property panel view (initially hidden).
func NewPropertyPanel() *PropertyPanel {
	p := &PropertyPanel{}

	p.title = &Text{
		Align:        furex.AlignItemStart,
		Content:      "属性",
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{0, 0, 0, 160},
	}

	view := furex.NewView(
		furex.TagName("property"),
		furex.Direction(furex.Column),
		furex.Width(propPanelWidth),
		furex.Display(furex.DisplayNone),
		furex.Handler(p.buildHandler()),
	)

	// Background fills the panel.
	view.AddChild(furex.NewView(
		furex.Position(furex.PositionAbsolute),
		furex.Top(0),
		furex.Left(0),
		furex.Grow(1),
		furex.Handler(&Sprite{
			NineSliceTexture: messageTextureAtlas.GetNineSlice("message_bg.png"),
			BlendAlpha:       true,
			Alpha:            0.85,
		}),
	))

	titleRow := furex.NewView(
		furex.Height(20),
		furex.MarginTop(8),
		furex.MarginLeft(10),
		furex.Handler(p.title),
	)
	view.AddChild(titleRow)

	body := furex.NewView(
		furex.ID("property-body"),
		furex.Direction(furex.Column),
		furex.MarginTop(4),
		furex.MarginBottom(8),
	)
	view.AddChild(body)

	p.view = view
	p.body = body

	return p
}

// View returns the root view of the panel.
func (p *PropertyPanel) View() *furex.View {
	return p.view
}

// buildHandler wires the panel's update + hover behavior.
func (p *PropertyPanel) buildHandler() furex.ViewHandler {
	p.handler.Extra = p
	p.handler.Update = p.update
	p.handler.MouseEnter = func(x, y int) bool {
		entry.GetGlobal(ecsInstance).UIHovered = true

		return true
	}
	p.handler.MouseLeave = func() {
		entry.GetGlobal(ecsInstance).UIHovered = false
	}

	return p.handler
}

// update shows/hides the panel based on the current selection and rebuilds the
// editable rows when the selected object changes.
func (p *PropertyPanel) update(v *furex.View) {
	global := entry.GetGlobal(ecsInstance)

	selected := global.Selected
	if selected != nil && !selected.Valid() {
		selected = nil
		global.Selected = nil
	}

	if selected == nil {
		if p.view.Attrs.Display != furex.DisplayNone {
			p.view.SetDisplay(furex.DisplayNone)
			p.boundEntry = nil
			global.UIHovered = false
		}

		return
	}

	if p.view.Attrs.Display != furex.DisplayFlex {
		p.view.SetDisplay(furex.DisplayFlex)
	}

	if p.boundEntry != selected || p.boundInst != global.SelectedInstance {
		p.rebuild(selected, global.SelectedInstance)
		p.boundEntry = selected
		p.boundInst = global.SelectedInstance
	}
}

// rebuild populates the body with rows for the given entry/instance.
func (p *PropertyPanel) rebuild(e *donburi.Entry, instIndex int) {
	p.body.RemoveAll()

	status := component.Status.Get(e)
	sprite := component.Sprite.Get(e)
	if sprite == nil || instIndex >= len(sprite.Instances) {
		return
	}

	inst := sprite.Instances[instIndex]

	if status != nil {
		p.title.Content = status.Name
	}

	// Position X / Y stepper rows (in world units).
	p.body.AddChild(p.newStepperRow("X", 5, func() float64 {
		return inst.Object.Position()[0]
	}, func(delta float64) {
		pos := inst.Object.Position()
		inst.Object.UpdatePosition(vector.NewVector(pos[0]+delta, pos[1]))
	}))
	p.body.AddChild(p.newStepperRow("Y", 5, func() float64 {
		return inst.Object.Position()[1]
	}, func(delta float64) {
		pos := inst.Object.Position()
		inst.Object.UpdatePosition(vector.NewVector(pos[0], pos[1]+delta))
	}))

	// Facing in degrees (stored internally as radians from north).
	p.body.AddChild(p.newStepperRow("朝向", 15, func() float64 {
		return inst.Face * 180 / math.Pi
	}, func(delta float64) {
		inst.Face += delta * math.Pi / 180
		if inst.Face > math.Pi {
			inst.Face -= 2 * math.Pi
		}
		if inst.Face < -math.Pi {
			inst.Face += 2 * math.Pi
		}
	}))

	if status != nil {
		p.body.AddChild(p.newStepperRow("HP", 10000, func() float64 {
			return float64(status.HP)
		}, func(delta float64) {
			status.HP += int(delta)
			if status.HP < 0 {
				status.HP = 0
			}
			if status.HP > status.MaxHP {
				status.MaxHP = status.HP
			}
		}))
	}
}

// newStepperRow builds a row "<label>  [-] <value> [+]" where the buttons
// apply -/+ step via the onStep callback and value reflects get().
func (p *PropertyPanel) newStepperRow(
	label string,
	step float64,
	get func() float64,
	onStep func(delta float64),
) *furex.View {
	row := furex.NewView(
		furex.Height(propRowHeight),
		furex.MarginLeft(10),
		furex.MarginRight(10),
		furex.MarginTop(4),
		furex.Direction(furex.Row),
		furex.Justify(furex.JustifySpaceBetween),
		furex.AlignItems(furex.AlignItemCenter),
	)

	// Label.
	row.AddChild(furex.NewView(
		furex.Width(48),
		furex.Height(propRowHeight),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      label,
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 1,
			ShadowColor:  color.NRGBA{0, 0, 0, 160},
		}),
	))

	// Controls group [-] value [+].
	controls := furex.NewView(
		furex.Direction(furex.Row),
		furex.AlignItems(furex.AlignItemCenter),
	)
	controls.AddChild(p.newStepButton("-", func() { onStep(-step) }))
	controls.AddChild(furex.NewView(
		furex.Width(80),
		furex.Height(propRowHeight),
		furex.MarginLeft(4),
		furex.MarginRight(4),
		furex.Handler(&Text{
			Align: furex.AlignItemCenter,
			Content: func() string {
				return fmt.Sprintf("%.0f", get())
			},
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 1,
			ShadowColor:  color.NRGBA{0, 0, 0, 160},
		}),
	))
	controls.AddChild(p.newStepButton("+", func() { onStep(step) }))

	row.AddChild(controls)

	return row
}

// newStepButton builds a small clickable button with a centered glyph.
func (p *PropertyPanel) newStepButton(glyph string, onClick func()) *furex.View {
	btn := &stepButton{onClick: onClick}
	view := furex.NewView(
		furex.Width(22),
		furex.Height(22),
		furex.Handler(btn.handlerFor(glyph)),
	)

	return view
}

// stepButton is a minimal clickable button used by the property panel.
type stepButton struct {
	onClick func()
	handler furex.ViewHandler
}

func (b *stepButton) handlerFor(glyph string) furex.ViewHandler {
	textHandler := &Text{
		Align:        furex.AlignItemCenter,
		Content:      glyph,
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  color.NRGBA{0, 0, 0, 200},
	}

	b.handler.Extra = b
	b.handler.Draw = func(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
		// button background
		bg := messageTextureAtlas.GetNineSlice("input_bg.png")
		if bg != nil {
			bg.Draw(screen, frame, nil)
		}
		textHandler.Handler().Draw(screen, frame, v)
	}
	b.handler.JustPressedMouseButtonLeft = func(frame image.Rectangle, x, y int) bool {
		if b.onClick != nil {
			b.onClick()
		}

		return true
	}
	b.handler.JustReleasedMouseButtonLeft = func(frame image.Rectangle, x, y int) {}
	b.handler.MouseEnter = func(x, y int) bool {
		entry.GetGlobal(ecsInstance).UIHovered = true

		return true
	}

	return b.handler
}
