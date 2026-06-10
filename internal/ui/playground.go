package ui

import (
	"image"
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
	"golang.org/x/image/math/f64"
)

// Due to the design of furex.View, root view cannot handle any events,
// so we need a global root as wrap.
var root = furex.NewView(furex.ID("Root"))

type PlaygroundUI struct {
	base      *furex.View
	eui       *ebitenui.UI
	euiRoot   *widget.Container
	propPanel *PropertyPanelEUI
	once      sync.Once
}

var _ UI = (*PlaygroundUI)(nil)

func NewPlaygroundUI(ecs *ecs.ECS) *PlaygroundUI {
	ecsInstance = ecs

	baseWrap := furex.NewView(
		furex.ID("Playground"),
		furex.Direction(furex.Column),
		furex.Justify(furex.JustifySpaceBetween),
		furex.Grow(1),
	)
	baseWrap.Handler.JustPressedMouseButtonLeft = func(frame image.Rectangle, x int, y int) bool {
		for _, c := range baseWrap.FilterByTagName("input") {
			if fh, ok := c.Handler.Extra.(Focusable); ok {
				fh.SetFocus(false)
			}
		}

		entry.GetGlobal(ecsInstance).UIFocus = false

		return false
	}

	root.AddChild(baseWrap)

	// Shared ebitenui root: one UI instance for ALL ebitenui components in this scene.
	euiRoot := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	eui := &ebitenui.UI{Container: euiRoot}

	return &PlaygroundUI{
		base:      baseWrap,
		eui:       eui,
		euiRoot:   euiRoot,
		propPanel: NewPropertyPanelEUI(eui),
	}
}

func (p *PlaygroundUI) Update(w, h int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	if s <= 0 {
		s = 1
	}
	furex.GlobalScale = s

	global := entry.GetGlobal(ecsInstance)
	if global.Loaded.Load() {
		p.once.Do(func() {
			p.buildFurexUI()
			p.euiRoot.AddChild(NewEUICommandView(s))
			p.buildEUITopRight(s, global)
		})
	}

	root.UpdateWithSize(w, h)

	// Recompute focus from the current ebitenui state each frame. Focused
	// widgets below will set this back to true during/after p.eui.Update().
	global.UIFocus = false

	// Single ebitenui update covers property panel, hotbar, checkbox, etc.
	p.eui.Update()
	p.propPanel.UpdateECS(w, h, s)
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {
	root.Draw(screen)
	// Single ebitenui draw for all migrated components.
	p.eui.Draw(screen)
}

// buildFurexUI sets up the remaining Furex layout (party list).
func (p *PlaygroundUI) buildFurexUI() {
	topView := furex.NewView(
		furex.Grow(1),
		furex.Direction(furex.Row),
		furex.Justify(furex.JustifySpaceBetween),
	)

	partyList := NewPartyList(nil)
	partyList.Attrs.MarginTop = 40
	partyList.Attrs.MarginLeft = UIPadding

	topView.AddChild(partyList)

	p.base.AddChild(topView)
}

// buildEUITopRight creates the ebitenui top-right column with HotBar and Checkbox.
func (p *PlaygroundUI) buildEUITopRight(scale float64, global *model.GlobalData) {
	pad := int(float64(UIPadding) * scale)

	// Outer container anchored to top-right, with screen-edge padding.
	col := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:   pad,
				Right: pad,
			}),
			widget.RowLayoutOpts.Spacing(int(50*scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	hotbar := p.buildEUIHotBar(2, 8, scale)
	col.AddChild(hotbar)

	checkbox := NewEUICheckbox(14, true, &global.ShowTargetRing, "显示目标圈", nil, scale)
	col.AddChild(checkbox)

	p.euiRoot.AddChild(col)
}

// buildEUIHotBar creates a rows×cols grid of ebitenui Buttons using game hotbar textures.
// All icons and click handlers are configured immediately (no two-step HotBarView/SetupHotBar).
func (p *PlaygroundUI) buildEUIHotBar(rows, cols int, scale float64) *widget.Container {
	slotPx := int(48 * scale)
	gap := int(2 * scale)

	hotbarContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(gap),
		)),
	)

	emptySlice := euiimage.NewNineSlice(
		hotbarAtlasTexture.GetNineSlice("hotbar_empty.png").Texture,
		[3]int{0, 48, 0}, [3]int{0, 48, 0},
	)

	for r := range rows {
		rowContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(gap),
			)),
		)

		for c := range cols {
			marker := model.WorldMarkerType(r*cols + c)
			if marker > model.WorldMarker4 {
				break
			}

			config := model.WorldMarkerConfigs[marker]
			idleImg, pressedImg := buildHotbarButtonImages(config.Texture, slotPx)

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

		hotbarContainer.AddChild(rowContainer)
	}

	return hotbarContainer
}

// buildHotbarButtonImages pre-composites the layered hotbar visuals into two
// *ebiten.Image: idle (icon + fg overlay) and pressed (idle + clicked overlay).
func buildHotbarButtonImages(icon *ebiten.Image, slotPx int) (*ebiten.Image, *ebiten.Image) {
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
