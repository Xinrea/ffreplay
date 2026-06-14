package renderer

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/layer"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignRight
	AlignCenter
)

type Renderer struct{}

func NewRenderer() *Renderer {
	initBackground()

	return &Renderer{}
}

var stackTextShadowOpt = &ui.ShadowOpt{
	Offset: 4,
	Color:  color.NRGBA{0, 0, 0, 200},
}

func RenderBuffList(canvas *ebiten.Image, buffs []*ui.UIBuff, x, y float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	// render buff icons
	for i, buff := range buffs {
		iconX := x + float64((i+1)*25)*s
		iconW := float64(ui.BuffWidth) * s
		iconH := float64(ui.BuffHeight) * s
		ui.TrackBuffTooltip(buff, image.Rect(
			int(iconX-iconW/2),
			int(y),
			int(iconX+iconW/2),
			int(y+iconH),
		))

		iconTexture := buff.Texture()
		geoM := texture.CenterGeoM(iconTexture)
		geoM.Scale(s, s)
		geoM.Translate(x+float64((i+1)*25)*s, y)
		canvas.DrawImage(iconTexture, &ebiten.DrawImageOptions{GeoM: geoM})

		if buff.Remain > 0 {
			ui.DrawText(
				canvas,
				formatSeconds(buff.Remain),
				14*s,
				x+float64((i+1)*25)*s,
				y+14*s,
				color.White,
				ui.AlignCenter,
				textShdowOpt,
			)
		}

		if buff.Stacks > 1 {
			ui.DrawText(
				canvas,
				strconv.Itoa(buff.Stacks),
				13*s,
				x+float64((i+1)*25)*s+6*s,
				y-7*s,
				color.White,
				ui.AlignCenter,
				stackTextShadowOpt,
			)
		}
	}
}

func (r *Renderer) Init(ecs *ecs.ECS) {
	ecs.AddRenderer(layer.Background, r.BackgroundRender)
	ecs.AddRenderer(layer.SkillRange, r.RangeRender)
	ecs.AddRenderer(layer.SkillRange, r.TelegraphRender)
	ecs.AddRenderer(layer.Background, r.WorldMarkerRender)
	ecs.AddRenderer(layer.Player, r.EnemyRender)
	ecs.AddRenderer(layer.Player, r.PlayerRender)
	ecs.AddRenderer(layer.Player, r.HeadMarkerRender)
	ecs.AddRenderer(layer.Player, r.SelectionRender)
	ecs.AddRenderer(layer.UI, r.UIRender)
}

func formatSeconds(seconds int64) string {
	minutes := seconds / 60
	hours := minutes / 60

	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%d", seconds)
}
