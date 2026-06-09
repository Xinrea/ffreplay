package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"golang.org/x/text/language"
)

// Base sizes in logical pixels at scale=1; all are multiplied by
// DeviceScaleFactor at runtime so the panel looks the same on all displays.
const (
	propEUIPanelWidth = 240
	propEUIRowHeight  = 32
	propEUIPadding    = 10
	propEUITitleH     = 26
	propEUIMargin     = 20
	propEUIFontSize   = 13.0
	propEUIRowSpacing = 6
	propEUILabelW     = 40
)

// propBinding ties an ebitenui TextInput to ECS get/set functions.
type propBinding struct {
	input *widget.TextInput
	get   func() float64
	set   func(float64)
}

// PropertyPanelEUI manages an ebitenui floating window for editing
// the selected playground object's properties.
// It does NOT own the ebitenui.UI; it holds a shared reference so windows
// are added to the same UI instance as all other ebitenui components.
type PropertyPanelEUI struct {
	ui           *ebitenui.UI // shared, not owned
	window       *widget.Window
	removeWindow widget.RemoveWindowFunc
	bindings     []propBinding
	// euiHovered replicates Furex enter/leave semantics so global.UIHovered
	// is properly cleared when the cursor leaves the ebitenui window.
	euiHovered bool

	boundEntry *donburi.Entry
	boundInst  int
	screenW    int
	screenH    int
	scale      float64
}

// NewPropertyPanelEUI constructs the property panel using the shared ebitenui UI.
func NewPropertyPanelEUI(sharedUI *ebitenui.UI) *PropertyPanelEUI {
	return &PropertyPanelEUI{
		ui:    sharedUI,
		scale: 1,
	}
}

// UpdateECS syncs ECS state (show/hide window, input values, UIHovered/UIFocus).
// The caller (PlaygroundUI) is responsible for calling ebitenui.UI.Update() and
// ebitenui.UI.Draw() — this method must NOT call them.
func (p *PropertyPanelEUI) UpdateECS(w, h int, scale float64) {
	p.screenW = w
	p.screenH = h
	p.scale = scale
	if p.scale <= 0 {
		p.scale = 1
	}

	global := entry.GetGlobal(ecsInstance)
	selected := global.Selected

	if selected != nil && !selected.Valid() {
		selected = nil
		global.Selected = nil
	}

	if selected == nil {
		if p.removeWindow != nil {
			p.removeWindow()
			p.removeWindow = nil
			p.bindings = nil
			p.boundEntry = nil
		}
		p.syncUIHovered(global)
		return
	}

	if p.boundEntry != selected || p.boundInst != global.SelectedInstance {
		p.rebuild(selected, global.SelectedInstance)
		p.boundEntry = selected
		p.boundInst = global.SelectedInstance
	} else {
		p.syncInputs()
	}

	p.syncUIHovered(global)

	for _, b := range p.bindings {
		if b.input.IsFocused() {
			global.UIFocus = true
			break
		}
	}
}

// syncUIHovered replicates Furex enter/leave semantics:
// when ebitenui hover transitions from true→false we explicitly clear
// global.UIHovered so it does not stay stuck after the window is dragged away.
func (p *PropertyPanelEUI) syncUIHovered(global *model.GlobalData) {
	nowHovered := input.UIHovered
	switch {
	case nowHovered && !p.euiHovered:
		// Mouse entered an ebitenui widget.
		global.UIHovered = true
	case !nowHovered && p.euiHovered:
		// Mouse left all ebitenui widgets — replicate MouseLeave.
		global.UIHovered = false
	case nowHovered:
		// Still inside; keep it set.
		global.UIHovered = true
	}
	p.euiHovered = nowHovered
}

// markerTypeName returns a short display name for a WorldMarkerType.
func markerTypeName(t model.WorldMarkerType) string {
	names := map[model.WorldMarkerType]string{
		model.WorldMarkerA: "标点 A",
		model.WorldMarkerB: "标点 B",
		model.WorldMarkerC: "标点 C",
		model.WorldMarkerD: "标点 D",
		model.WorldMarker1: "标点 1",
		model.WorldMarker2: "标点 2",
		model.WorldMarker3: "标点 3",
		model.WorldMarker4: "标点 4",
	}
	if name, ok := names[t]; ok {
		return name
	}
	return "标点"
}

// rebuild closes any existing window and opens a new one for the given entry.
func (p *PropertyPanelEUI) rebuild(e *donburi.Entry, instIndex int) {
	if p.removeWindow != nil {
		p.removeWindow()
		p.removeWindow = nil
	}
	p.bindings = nil

	type spec struct {
		label string
		get   func() float64
		set   func(float64)
	}

	var specs []spec
	var titleStr string

	// WorldMarker: position-only, no rotation.
	if markerData := component.WorldMarker.Get(e); markerData != nil {
		titleStr = markerTypeName(markerData.Type)
		specs = []spec{
			{
				label: "X",
				get:   func() float64 { return markerData.Position[0] },
				set:   func(v float64) { markerData.Position[0] = v },
			},
			{
				label: "Y",
				get:   func() float64 { return markerData.Position[1] },
				set:   func(v float64) { markerData.Position[1] = v },
			},
		}
	} else {
		// Sprite-based entity (player / enemy).
		status := component.Status.Get(e)
		sprite := component.Sprite.Get(e)
		if sprite == nil || instIndex >= len(sprite.Instances) {
			return
		}

		inst := sprite.Instances[instIndex]

		specs = []spec{
			{
				label: "X",
				get:   func() float64 { return inst.Object.Position()[0] },
				set: func(v float64) {
					pos := inst.Object.Position()
					inst.Object.UpdatePosition(vector.NewVector(v, pos[1]))
				},
			},
			{
				label: "Y",
				get:   func() float64 { return inst.Object.Position()[1] },
				set: func(v float64) {
					pos := inst.Object.Position()
					inst.Object.UpdatePosition(vector.NewVector(pos[0], v))
				},
			},
			{
				label: "朝向",
				get:   func() float64 { return inst.Face * 180 / math.Pi },
				set: func(v float64) {
					rad := v * math.Pi / 180
					for rad > math.Pi {
						rad -= 2 * math.Pi
					}
					for rad < -math.Pi {
						rad += 2 * math.Pi
					}
					inst.Face = rad
				},
			},
		}

		titleStr = "属性"
		if status != nil {
			titleStr = status.Name
			specs = append(specs, spec{
				label: "HP",
				get:   func() float64 { return float64(status.HP) },
				set: func(v float64) {
					status.HP = int(v)
					if status.HP < 0 {
						status.HP = 0
					}
					if status.HP > status.MaxHP {
						status.MaxHP = status.HP
					}
				},
			})
		}
	}

	s := p.scale
	panelW := int(float64(propEUIPanelWidth) * s)
	padding := int(float64(propEUIPadding) * s)
	rowH := int(float64(propEUIRowHeight) * s)
	titleBarH := int(float64(propEUITitleH) * s)
	rowSpacing := int(float64(propEUIRowSpacing) * s)
	labelMaxW := int(float64(propEUILabelW) * s)
	margin := int(float64(propEUIMargin) * s)
	fontSize := propEUIFontSize * s
	tiPad := &widget.Insets{
		Left:   int(6 * s),
		Right:  int(6 * s),
		Top:    int(4 * s),
		Bottom: int(4 * s),
	}

	face := newEUIFace(fontSize)

	// Build each property row and collect bindings.
	rowContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(color.NRGBA{20, 22, 35, 220})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    padding,
				Bottom: padding,
				Left:   padding,
				Right:  padding,
			}),
			widget.RowLayoutOpts.Spacing(rowSpacing),
		)),
	)

	for i := range specs {
		sp := specs[i]
		ti := widget.NewTextInput(
			widget.TextInputOpts.Face(&face),
			widget.TextInputOpts.Image(&widget.TextInputImage{
				Idle:     euiimage.NewBorderedNineSliceColor(color.NRGBA{38, 40, 58, 235}, color.NRGBA{70, 72, 100, 180}, 1),
				Disabled: euiimage.NewNineSliceColor(color.NRGBA{28, 30, 42, 200}),
			}),
			widget.TextInputOpts.Color(&widget.TextInputColor{
				Idle:          color.NRGBA{220, 222, 235, 255},
				Disabled:      color.NRGBA{120, 122, 140, 255},
				Caret:         color.NRGBA{180, 200, 255, 255},
				DisabledCaret: color.NRGBA{80, 80, 100, 255},
			}),
			widget.TextInputOpts.Padding(tiPad),
			widget.TextInputOpts.SubmitOnEnter(true),
			widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
				v, err := strconv.ParseFloat(args.InputText, 64)
				if err == nil {
					sp.set(v)
				}
				args.TextInput.SetText(fmt.Sprintf("%.2f", sp.get()))
			}),
			widget.TextInputOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true, MaxHeight: rowH}),
			),
		)
		ti.SetText(fmt.Sprintf("%.2f", sp.get()))

		labelFace := newEUIFace(fontSize)
		label := widget.NewText(
			widget.TextOpts.Text(sp.label, &labelFace, color.NRGBA{180, 185, 210, 255}),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: labelMaxW}),
			),
		)

		row := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(int(8*s)),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			),
		)
		row.AddChild(label)
		row.AddChild(ti)
		rowContainer.AddChild(row)

		p.bindings = append(p.bindings, propBinding{input: ti, get: sp.get, set: sp.set})
	}

	// Title bar (drag handle).
	titleFace := newEUIFace(fontSize)
	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			euiimage.NewBorderedNineSliceColor(
				color.NRGBA{30, 32, 50, 245},
				color.NRGBA{60, 65, 100, 200},
				1,
			),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text(titleStr, &titleFace, color.NRGBA{220, 225, 255, 255}),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	))

	win := widget.NewWindow(
		widget.WindowOpts.Contents(rowContainer),
		widget.WindowOpts.TitleBar(titleBar, titleBarH),
		widget.WindowOpts.Draggable(),
	)

	winH := titleBarH + padding*2 + len(specs)*(rowH+rowSpacing) - rowSpacing + padding
	x := p.screenW - panelW - margin
	y := margin
	if x < 0 {
		x = 0
	}
	win.SetLocation(image.Rect(x, y, x+panelW, y+winH))

	p.window = win
	p.removeWindow = p.ui.AddWindow(win)
}

// syncInputs refreshes all TextInput values from ECS when not focused.
func (p *PropertyPanelEUI) syncInputs() {
	for _, b := range p.bindings {
		if !b.input.IsFocused() {
			b.input.SetText(fmt.Sprintf("%.2f", b.get()))
		}
	}
}

// newEUIFace creates a GoTextFace of the given pixel size, reusing the loaded font source.
func newEUIFace(size float64) text.Face {
	return &text.GoTextFace{
		Source:    fontSource,
		Direction: text.DirectionLeftToRight,
		Size:      size,
		Language:  language.SimplifiedChinese,
	}
}
