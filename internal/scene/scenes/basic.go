package scenes

import (
	"image/color"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/renderer"
	"github.com/Xinrea/ffreplay/internal/system"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

type BasicScene struct {
	ecs      *ecs.ECS
	system   *system.System
	renderer *renderer.Renderer
	ui       ui.UI
	global   *model.GlobalData
	camera   *model.CameraData
	screenW  int
	screenH  int
}

func (bs *BasicScene) Reset() {
	bs.system.Reset()
}

func (bs *BasicScene) Update() {
	bs.ecs.Update()
	bs.ui.Update(bs.screenW, bs.screenH)
}

func (bs *BasicScene) Layout(w, h int) {
	bs.system.Layout(w, h)
	bs.screenW = w
	bs.screenH = h
}

func (bs *BasicScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	bs.ecs.Draw(screen)
	bs.ui.Draw(screen)
}
