package scenes

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/renderer"
	"github.com/Xinrea/ffreplay/internal/system"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type PlayGroundScene struct {
	BasicScene
}

func NewPlayGroundScene() *PlayGroundScene {
	ecs := ecs.NewECS(donburi.NewWorld())
	entry.SetContext(ecs)

	system := system.NewSystem()
	renderer := renderer.NewRenderer()
	ui := ui.NewPlaygroundUI(ecs)

	globalEntry := entry.NewGlobal()
	global := component.Global.Get(globalEntry)
	cameraEntry := entry.NewCamera()
	camera := component.Camera.Get(cameraEntry)

	entry.NewMap(nil)

	ms := &PlayGroundScene{
		BasicScene: BasicScene{
			ecs:      ecs,
			system:   system,
			renderer: renderer,
			ui:       ui,
			global:   global,
			camera:   camera,
		},
	}

	ms.system.Init(ms.ecs)
	ms.renderer.Init(ms.ecs)

	m := model.MapCache[77]
	config := m.Load()
	current := config.Maps[config.CurrentMap]
	ms.camera.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)
	component.Map.Get(component.Map.MustFirst(ms.ecs.World)).Config = config

	global.Loaded.Store(true)

	return ms
}
