package renderer

import (
	_ "embed"
	"fmt"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/layer"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignRight
	AlignCenter
)

type Renderer struct {
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func RenderBuffList(canvas *ebiten.Image, tick int64, buffs []*model.Buff, x, y float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	// render buff icons
	for i, buff := range buffs {
		iconTexture := buff.Texture()
		geoM := texture.CenterGeoM(iconTexture)
		geoM.Scale(s, s)
		geoM.Translate(x+float64((i+1)*25)*s, y)
		canvas.DrawImage(iconTexture, &ebiten.DrawImageOptions{GeoM: geoM})
		if buff.Remain > 0 {
			ui.DrawText(canvas, formatSeconds(buff.Remain), 14*s, x+float64((i+1)*25)*s, y+12*s, color.White, furex.AlignItemCenter, textShdowOpt)
		}
		if buff.Stacks > 1 {
			canvas.DrawImage(model.BuffStackBG, &ebiten.DrawImageOptions{GeoM: geoM})
			ui.DrawText(canvas, strconv.Itoa(buff.Stacks), 10*s, x+float64((i+1)*25)*s+6*s, y-16*s, color.NRGBA{0, 0, 0, 255}, furex.AlignItemCenter, textShdowOpt)
		}
	}
}

func (r *Renderer) Init(ecs *ecs.ECS) {
	ecs.AddRenderer(layer.Background, r.BackgroundRender)
	ecs.AddRenderer(layer.SkillRange, r.RangeRender)
	ecs.AddRenderer(layer.Background, r.WorldMarkerRender)
	ecs.AddRenderer(layer.Player, r.EnemyRender)
	ecs.AddRenderer(layer.Player, r.PlayerRender)
	ecs.AddRenderer(layer.UI, r.UIRender)
}

func formatSeconds(seconds int64) string {
	minutes := seconds / 60
	hours := minutes / 60
	if hours > 0 {
		return fmt.Sprintf("%d时", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分", minutes)
	}
	return fmt.Sprintf("%d", seconds)
}
