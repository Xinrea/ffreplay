package renderer

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) UIRender(ecs *ecs.ECS, screen *ebiten.Image) {
	// render debug info
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	x, y := ebiten.CursorPosition()
	wx, wy := camera.ScreenToWorld(float64(x), float64(y))
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor: %f, %f", wx, wy), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tick: %d, TPS: %.2f, FPS: %.2f", entry.GetTick(ecs), ebiten.ActualTPS(), ebiten.ActualFPS()), 0, 15)

	memberList := []*model.StatusData{}
	tag.PartyMember.Each(ecs.World, func(e *donburi.Entry) {
		sprite := component.Sprite.Get(e)
		if !sprite.Initialized {
			return
		}
		member := component.Status.Get(e)
		memberList = append(memberList, member)
	})
	r.PartyList.Render(entry.GetTick(ecs), screen, memberList)

	// render boss health bar
	cnt := 0
	w, h := camera.WindowSize()
	for e := range tag.Enemy.Iter(ecs.World) {
		enemy := component.Status.Get(e)
		if !enemy.IsBoss {
			continue
		}
		gap := 60
		percent := float64(enemy.HP) / float64(enemy.MaxHP)
		healthLeft := w - float64(r.EnemyHealthBar.w) - 50
		healthRight := healthLeft + float64(r.EnemyHealthBar.w)
		r.EnemyHealthBar.Render(screen, healthLeft, float64(40+cnt*gap), percent)
		if enemy.Casting != nil {
			p := float64(util.TickToMS(entry.GetTick(ecs)-enemy.Casting.ApplyTick)) / float64(enemy.Casting.Duration)
			DrawText(screen, fmt.Sprintf("[%d]%s", enemy.Casting.Ability.Guid, enemy.Casting.Ability.Name), 7, healthRight, float64(10+cnt*gap), r.EnemyHealthBar.Color, AlignRight)
			r.EnemyCasting.Render(screen, healthRight-float64(r.EnemyCasting.w), float64(30+cnt*gap), p)
		}
		DrawText(screen, enemy.Name, 7, healthLeft, float64(20+cnt*gap), r.EnemyHealthBar.Color, AlignLeft)
		DrawText(screen, fmt.Sprintf("HP: %d/%d", enemy.HP, enemy.MaxHP), 7, healthLeft, float64(50+cnt*gap), r.EnemyHealthBar.Color, AlignLeft)
		DrawText(screen, fmt.Sprintf("%.2f%%", percent*100), 7, healthRight, float64(50+cnt*gap), r.EnemyHealthBar.Color, AlignRight)
		cnt++
	}

	if !global.Loaded.Load() {
		DrawFilledRect(screen, 0, 0, w, h, color.RGBA{0, 0, 0, 128})
		DrawText(screen, fmt.Sprintf("预处理中: %d", global.LoadCount.Load()), 14, w/2, h/2, color.White, AlignCenter)
	}

	// render play progress
	current := float64(entry.GetTick(ecs)) / 60
	p := 0.0
	if global.FightDuration.Load() > 0 {
		p = current / float64(global.FightDuration.Load())
	}
	DrawText(screen, fmt.Sprintf("%.1fs / %.1fs", current, float64(global.FightDuration.Load())/1000), 7, w-30, h-120, color.White, AlignRight)
	r.PlayProgress.Render(screen, w-float64(r.PlayProgress.w)-30, h-100, p)

	// Draw shortkey prompt
	DrawText(screen, fmt.Sprintf("当前播放速度: %.1f", float64(entry.GetSpeed(ecs))/10.0), 7, w-30, h-90, color.White, AlignRight)
	DrawText(screen, "快退: 方向键左 | 快进: 方向键右", 7, w-30, h-70, color.White, AlignRight)
	DrawText(screen, "移动视角 W/A/S/D | 旋转视角: E/Q", 7, w-30, h-50, color.White, AlignRight)
	DrawText(screen, "暂停: SPACE | 播放速度: 方向键（上下）| 回到开始: R", 7, w-30, h-30, color.White, AlignRight)
}
