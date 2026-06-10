package ui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func shortName(name string, limit int) string {
	runes := []rune(name)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}

	return name
}

type euiDamageHistory struct {
	widget      *widget.Widget
	scale       float64
	memberCount int
}

func NewEUIDamageHistoryView(memberCount int, scale float64) *euiDamageHistory {
	if scale <= 0 {
		scale = 1
	}
	view := &euiDamageHistory{
		scale:       scale,
		memberCount: memberCount,
	}
	partyHeight := memberCount*PlayerItemHeight + PartyListBGExtra
	view.widget = widget.NewWidget(
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding: &widget.Insets{
				Left: int(float64(UIPadding+10) * scale),
				Top:  int(float64(UIPadding+SingleBarHeight+10+partyHeight+40) * scale),
			},
		}),
	)
	return view
}

func (d *euiDamageHistory) GetWidget() *widget.Widget {
	return d.widget
}

func (d *euiDamageHistory) SetLocation(rect image.Rectangle) {
	d.widget.Rect = rect
}

func (d *euiDamageHistory) PreferredSize() (int, int) {
	return int(300 * d.scale), int(160 * d.scale)
}

func (d *euiDamageHistory) Validate() {}

func (d *euiDamageHistory) Update(updObj *widget.UpdateObject) {
	d.widget.Update(updObj)
}

func (d *euiDamageHistory) Render(screen *ebiten.Image) {
	d.widget.Render(screen)

	global := entry.GetGlobal(ecsInstance)
	if global.TargetPlayer == nil {
		return
	}

	frame := d.widget.Rect
	bg := texture.NewNineSlice(
		texture.NewTextureFromFile("asset/partylist_bg.png"),
		PartyListBGNineSliceConfig[0],
		PartyListBGNineSliceConfig[1],
		PartyListBGNineSliceConfig[2],
		PartyListBGNineSliceConfig[3],
	)
	bgFrame := frame
	bgFrame.Min.X -= int(40 * d.scale)
	bgFrame.Min.Y -= int(20 * d.scale)
	bgFrame.Max.Y += int(10 * d.scale)
	bg.Draw(screen, bgFrame, nil)

	x := float64(frame.Min.X)
	y := float64(frame.Min.Y)
	d.drawHeader(screen, x, y)

	instance := component.Sprite.Get(global.TargetPlayer).Instances[0]
	for i, damage := range instance.GetHistoryDamageTaken(5) {
		d.drawItem(screen, damage, x, y+float64(20+i*28)*d.scale)
	}
}

func (d *euiDamageHistory) drawHeader(screen *ebiten.Image, x, y float64) {
	d.drawText(screen, "时间点", x, y+7*d.scale, AlignStart)
	d.drawText(screen, "伤害名", x+80*d.scale, y+7*d.scale, AlignStart)
	d.drawText(screen, "最终伤害量", x+180*d.scale, y+7*d.scale, AlignStart)
	d.drawText(screen, "减伤", x+260*d.scale, y+7*d.scale, AlignStart)
}

func (d *euiDamageHistory) drawItem(screen *ebiten.Image, damage model.DamageTaken, x, y float64) {
	d.drawText(screen, formatDuration(float64(util.TickToMS(damage.Tick))/1000), x, y+7*d.scale, AlignStart)

	iconPath := "asset/ui/d_physical.png"
	switch damage.Type {
	case model.Physical:
		iconPath = "asset/ui/d_physical.png"
	case model.Magical:
		iconPath = "asset/ui/d_magical.png"
	case model.Special:
		iconPath = "asset/ui/d_special.png"
	default:
		log.Println("Unknown damage type:", damage)
	}
	d.drawScaled(screen, texture.NewTextureFromFile(iconPath), x+80*d.scale, y, 14*d.scale, 14*d.scale)
	d.drawText(screen, shortName(damage.Ability.Name, 10), x+96*d.scale, y+7*d.scale, AlignStart)
	d.drawText(screen, fmt.Sprintf("%d", damage.Amount), x+180*d.scale, y+7*d.scale, AlignStart)
	d.drawText(screen, fmt.Sprintf("%.0f%%", (1-damage.Multiplier)*100), x+260*d.scale, y+7*d.scale, AlignStart)

	buffX := x + 305*d.scale
	for i, b := range damage.RelatedBuffs {
		buff := &model.Buff{
			ID:   b.ID,
			Name: b.Name,
			Icon: b.Icon,
		}
		d.drawBuff(screen, buff, buffX+float64(i*BuffWidth)*d.scale, y-7*d.scale)
	}
}

func (d *euiDamageHistory) drawText(screen *ebiten.Image, content string, x, y float64, align TextAlign) {
	DrawText(screen, content, 14*d.scale, x, y, color.White, align,
		&ShadowOpt{Color: color.NRGBA{22, 45, 87, 128}, Offset: 2 * d.scale})
}

func (d *euiDamageHistory) drawBuff(screen *ebiten.Image, buff *model.Buff, x, y float64) {
	d.drawScaled(screen, buff.Texture(), x, y, BuffWidth*d.scale, BuffHeight*d.scale)
	if buff.Stacks > 1 {
		DrawText(screen, strconv.Itoa(buff.Stacks), BuffStackFontSize*d.scale, x+float64(BuffStackLeft+BuffStackFontSize)*d.scale, y+float64(BuffStackTop)*d.scale+BuffStackFontSize*d.scale/2, color.White, AlignEnd,
			&ShadowOpt{Color: color.NRGBA{0, 0, 0, 200}, Offset: EUIBuffStackShadow * d.scale})
	}
	DrawText(screen, formatSeconds(buff.Remain), BuffRemainFontSize*d.scale, x+BuffWidth*d.scale/2, y+float64(BuffHeight+BuffRemainTop)*d.scale, color.White, AlignCenter,
		&ShadowOpt{Color: color.NRGBA{0, 0, 0, 128}, Offset: 1 * d.scale})
}

func (d *euiDamageHistory) drawScaled(screen *ebiten.Image, img *ebiten.Image, x, y, w, h float64) {
	bounds := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/float64(bounds.Dx()), h/float64(bounds.Dy()))
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}
