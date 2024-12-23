package renderer

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

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
	tick := global.Tick / 10
	x, y := ebiten.CursorPosition()
	wx, wy := camera.ScreenToWorld(float64(x), float64(y))
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor: %f, %f", wx, wy), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tick: %d, Time: %d, TPS: %.2f, FPS: %.2f", entry.GetTick(ecs), util.TickToMS(entry.GetTick(ecs)), ebiten.ActualTPS(), ebiten.ActualFPS()), 0, 15)

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
		sprite := component.Sprite.Get(e)
		if !sprite.Initialized {
			continue
		}
		enemy := component.Status.Get(e)
		mainInstance := sprite.Instances[0]
		if enemy.Role != model.Boss || !mainInstance.IsActive(tick) {
			continue
		}
		gap := 100
		percent := float64(enemy.HP) / float64(enemy.MaxHP)
		healthLeft := w - float64(r.EnemyHealthBar.w) - 50
		healthRight := healthLeft + float64(r.EnemyHealthBar.w)
		r.EnemyHealthBar.Render(screen, healthLeft, float64(40+cnt*gap), percent)
		casting := mainInstance.GetCast()
		if casting != nil && casting.StartTick != -1 {
			castName := casting.Name
			if global.Debug {
				castName = fmt.Sprintf("[%d]%s", casting.ID, casting.Name)
			}
			p := float64(util.TickToMS(entry.GetTick(ecs)-casting.StartTick)) / float64(casting.Cast)
			DrawText(screen, castName, 14, healthRight, float64(10+cnt*gap), r.EnemyHealthBar.Color, AlignRight)
			r.EnemyCasting.Render(screen, healthRight-float64(r.EnemyCasting.w), float64(30+cnt*gap), p)
		}
		DrawText(screen, enemy.Name, 14, healthLeft, float64(20+cnt*gap), r.EnemyHealthBar.Color, AlignLeft)
		DrawText(screen, fmt.Sprintf("HP: %s / %s", formatInt(enemy.HP), formatInt(enemy.MaxHP)), 14, healthLeft, float64(50+cnt*gap), r.EnemyHealthBar.Color, AlignLeft)
		DrawText(screen, fmt.Sprintf("%.2f%%", percent*100), 14, healthRight, float64(50+cnt*gap), r.EnemyHealthBar.Color, AlignRight)
		// render buffs on boss
		RenderBuffList(screen, tick, enemy.BuffList.Buffs(), healthLeft+10, float64(80+cnt*gap), ebiten.Monitor().DeviceScaleFactor())
		cnt++
	}

	if !global.Loaded.Load() {
		DrawFilledRect(screen, 0, 0, w, h, color.RGBA{0, 0, 0, 128})
		DrawText(screen, fmt.Sprintf("预处理中: %d/%d", global.LoadCount.Load(), global.LoadTotal), 28, w/2, h/2, color.White, AlignCenter)
	}

	// render play progress
	current := float64(entry.GetTick(ecs)) / 60
	p := 0.0
	if global.FightDuration.Load() > 0 {
		p = current / (float64(global.FightDuration.Load()) / 1000)
	}
	DrawText(screen, fmt.Sprintf("%s / %s", formatDuration(current), formatDuration(float64(global.FightDuration.Load())/1000)), 14, w-30, h-120, color.White, AlignRight)
	r.PlayProgress.Render(screen, w-float64(r.PlayProgress.w)-30, h-100, p)
	for _, p := range global.Phases {
		seperator := float64(p) / float64(util.MSToTick(global.FightDuration.Load()))
		DrawFilledRect(screen, w-30-(1-seperator)*float64(r.PlayProgress.w), h-100, 4, float64(r.PlayProgress.h), color.NRGBA{188, 61, 136, 255})
	}

	// render target player casting history
	if global.TargetPlayer != nil {
		player := component.Sprite.Get(global.TargetPlayer)
		// using a area on middle bottom of screen, 600x40
		cx := w/2 + 300
		cy := h - 100
		DrawFilledRect(screen, w/2-350, h-200, 700, 150, color.NRGBA{0, 0, 0, 100})
		casts := player.Instances[0].GetHistoryCast(tick)
		currentCasting := player.Instances[0].GetCast()
		if currentCasting != nil {
			casts = append(casts, currentCasting)
		}
		for _, c := range casts {
			if c.ID == 7 {
				continue
			}
			delta := tick - c.StartTick
			if c.Cast > 0 {
				// render casting background
				DrawFilledRect(screen, cx-float64(delta), cy, float64(min(delta, util.MSToTick(c.Cast))), 10, color.NRGBA{230, 255, 255, 128})
			}
			RenderCasting(screen, tick, c, cx-float64(delta), cy)
		}
	}

	// Draw shortkey prompt
	DrawText(screen, fmt.Sprintf("当前播放速度: %.1f", float64(entry.GetSpeed(ecs))/10.0), 14, w-30, h-90, color.White, AlignRight)
	DrawText(screen, "快退: 方向键左 | 快进: 方向键右", 14, w-30, h-70, color.White, AlignRight)
	DrawText(screen, "移动视角 W/A/S/D | 旋转视角: E/Q | 调试模式：`", 14, w-30, h-50, color.White, AlignRight)
	DrawText(screen, "暂停: SPACE | 播放速度: 方向键（上下）| 回到开始: R", 14, w-30, h-30, color.White, AlignRight)

	// Draw player selection prompt
	DrawText(screen, "锁定玩家: 1-8 | 解除锁定: ESC", 14, 30, h-30, color.White, AlignLeft)
}

func formatDuration(s float64) string {
	minutes := int(s) / 60
	seconds := int(s) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatInt(n int) string {
	// 将 int64 转换为字符串
	str := strconv.FormatInt(int64(n), 10)

	// 计算整数的长度
	length := len(str)
	if length <= 3 {
		return str // 如果长度小于等于3，直接返回
	}

	// 使用 strings.Builder 来构建结果字符串
	var builder strings.Builder
	for i, digit := range str {
		// 每三位添加一个逗号
		if i != 0 && (length-i)%3 == 0 {
			builder.WriteRune(',')
		}
		builder.WriteRune(digit)
	}
	return builder.String()
}
