package renderer

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

const PARTY_MEMBER_GAP = 40

type PartyList struct {
	x        float64
	y        float64
	progress *ProgressBar
}

func NewPartyList(x, y float64) *PartyList {
	return &PartyList{
		x:        x,
		y:        y,
		progress: NewProgressBar(150, 6, color.NRGBA{230, 255, 255, 255}),
	}
}

func (p *PartyList) Render(tick int64, canvas *ebiten.Image, members []*model.StatusData) {
	for i, member := range members {
		p.renderOne(tick, canvas, member, f64.Vec2{p.x, p.y + float64(i*PARTY_MEMBER_GAP)})
	}
}

func (p *PartyList) renderOne(tick int64, canvas *ebiten.Image, member *model.StatusData, pos f64.Vec2) {
	s := ebiten.Monitor().DeviceScaleFactor()
	// render icon
	iconTexture := member.RoleTexture()
	m := iconTexture.GetGeoM()
	m.Scale(0.6, 0.6)
	m.Translate(pos[0], pos[1])
	m.Scale(s, s)
	canvas.DrawImage(iconTexture.Img(), &ebiten.DrawImageOptions{GeoM: m})
	DrawText(canvas, member.Name, 7, pos[0]+25, pos[1]-15, color.White, AlignLeft)
	RenderBuffList(canvas, tick, member.BuffList.Buffs(), pos[0]+200, pos[1], s)
	// render HP bar
	progress := float64(member.HP) / float64(member.MaxHP)
	if progress > 1 {
		progress = 1
	}
	p.progress.Render(canvas, pos[0]+25, pos[1]+5, progress)
	DrawText(canvas, fmt.Sprintf("%d", member.HP), 7, pos[0]+175, pos[1]+10, color.White, AlignRight)
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
