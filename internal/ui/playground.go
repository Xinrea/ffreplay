package ui

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

type PlaygroundUI struct {
	eui       *ebitenui.UI
	euiRoot   *widget.Container
	propPanel *PropertyPanelEUI
	once      sync.Once
}

var _ UI = (*PlaygroundUI)(nil)

func NewPlaygroundUI(ecs *ecs.ECS) *PlaygroundUI {
	ecsInstance = ecs

	// Shared ebitenui root: one UI instance for ALL ebitenui components in this scene.
	euiRoot := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	eui := &ebitenui.UI{Container: euiRoot}

	return &PlaygroundUI{
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

	global := entry.GetGlobal(ecsInstance)
	if global.Loaded.Load() {
		p.once.Do(func() {
			p.euiRoot.AddChild(NewEUIPlaygroundPartyList(s))
			p.euiRoot.AddChild(NewEUICommandView(s))
		})
	}

	// Recompute focus from the current ebitenui state each frame. Focused
	// widgets below will set this back to true during/after p.eui.Update().
	global.UIFocus = false
	global.UIHovered = false

	SyncBuffLists(entry.GetTick(ecsInstance))
	BeginBuffTooltipFrame()

	// Single ebitenui update covers all ebitenui-managed components.
	p.eui.Update()
	p.propPanel.UpdateECS(w, h, s)
	SyncEUIInputState(global, p.eui)
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {
	// Single ebitenui draw for all migrated components.
	p.eui.Draw(screen)
	DrawBuffTooltip(screen, ebiten.Monitor().DeviceScaleFactor())

	// Apply playground grab/move cursor after ebitenui, which resets the shape each frame.
	if entry.GetGlobal(ecsInstance).PlaygroundMoveCursor {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)
	}
}
