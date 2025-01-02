package renderer

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

var textShdowOpt = &ui.ShadowOpt{Offset: 2, Color: color.NRGBA{0, 0, 0, 255}}

func (r *Renderer) UIRender(ecs *ecs.ECS, screen *ebiten.Image) {
	// render debug info
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	tick := global.Tick / 10
	s := ebiten.Monitor().DeviceScaleFactor()
	x, y := ebiten.CursorPosition()
	wx, wy := camera.ScreenToWorld(float64(x), float64(y))
	w, h := camera.WindowSize()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor: %f, %f, Debug: %t", wx, wy, global.Debug), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tick: %d, Time: %d, TPS: %.2f, FPS: %.2f", entry.GetTick(ecs), util.TickToMS(entry.GetTick(ecs)), ebiten.ActualTPS(), ebiten.ActualFPS()), 0, 15)

	if !global.Loaded.Load() {
		ui.DrawText(screen, fmt.Sprintf("预处理中: %d/%d", global.LoadCount.Load(), global.LoadTotal), 28*s, w/2*s, h/2*s, color.White, furex.AlignItemCenter, textShdowOpt)
		return
	}

	// render target player casting history
	if global.TargetPlayer != nil {
		player := component.Sprite.Get(global.TargetPlayer)
		casts := player.Instances[0].GetHistoryCast(tick)
		currentCasting := player.Instances[0].GetCast()
		if currentCasting != nil {
			casts = append(casts, currentCasting)
		}
		sk := NewSkillTimeline(casts)
		sk.Render(global.Debug, screen, w/2, h-200, tick)
	}
}
