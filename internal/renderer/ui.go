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
		DrawText(screen, fmt.Sprintf("HP: %s / %s", formatInt(enemy.HP), formatInt(enemy.MaxHP)), 7, healthLeft, float64(50+cnt*gap), r.EnemyHealthBar.Color, AlignLeft)
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
		p = current / (float64(global.FightDuration.Load()) / 1000)
	}
	DrawText(screen, fmt.Sprintf("%s / %s", formatDuration(current), formatDuration(float64(global.FightDuration.Load())/1000)), 7, w-30, h-120, color.White, AlignRight)
	r.PlayProgress.Render(screen, w-float64(r.PlayProgress.w)-30, h-100, p)

	// Draw shortkey prompt
	DrawText(screen, fmt.Sprintf("当前播放速度: %.1f", float64(entry.GetSpeed(ecs))/10.0), 7, w-30, h-90, color.White, AlignRight)
	DrawText(screen, "快退: 方向键左 | 快进: 方向键右", 7, w-30, h-70, color.White, AlignRight)
	DrawText(screen, "移动视角 W/A/S/D | 旋转视角: E/Q", 7, w-30, h-50, color.White, AlignRight)
	DrawText(screen, "暂停: SPACE | 播放速度: 方向键（上下）| 回到开始: R", 7, w-30, h-30, color.White, AlignRight)
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
