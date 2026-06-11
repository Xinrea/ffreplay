package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"golang.org/x/image/math/f64"
	"golang.org/x/text/language"
)

// Base sizes in logical pixels at scale=1; all are multiplied by
// DeviceScaleFactor at runtime so the panel looks the same on all displays.
const (
	propEUIPanelWidth = 320
	propEUIRowHeight  = 28
	propEUIPadding    = 10
	propEUITitleH     = 28
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

// PropertyPanelEUI manages a right-sidebar panel for editing
// the selected playground object's properties.
// It does NOT own the ebitenui.UI; it holds a shared reference so the
// sidebar is added to the same UI root as all other ebitenui components.
type PropertyPanelEUI struct {
	ui             *ebitenui.UI // shared, not owned
	wrapper        *widget.Container
	wrapperInRoot  bool
	bindings       []propFieldBinding
	stringBindings []propStringFieldBinding
	buffManager    *EUIBuffListManager
	jobPicker      *EUIJobPicker
	scrub          *propScrubState

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

	p.updateScrub()

	if !p.wrapperInRoot || p.boundEntry != selected || p.boundInst != global.SelectedInstance {
		p.rebuild(selected, global.SelectedInstance)
		p.boundEntry = selected
		p.boundInst = global.SelectedInstance
	} else if selected != nil {
		p.syncInputs()
	}

	for _, b := range p.bindings {
		if b.input != nil && b.input.IsFocused() {
			global.UIFocus = true
			break
		}
	}
	for _, b := range p.stringBindings {
		if b.input != nil && b.input.IsFocused() {
			global.UIFocus = true
			break
		}
	}
	if p.buffManager != nil && (p.buffManager.Focused() || p.buffManager.ComboOpen()) {
		global.UIFocus = true
	}
	if p.jobPicker != nil && (p.jobPicker.Focused() || p.jobPicker.ComboOpen()) {
		global.UIFocus = true
	}
	if p.scrub != nil && p.scrub.active {
		global.UIFocus = true
	}
	if p.cursorOverPanel() {
		global.UIHovered = true
	}
}

func (p *PropertyPanelEUI) cursorOverPanel() bool {
	if p.wrapper == nil {
		return false
	}
	rect := p.wrapper.GetWidget().Rect
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return image.Pt(mx, my).In(rect)
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

func isEditableSelection(e *donburi.Entry) bool {
	return e.HasComponent(component.WorldMarker) || e.HasComponent(component.Sprite)
}

func isPlayerEntry(e *donburi.Entry) bool {
	return e.HasComponent(tag.Player)
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
	if p.jobPicker != nil {
		p.jobPicker.Close()
	}
	if p.buffManager != nil {
		p.buffManager.Close()
	}
	p.bindings = nil
	p.stringBindings = nil
	p.buffManager = nil
	p.jobPicker = nil
	p.scrub = nil

	st := newPropInspectorStyle(p.scale)
	titleBarH := int(float64(propEUITitleH) * p.scale)

	var transformFields []propFieldSpec
	var statusFields []propFieldSpec
	var isPlayer bool
	titleStr := "属性"
	hasSelection := e != nil

	if hasSelection {
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
				if isPlayerEntry(e) {
					isPlayer = true
					p.buffManager = NewEUIBuffListManager(newPropInspectorStyle(p.scale), p.ui, status.EnsureBuffList())
				}
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
	}

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
		}
		panel.AddChild(section)
	}

	// --- Always-visible sections ---

	// Waymarker grid: 2×4 marker buttons inside the sidebar.
	waymarkerSection, waymarkerBody := newPropSection("标点", st)
	waymarkerBody.AddChild(buildWaymarkerGrid(st))
	panel.AddChild(waymarkerSection)

	// Display toggle: show target ring checkbox.
	globalData := entry.GetGlobal(ecsInstance)
	displaySection, displayBody := newPropSection("显示", st)
	checkboxRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	checkboxRow.AddChild(NewEUICheckbox(14, true, &globalData.ShowTargetRing, "显示目标圈", nil, st.scale))
	displayBody.AddChild(checkboxRow)
	panel.AddChild(displaySection)

	// --- Selection-only sections ---

	if hasSelection {
		addSection("变换", transformFields)

		if isPlayer {
			section, body := newPropSection("玩家", st)
			if status := component.Status.Get(e); status != nil {
				nameRow, nameBinding := buildPropStringFieldRow(st, "名字", func() string {
					return status.Name
				}, func(name string) {
					status.Name = name
				})
				body.AddChild(nameRow)
				p.stringBindings = append(p.stringBindings, nameBinding)

				p.jobPicker = NewEUIJobPicker(st, p.ui, status.Role, func(r role.RoleType) {
					status.Role = r
				})
				body.AddChild(p.jobPicker.Container())
			}
			panel.AddChild(section)
		}

		addSection("状态", statusFields)

		if p.buffManager != nil {
			section, body := newPropSection("Buff", st)
			body.AddChild(p.buffManager.Container())
			panel.AddChild(section)
		}
	}

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
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.panelW, titleBarH),
		),
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

	// Scrollable content area: panel sections scroll if taller than window.
	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(panel),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: euiimage.NewNineSliceColor(color.NRGBA{24, 26, 38, 235}),
			Mask: euiimage.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
		}),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	// Outer content: title bar (fixed) + scrollable panel.
	contentOuter := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)
	contentOuter.AddChild(titleBar)
	contentOuter.AddChild(scrollContainer)

	// Ensure the right-sidebar wrapper is in the root container.
	if !p.wrapperInRoot {
		p.wrapper = widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(st.panelW, p.screenH),
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionEnd,
					VerticalPosition:   widget.AnchorLayoutPositionStart,
				}),
			),
		)
		p.ui.Container.AddChild(p.wrapper)
		p.wrapperInRoot = true
	}

	// Populate the wrapper with the new sidebar content.
	p.wrapper.RemoveChildren()
	p.wrapper.AddChild(contentOuter)
}

// buildWaymarkerGrid creates a 2-row × 4-column grid of world-marker buttons
// that toggle markers at the current cursor world position.
func buildWaymarkerGrid(st propInspectorStyle) *widget.Container {
	slotPx := int(48 * st.scale)
	gap := int(2 * st.scale)

	grid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(gap),
		)),
	)

	emptySlice := euiimage.NewNineSlice(
		hotbarAtlasTexture.GetNineSlice("hotbar_empty.png").Texture,
		[3]int{0, 48, 0}, [3]int{0, 48, 0},
	)

	for r := range 2 {
		rowContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(gap),
			)),
		)

		for c := range 4 {
			marker := model.WorldMarkerType(r*4 + c)
			if marker > model.WorldMarker4 {
				break
			}

			config := model.WorldMarkerConfigs[marker]
			idleImg, pressedImg := buildWaymarkerButtonImages(config.Texture, slotPx)

			m := marker // capture for closure

			btn := widget.NewButton(
				widget.ButtonOpts.Image(&widget.ButtonImage{
					Idle:         euiimage.NewNineSlice(idleImg, [3]int{0, slotPx, 0}, [3]int{0, slotPx, 0}),
					Hover:        euiimage.NewNineSlice(idleImg, [3]int{0, slotPx, 0}, [3]int{0, slotPx, 0}),
					Pressed:      euiimage.NewNineSlice(pressedImg, [3]int{0, slotPx, 0}, [3]int{0, slotPx, 0}),
					PressedHover: euiimage.NewNineSlice(pressedImg, [3]int{0, slotPx, 0}, [3]int{0, slotPx, 0}),
					Disabled:     emptySlice,
				}),
				widget.ButtonOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.RowLayoutData{
						MaxWidth:  slotPx,
						MaxHeight: slotPx,
					}),
				),
				widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
					g := entry.GetGlobal(ecsInstance)
					camera := entry.GetCamera(ecsInstance)
					for markerEntry := range component.WorldMarker.Iter(ecsInstance.World) {
						md := component.WorldMarker.Get(markerEntry)
						if md.Type == m {
							markerEntry.Remove()
							return
						}
					}
					x, y := ebiten.CursorPosition()
					wx, wy := camera.ScreenToWorld(float64(x), float64(y))
					entry.NewWorldMarker(ecsInstance, m, f64.Vec2{wx, wy})
					g.WorldMarkerSelected = int(m)
				}),
			)
			rowContainer.AddChild(btn)
		}

		grid.AddChild(rowContainer)
	}

	return grid
}

// buildWaymarkerButtonImages pre-composites layered hotbar visuals into
// idle (icon + fg overlay) and pressed (idle + clicked overlay) images.
func buildWaymarkerButtonImages(icon *ebiten.Image, slotPx int) (*ebiten.Image, *ebiten.Image) {
	fg := hotbarAtlasTexture.GetNineSlice("hotbar_fg.png").Texture
	clicked := hotbarAtlasTexture.GetNineSlice("hotbar_clicked.png").Texture

	idle := ebiten.NewImage(slotPx, slotPx)
	overlayImage(idle, icon)
	overlayImage(idle, fg)

	pressed := ebiten.NewImage(slotPx, slotPx)
	overlayImage(pressed, idle)
	overlayImage(pressed, clicked)

	return idle, pressed
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
	for i := range p.stringBindings {
		b := &p.stringBindings[i]
		if b.input != nil && !b.input.IsFocused() {
			b.syncText()
		}
	}
	if p.buffManager != nil {
		p.buffManager.RefreshIfChanged()
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
