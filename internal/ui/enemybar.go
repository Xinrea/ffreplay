package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

const (
	EnemyBarWidth        = 500
	EnemyHeaderHeight    = 18
	EnemyHPBarHeight     = 10
	EnemyRowSpacing      = 5
	EnemyNameTextSize    = 13
	EnemyPercentTextSize = 13
	EnemyCastBarHeight   = 12
)

var enemyNameColor = color.NRGBA{252, 183, 190, 255}

func EUIEnemyBarsView(scale float64) *widget.Container {
	if scale <= 0 {
		scale = 1
	}

	pad := int(float64(UIPadding) * scale)
	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:   pad,
				Right: pad,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	view.GetWidget().OnUpdate = func(w widget.HasWidget) {
		view.RemoveChildren()
		for e := range tag.Enemy.Iter(ecsInstance.World) {
			sprite := component.Sprite.Get(e)
			if !sprite.Initialized {
				continue
			}
			enemy := component.Status.Get(e)
			if (enemy.Role != role.Boss && enemy.Role != role.Special) ||
				!sprite.Instances[0].IsActive(entry.GetTick(ecsInstance)) {
				continue
			}
			view.AddChild(newEUIEnemyBar(e, scale))
		}
	}

	return view
}

type euiEnemyBar struct {
	widget *widget.Widget
	enemy  *donburi.Entry
	scale  float64
}

func newEUIEnemyBar(enemy *donburi.Entry, scale float64) *euiEnemyBar {
	if scale <= 0 {
		scale = 1
	}

	b := &euiEnemyBar{
		enemy: enemy,
		scale: scale,
	}
	b.widget = widget.NewWidget(
		widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
		}),
	)
	return b
}

func (b *euiEnemyBar) GetWidget() *widget.Widget {
	return b.widget
}

func (b *euiEnemyBar) PreferredSize() (int, int) {
	s := b.scale
	header := float64(EnemyHeaderHeight) * s
	hp := float64(EnemyHPBarHeight) * s
	spacing := float64(EnemyRowSpacing) * s
	buffH := float64(BuffHeight+BuffRemainFontSize+BuffRemainTop) * s
	if buffH < float64(BuffHeight)*s {
		buffH = float64(BuffHeight) * s
	}

	return int(float64(EnemyBarWidth) * s),
		int(header + spacing + hp + spacing + buffH)
}

func (b *euiEnemyBar) SetLocation(rect image.Rectangle) {
	b.widget.Rect = rect
}

func (b *euiEnemyBar) Validate() {}

func (b *euiEnemyBar) Update(updObj *widget.UpdateObject) {
	b.widget.Update(updObj)
}

func (b *euiEnemyBar) Render(screen *ebiten.Image) {
	b.widget.Render(screen)

	status := component.Status.Get(b.enemy)
	sprite := component.Sprite.Get(b.enemy)
	frame := b.widget.Rect
	x := float64(frame.Min.X)
	y := float64(frame.Min.Y)
	s := b.scale
	barW := float64(EnemyBarWidth) * s

	cast := sprite.Instances[0].GetCast()
	b.drawHeader(screen, status, cast, x, y, barW)

	hpY := y + float64(EnemyHeaderHeight+EnemyRowSpacing)*s
	hpProgress := 0.0
	if status.MaxHP > 0 {
		hpProgress = float64(status.HP) / float64(status.MaxHP)
	}
	drawNineSliceBar(
		screen,
		image.Rect(int(x), int(hpY), int(x+barW), int(hpY+float64(EnemyHPBarHeight)*s)),
		barAtlas.GetNineSlice("red_bar_bg.png"),
		barAtlas.GetNineSlice("red_bar_fg.png"),
		hpProgress,
		nil,
	)

	buffX := x
	buffY := hpY + float64(EnemyHPBarHeight+EnemyRowSpacing)*s
	for i, buff := range UIBuffsFor(status.BuffList) {
		b.drawBuff(screen, buff, buffX+float64(i*BuffWidth)*s, buffY)
	}
}

func (b *euiEnemyBar) drawHeader(screen *ebiten.Image, status *model.StatusData, cast *model.Skill, x, y, barW float64) {
	s := b.scale
	headerH := float64(EnemyHeaderHeight) * s
	percent := 0.0
	if status.MaxHP > 0 {
		percent = float64(status.HP) / float64(status.MaxHP) * 100
	}
	percentText := fmt.Sprintf("%.1f%%", percent)
	percentW, _ := measureText(percentText, EnemyPercentTextSize*s)
	percentPad := 4 * s

	if cast != nil {
		b.drawCastBar(screen, cast, x, y, barW, headerH)
		return
	}

	DrawText(
		screen,
		percentText,
		EnemyPercentTextSize*s,
		x,
		y+headerH/2,
		enemyNameColor,
		AlignStart,
		&ShadowOpt{Color: color.NRGBA{0, 0, 0, 128}, Offset: 1 * s},
	)
	DrawText(
		screen,
		status.Name,
		EnemyNameTextSize*s,
		x+percentW+percentPad,
		y+headerH/2,
		enemyNameColor,
		AlignStart,
		&ShadowOpt{Color: color.NRGBA{0, 0, 0, 128}, Offset: 1 * s},
	)
}

func (b *euiEnemyBar) drawCastBar(screen *ebiten.Image, cast *model.Skill, x, y, barW, headerH float64) {
	s := b.scale
	barH := float64(EnemyCastBarHeight) * s
	if barH > headerH {
		barH = headerH
	}
	castY := y + (headerH-barH)/2
	castProgress := 0.0
	if cast.Cast > 0 {
		castProgress = float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast)
	}

	drawNineSliceBar(
		screen,
		image.Rect(int(x), int(castY), int(x+barW), int(castY+barH)),
		castAtlas.GetNineSlice("casting_frame.png"),
		castAtlas.GetNineSlice("casting_fg.png"),
		castProgress,
		nil,
	)

	DrawText(
		screen,
		cast.Name,
		CastNameTextSize*s,
		x+barW,
		castY+barH/2,
		color.White,
		AlignEnd,
		&ShadowOpt{Color: color.NRGBA{240, 152, 0, 128}, Offset: 1 * s},
	)
}

func (b *euiEnemyBar) drawBuff(screen *ebiten.Image, buff *UIBuff, x, y float64) {
	s := b.scale
	TrackBuffTooltip(buff, BuffHitRect(x, y, s))
	b.drawScaled(screen, buff.Texture(), x, y, BuffWidth*s, BuffHeight*s)
	if buff.Stacks > 1 {
		DrawText(
			screen,
			strconv.Itoa(buff.Stacks),
			BuffStackFontSize*s,
			x+float64(BuffStackLeft+BuffStackFontSize)*s,
			y+float64(BuffStackTop)*s+BuffStackFontSize*s/2,
			color.White,
			AlignEnd,
			&ShadowOpt{Color: color.NRGBA{0, 0, 0, 200}, Offset: EUIBuffStackShadow * s},
		)
	}
	DrawText(
		screen,
		formatSeconds(buff.Remain),
		BuffRemainFontSize*s,
		x+BuffWidth*s/2,
		y+float64(BuffHeight+BuffRemainTop)*s,
		color.White,
		AlignCenter,
		&ShadowOpt{Color: color.NRGBA{0, 0, 0, 128}, Offset: 1 * s},
	)
}

func (b *euiEnemyBar) drawScaled(screen *ebiten.Image, img *ebiten.Image, x, y, w, h float64) {
	bounds := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/float64(bounds.Dx()), h/float64(bounds.Dy()))
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}
