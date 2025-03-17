package scenes

import (
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/renderer"
	"github.com/Xinrea/ffreplay/internal/system"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

type FallOfFaithScene struct {
	ecs      *ecs.ECS
	system   *system.System
	renderer *renderer.Renderer
	ui       ui.UI
	global   *model.GlobalData
	camera   *model.CameraData
	screenW  int
	screenH  int
}

func NewFallOfFaithScene() *FallOfFaithScene {
	ecs := ecs.NewECS(donburi.NewWorld())
	system := system.NewSystem()
	renderer := renderer.NewRenderer()
	ui := ui.NewPlaygroundUI(ecs)

	globalEntry := entry.NewGlobal(ecs)
	global := component.Global.Get(globalEntry)
	cameraEntry := entry.NewCamera(ecs)
	camera := component.Camera.Get(cameraEntry)

	entry.NewMap(ecs, nil)

	ms := &FallOfFaithScene{
		ecs:      ecs,
		system:   system,
		renderer: renderer,
		ui:       ui,
		global:   global,
		camera:   camera,
	}

	ms.setup()

	global.Loaded.Store(true)

	return ms
}

func (ms *FallOfFaithScene) setup() {
	ms.system.Init(ms.ecs)
	ms.renderer.Init(ms.ecs)

	m := model.MapCache[77]
	config := m.Load()
	current := config.Maps[config.CurrentMap]
	ms.camera.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)
	component.Map.Get(component.Map.MustFirst(ms.ecs.World)).Config = config

	defaultPlayer := entry.NewPlayer(ms.ecs, role.H2, f64.Vec2{current.Offset.X * 25, current.Offset.Y * 25}, nil)
	playerStatus := component.Status.Get(defaultPlayer)
	playerStatus.AddHeadMarker(model.HeadMarkerType1)

	entry.GetGlobal(ms.ecs).TargetPlayer = defaultPlayer

	// create a dummy enemy
	enemy := entry.NewEnemy(ms.ecs, f64.Vec2{current.Offset.X * 25, current.Offset.Y * 25}, 1.0, 0, -1, "dummy", true, 1)
	enemyStatus := component.Status.Get(enemy)
	enemyStatus.Charater = texture.NewTextureFromFile("asset/boss/1.png")
}

func (ms *FallOfFaithScene) Reset() {
	ms.system.Reset()
}

func (ms *FallOfFaithScene) Update() {
	ms.ecs.Update()
	ms.ui.Update(ms.screenW, ms.screenH)
}

func (ms *FallOfFaithScene) Layout(w, h int) {
	ms.system.Layout(w, h)
	ms.screenW = w
	ms.screenH = h
}

func (ms *FallOfFaithScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	ms.ecs.Draw(screen)
	ms.ui.Draw(screen)
}
