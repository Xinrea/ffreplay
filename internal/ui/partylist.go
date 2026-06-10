package ui

import (
	"image"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	euiinput "github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

const (
	PartyListWidth   = 300
	PartyListBGExtra = 20
)

var PartyListBGNineSliceConfig = [4]int{5, 14, 0, 0}

const (
	PlayerItemHeight = 48
	StatusPartWidth  = 210
	JobIconSize      = 38

	HoverSpriteWidth  = 275
	HoverSpriteHeight = 40
	HoverSpriteTop    = 10
	HoverSpriteLeft   = 30

	CastNameTextSize = 12
	NameTextSize     = 13
	HMPTextSize      = 14

	CastBarTop    = 5
	CastBarHeight = 12
	NameTop       = 14
	HPBarTop      = 26

	HPBarWidth   = 125
	MPBarWidth   = 75
	HMPBarHeight = 8

	BuffListOffsetY = 20
)

type euiPartyList struct {
	widget  *widget.Widget
	players []*donburi.Entry
	entries func() []*donburi.Entry
	scale   float64
	hovered int
	top     int
	left    int
}

func NewEUIReplayPartyList(players []*donburi.Entry, scale float64) *euiPartyList {
	return newEUIPartyList(players, nil, UIPadding+SingleBarHeight+10, UIPadding, scale)
}

func NewEUIPlaygroundPartyList(scale float64) *euiPartyList {
	return newEUIPartyList(nil, currentPlayerEntries, 40, UIPadding, scale)
}

func newEUIPartyList(players []*donburi.Entry, entries func() []*donburi.Entry, top, left int, scale float64) *euiPartyList {
	if scale <= 0 {
		scale = 1
	}
	pl := &euiPartyList{
		players: players,
		entries: entries,
		scale:   scale,
		hovered: -1,
		top:     top,
		left:    left,
	}
	pl.widget = widget.NewWidget(
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding: &widget.Insets{
				Top:  int(float64(top) * scale),
				Left: int(float64(left) * scale),
			},
		}),
		widget.WidgetOpts.CursorHovered(euiinput.CURSOR_POINTER),
		widget.WidgetOpts.CursorMoveHandler(func(args *widget.WidgetCursorMoveEventArgs) {
			pl.hovered = pl.indexAt(args.OffsetY)
		}),
		widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
			pl.hovered = pl.indexAt(args.OffsetY)
		}),
		widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
			pl.hovered = -1
		}),
		widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
			if args.Button != ebiten.MouseButtonLeft {
				return
			}
			players := pl.currentPlayers()
			idx := pl.indexAt(args.OffsetY)
			if idx >= 0 && idx < len(players) {
				entry.GetGlobal(ecsInstance).TargetPlayer = players[idx]
			}
		}),
	)
	return pl
}

func currentPlayerEntries() []*donburi.Entry {
	players := []*donburi.Entry{}
	for p := range tag.Player.Iter(ecsInstance.World) {
		if !p.Valid() {
			continue
		}
		players = append(players, p)
	}
	return players
}

func (p *euiPartyList) currentPlayers() []*donburi.Entry {
	if p.entries != nil {
		return p.entries()
	}
	return p.players
}

func (p *euiPartyList) indexAt(offsetY int) int {
	idx := int(float64(offsetY) / (float64(PlayerItemHeight) * p.scale))
	if idx < 0 || idx >= len(p.currentPlayers()) {
		return -1
	}
	return idx
}

func (p *euiPartyList) GetWidget() *widget.Widget {
	return p.widget
}

func (p *euiPartyList) SetLocation(rect image.Rectangle) {
	p.widget.Rect = rect
}

func (p *euiPartyList) PreferredSize() (int, int) {
	players := p.currentPlayers()
	return int(float64(PartyListWidth) * p.scale),
		int(float64(len(players)*PlayerItemHeight+PartyListBGExtra) * p.scale)
}

func (p *euiPartyList) Validate() {}

func (p *euiPartyList) Update(updObj *widget.UpdateObject) {
	p.widget.Update(updObj)
}

func (p *euiPartyList) Render(screen *ebiten.Image) {
	p.widget.Render(screen)

	frame := p.widget.Rect
	players := p.currentPlayers()
	if p.hovered >= len(players) {
		p.hovered = -1
	}
	bg := texture.NewNineSlice(
		texture.NewTextureFromFile("asset/partylist_bg.png"),
		PartyListBGNineSliceConfig[0],
		PartyListBGNineSliceConfig[1],
		PartyListBGNineSliceConfig[2],
		PartyListBGNineSliceConfig[3],
	)
	bg.Draw(screen, frame, nil)

	for i, player := range players {
		p.renderPlayer(screen, player, i)
	}
}

func (p *euiPartyList) renderPlayer(screen *ebiten.Image, playerEntry *donburi.Entry, index int) {
	status := component.Status.Get(playerEntry)
	sprite := component.Sprite.Get(playerEntry)
	s := p.scale
	x := float64(p.widget.Rect.Min.X)
	y := float64(p.widget.Rect.Min.Y) + float64(index*PlayerItemHeight)*s

	if p.hovered == index {
		p.drawScaled(screen, texture.NewTextureFromFile("asset/partylist_hover.png"), x+HoverSpriteLeft*s, y+HoverSpriteTop*s, HoverSpriteWidth*s, HoverSpriteHeight*s)
	}
	if entry.GetGlobal(ecsInstance).TargetPlayer == playerEntry {
		p.drawScaled(screen, texture.NewTextureFromFile("asset/partylist_selected.png"), x+HoverSpriteLeft*s, y+HoverSpriteTop*s, HoverSpriteWidth*s, HoverSpriteHeight*s)
	}

	p.drawScaled(screen, status.RoleTexture(), x, y+5*s, JobIconSize*s, JobIconSize*s)

	statusX := x + float64(JobIconSize+5)*s
	cast := sprite.Instances[0].GetCast()
	hpY := y + HPBarTop*s

	if cast != nil {
		p.drawCastBar(screen, cast, statusX, y+float64(CastBarTop)*s)
	} else {
		DrawText(screen, status.Name, NameTextSize*s, statusX, y+float64(NameTop)*s, color.White, AlignStart,
			&ShadowOpt{Color: color.NRGBA{22, 45, 87, 128}, Offset: 2 * s})
	}

	hpProgress := 0.0
	if status.MaxHP > 0 {
		hpProgress = float64(status.HP) / float64(status.MaxHP)
	}
	drawNineSliceBar(
		screen,
		image.Rect(int(statusX), int(hpY), int(statusX+HPBarWidth*s), int(hpY+HMPBarHeight*s)),
		barAtlas.GetNineSlice("normal_bar_bg.png"),
		barAtlas.GetNineSlice("normal_bar_fg.png"),
		hpProgress,
		nil,
	)
	DrawText(screen, strconv.Itoa(status.HP), HMPTextSize*s, statusX+HPBarWidth*s, hpY+15*s, color.White, AlignEnd,
		&ShadowOpt{Color: color.NRGBA{22, 45, 87, 128}, Offset: 2 * s})

	mpX := statusX + (HPBarWidth+10)*s
	mpProgress := 0.0
	if status.MaxMana > 0 {
		mpProgress = float64(status.Mana) / float64(status.MaxMana)
	}
	drawNineSliceBar(
		screen,
		image.Rect(int(mpX), int(hpY), int(mpX+MPBarWidth*s), int(hpY+HMPBarHeight*s)),
		barAtlas.GetNineSlice("normal_bar_bg.png"),
		barAtlas.GetNineSlice("normal_bar_fg.png"),
		mpProgress,
		nil,
	)
	DrawText(screen, strconv.Itoa(status.Mana), HMPTextSize*s, mpX+MPBarWidth*s, hpY+15*s, color.White, AlignEnd,
		&ShadowOpt{Color: color.NRGBA{22, 45, 87, 128}, Offset: 2 * s})

	buffX := statusX + StatusPartWidth*s + 5*s
	buffY := y + BuffListOffsetY*s
	for i, buff := range UIBuffsFor(status.BuffList) {
		p.drawBuff(screen, buff, buffX+float64(i*BuffWidth)*s, buffY)
	}
}

func (p *euiPartyList) drawCastBar(screen *ebiten.Image, cast *model.Skill, statusX, castBarY float64) {
	s := p.scale
	barH := float64(CastBarHeight) * s
	barW := float64(StatusPartWidth) * s
	castProgress := 0.0
	if cast.Cast > 0 {
		castProgress = float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast)
	}

	drawNineSliceBar(
		screen,
		image.Rect(int(statusX), int(castBarY), int(statusX+barW), int(castBarY+barH)),
		castAtlas.GetNineSlice("casting_frame.png"),
		castAtlas.GetNineSlice("casting_fg.png"),
		castProgress,
		nil,
	)

	textX := statusX + barW
	textY := castBarY + barH/2
	DrawText(screen, cast.Name, CastNameTextSize*s, textX, textY, color.White, AlignEnd,
		&ShadowOpt{Color: color.NRGBA{240, 152, 0, 128}, Offset: 1 * s})
}

func (p *euiPartyList) drawBuff(screen *ebiten.Image, buff *UIBuff, x, y float64) {
	s := p.scale
	TrackBuffTooltip(buff, BuffHitRect(x, y, s))
	p.drawScaled(screen, buff.Texture(), x, y, BuffWidth*s, BuffHeight*s)
	if buff.Stacks > 1 {
		DrawText(screen, strconv.Itoa(buff.Stacks), BuffStackFontSize*s, x+float64(BuffStackLeft+BuffStackFontSize)*s, y+float64(BuffStackTop)*s+BuffStackFontSize*s/2, color.White, AlignEnd,
			&ShadowOpt{Color: color.NRGBA{0, 0, 0, 200}, Offset: EUIBuffStackShadow * s})
	}
	DrawText(screen, formatSeconds(buff.Remain), BuffRemainFontSize*s, x+BuffWidth*s/2, y+float64(BuffHeight+BuffRemainTop)*s, color.White, AlignCenter,
		&ShadowOpt{Color: color.NRGBA{0, 0, 0, 128}, Offset: 1 * s})
}

func (p *euiPartyList) drawScaled(screen *ebiten.Image, img *ebiten.Image, x, y, w, h float64) {
	bounds := img.Bounds()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/float64(bounds.Dx()), h/float64(bounds.Dy()))
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}
