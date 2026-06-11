package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"golang.org/x/text/language"
)

// Base sizes in logical pixels at scale=1; all are multiplied by
// DeviceScaleFactor at runtime so the panel looks the same on all displays.
const (
	propEUIPanelWidth = 320
	propEUIRowHeight  = 28
	propEUIPadding    = 10
	propEUITitleH     = 28
	propEUIMargin     = 20
	propEUIFontSize   = 13.0
	propEUIRowSpacing = 4
	propEUILabelW     = 56
)

type propScrubState struct {
	active bool
	lastX  int
	get    func() float64
	set    func(float64)
	sense  float64
}

// PropertyPanelEUI manages an ebitenui floating window for editing
// the selected playground object's properties.
// It does NOT own the ebitenui.UI; it holds a shared reference so windows
// are added to the same UI instance as all other ebitenui components.
type PropertyPanelEUI struct {
	ui           *ebitenui.UI // shared, not owned
	window       *widget.Window
	removeWindow widget.RemoveWindowFunc
	bindings     []propFieldBinding
	euiHovered   bool
	scrub        *propScrubState

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
	if selected != nil && !isEditableSelection(selected) {
		selected = nil
		global.Selected = nil
	}

	if selected == nil {
		if p.removeWindow != nil {
			p.removeWindow()
			p.removeWindow = nil
			p.bindings = nil
			p.boundEntry = nil
			p.scrub = nil
		}
		p.syncUIHovered(global)
		return
	}

	p.updateScrub()

	if p.boundEntry != selected || p.boundInst != global.SelectedInstance {
		p.rebuild(selected, global.SelectedInstance)
		p.boundEntry = selected
		p.boundInst = global.SelectedInstance
	} else {
		p.syncInputs()
	}

	p.syncUIHovered(global)

	for _, b := range p.bindings {
		if b.input != nil && b.input.IsFocused() {
			global.UIFocus = true
			break
		}
	}
	if p.scrub != nil && p.scrub.active {
		global.UIFocus = true
	}
}

func (p *PropertyPanelEUI) updateScrub() {
	if p.scrub == nil || !p.scrub.active {
		return
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		p.scrub.active = false
		return
	}

	mx, _ := ebiten.CursorPosition()
	dx := mx - p.scrub.lastX
	p.scrub.lastX = mx
	if dx != 0 {
		p.scrub.set(p.scrub.get()+float64(dx)*p.scrub.sense)
		p.syncInputs()
	}
}

func (p *PropertyPanelEUI) beginScrub(spec propFieldSpec) {
	mx, _ := ebiten.CursorPosition()
	p.scrub = &propScrubState{
		active: true,
		lastX:  mx,
		get:    spec.Get,
		set:    spec.Set,
		sense:  spec.ScrubSense,
	}
}

// syncUIHovered replicates Furex enter/leave semantics:
// when ebitenui hover transitions from true→false we explicitly clear
// global.UIHovered so it does not stay stuck after the window is dragged away.
func (p *PropertyPanelEUI) syncUIHovered(global *model.GlobalData) {
	nowHovered := input.UIHovered
	switch {
	case nowHovered && !p.euiHovered:
		global.UIHovered = true
	case !nowHovered && p.euiHovered:
		global.UIHovered = false
	case nowHovered:
		global.UIHovered = true
	}
	p.euiHovered = nowHovered
}

func isEditableSelection(e *donburi.Entry) bool {
	return e.HasComponent(component.WorldMarker) || e.HasComponent(component.Sprite)
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

func (p *PropertyPanelEUI) rebuild(e *donburi.Entry, instIndex int) {
	if p.removeWindow != nil {
		p.removeWindow()
		p.removeWindow = nil
	}
	p.bindings = nil
	p.scrub = nil

	var transformFields []propFieldSpec
	var statusFields []propFieldSpec
	titleStr := "属性"

	if e.HasComponent(component.WorldMarker) {
		markerData := component.WorldMarker.Get(e)
		titleStr = markerTypeName(markerData.Type)
		transformFields = []propFieldSpec{
			{
				Label: "X", Get: func() float64 { return markerData.Position[0] },
				Set: func(v float64) { markerData.Position[0] = v },
				Step: 1, ScrubSense: propEUIScrubSensePos,
			},
			{
				Label: "Y", Get: func() float64 { return markerData.Position[1] },
				Set: func(v float64) { markerData.Position[1] = v },
				Step: 1, ScrubSense: propEUIScrubSensePos,
			},
		}
	} else if e.HasComponent(component.Sprite) {
		status := component.Status.Get(e)
		sprite := component.Sprite.Get(e)
		if sprite == nil || instIndex >= len(sprite.Instances) {
			return
		}

		inst := sprite.Instances[instIndex]
		transformFields = []propFieldSpec{
			{
				Label: "X", Get: func() float64 { return inst.Object.Position()[0] },
				Set: func(v float64) {
					pos := inst.Object.Position()
					inst.Object.UpdatePosition(vector.NewVector(v, pos[1]))
				},
				Step: 1, ScrubSense: propEUIScrubSensePos,
			},
			{
				Label: "Y", Get: func() float64 { return inst.Object.Position()[1] },
				Set: func(v float64) {
					pos := inst.Object.Position()
					inst.Object.UpdatePosition(vector.NewVector(pos[0], v))
				},
				Step: 1, ScrubSense: propEUIScrubSensePos,
			},
			{
				Label: "朝向", Get: func() float64 { return inst.Face * 180 / math.Pi },
				Set: func(v float64) {
					rad := v * math.Pi / 180
					for rad > math.Pi {
						rad -= 2 * math.Pi
					}
					for rad < -math.Pi {
						rad += 2 * math.Pi
					}
					inst.Face = rad
				},
				Step: 5, ScrubSense: propEUIScrubSenseRot,
				SliderMin: -180, SliderMax: 180, HasSlider: true,
			},
		}

		if status != nil {
			titleStr = status.Name
			statusFields = []propFieldSpec{
				{
					Label: "HP", Get: func() float64 { return float64(status.HP) },
					Set: func(v float64) {
						status.HP = int(v)
						if status.HP < 0 {
							status.HP = 0
						}
						if status.HP > status.MaxHP {
							status.HP = status.MaxHP
						}
					},
					Step: 100, Format: "%.0f", ScrubSense: propEUIScrubSenseHP,
					SliderMin: 0,
					SliderMaxFunc: func() float64 {
						max := float64(status.MaxHP)
						if max < 1 {
							return 1
						}
						return max
					},
					HasSlider: true,
				},
				{
					Label: "HP 上限", Get: func() float64 { return float64(status.MaxHP) },
					Set: func(v float64) {
						status.MaxHP = int(v)
						if status.MaxHP < 1 {
							status.MaxHP = 1
						}
						if status.HP > status.MaxHP {
							status.HP = status.MaxHP
						}
					},
					Step: 1000, Format: "%.0f", ScrubSense: propEUIScrubSenseHP * 10,
					SliderMin: 1, SliderMax: propMaxHPSliderMax, HasSlider: true,
				},
			}
		}
	} else {
		return
	}

	st := newPropInspectorStyle(p.scale)
	margin := int(float64(propEUIMargin) * p.scale)
	titleBarH := int(float64(propEUITitleH) * p.scale)

	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(color.NRGBA{24, 26, 38, 235})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    st.padding,
				Bottom: st.padding,
				Left:   st.padding,
				Right:  st.padding,
			}),
			widget.RowLayoutOpts.Spacing(st.rowSpacing),
		)),
	)

	contentH := st.padding * 2
	onScrub := func(spec propFieldSpec) { p.beginScrub(spec) }

	addSection := func(title string, fields []propFieldSpec) {
		if len(fields) == 0 {
			return
		}
		section, body := newPropSection(title, st)
		for _, field := range fields {
			row, binding := buildPropFieldRow(st, field, onScrub)
			body.AddChild(row)
			p.bindings = append(p.bindings, binding)
			contentH += fieldBlockHeight(st, field)
		}
		contentH += int(float64(propEUISectionHeader)*st.scale) + st.rowSpacing
		panel.AddChild(section)
	}

	addSection("变换", transformFields)
	addSection("状态", statusFields)

	titleFace := newEUIFace(st.fontSize)
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
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding: &widget.Insets{
				Left: int(10 * st.scale),
			},
		})),
	))

	win := widget.NewWindow(
		widget.WindowOpts.Contents(panel),
		widget.WindowOpts.TitleBar(titleBar, titleBarH),
		widget.WindowOpts.Draggable(),
	)

	winH := titleBarH + contentH
	x := p.screenW - st.panelW - margin
	y := margin
	if x < 0 {
		x = 0
	}
	win.SetLocation(image.Rect(x, y, x+st.panelW, y+winH))

	p.window = win
	p.removeWindow = p.ui.AddWindow(win)
}

func fieldBlockHeight(st propInspectorStyle, field propFieldSpec) int {
	h := st.rowH + st.rowSpacing
	if field.HasSlider {
		h += st.sliderH + int(4*st.scale)
	}
	return h
}

// syncInputs refreshes all field widgets from ECS when not being edited.
func (p *PropertyPanelEUI) syncInputs() {
	for i := range p.bindings {
		b := &p.bindings[i]
		if b.input != nil && !b.input.IsFocused() {
			b.syncText()
		}
		b.syncSlider()
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
