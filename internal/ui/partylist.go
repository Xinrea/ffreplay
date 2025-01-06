package ui

import (
	"image"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

func NewPartyList(players []*donburi.Entry) *furex.View {
	view := furex.NewView(furex.TagName("PartyList"), furex.MarginTop(10), furex.Direction(furex.Column), furex.AlignItems(furex.AlignItemStart), furex.Justify(furex.JustifyStart))

	view.AddChild(furex.NewView(furex.TagName("PartyListBG"), furex.Width(300), furex.Height(48), furex.Position(furex.PositionAbsolute), furex.Handler(&Sprite{NineSliceTexture: texture.NewNineSlice(texture.NewTextureFromFile("asset/partylist_bg.png"), 5, 14, 0, 0)})))
	for _, p := range players {
		view.AddChild(NewPlayerItem(p))
	}
	return view
}

type PlayerItem struct {
	Player   *donburi.Entry
	Hovered  bool
	Selected bool

	handler furex.ViewHandler
}

var _ furex.HandlerProvider = (*PlayerItem)(nil)

func (p *PlayerItem) Handler() furex.ViewHandler {
	p.handler.Extra = p
	p.handler.Update = p.Update
	p.handler.MouseEnter = p.HandleMouseEnter
	p.handler.MouseLeave = p.HandleMouseLeave
	p.handler.JustPressedMouseButtonLeft = p.HandleJustPressedMouseButtonLeft
	p.handler.JustReleasedMouseButtonLeft = p.HandleJustReleasedMouseButtonLeft
	return p.handler
}

func (p *PlayerItem) Update(v *furex.View) {
	// status := component.Status.Get(p.Player)
	if p.Hovered {
		v.MustGetByID("hover").Attrs.Display = furex.DisplayFlex
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		v.MustGetByID("hover").Attrs.Display = furex.DisplayNone
	}
	targetPlayer := entry.GetGlobal(ecsInstance).TargetPlayer
	if targetPlayer == p.Player {
		v.MustGetByID("selected").Attrs.Display = furex.DisplayFlex
	} else {
		v.MustGetByID("selected").Attrs.Display = furex.DisplayNone
	}

	// if player is casting, hide name
	if component.Sprite.Get(p.Player).Instances[0].GetCast() != nil {
		v.MustGetByID("name").Attrs.Hidden = true
		v.MustGetByID("cast").Attrs.Hidden = false
	} else {
		v.MustGetByID("name").Attrs.Hidden = false
		v.MustGetByID("cast").Attrs.Hidden = true
	}
}

func (p *PlayerItem) HandleJustPressedMouseButtonLeft(_ image.Rectangle, x, y int) bool {
	entry.GetGlobal(ecsInstance).TargetPlayer = p.Player
	return true
}

func (p *PlayerItem) HandleJustReleasedMouseButtonLeft(_ image.Rectangle, x, y int) {
}

func (p *PlayerItem) HandleMouseEnter(x, y int) bool {
	p.Hovered = true
	return true
}

func (p *PlayerItem) HandleMouseLeave() {
	p.Hovered = false
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
}

func NewPlayerItem(playerEntry *donburi.Entry) *furex.View {
	item := &PlayerItem{
		Player: playerEntry,
	}
	player := component.Status.Get(playerEntry)
	view := furex.NewView(
		furex.ID(strconv.Itoa(int(player.ID))),
		furex.Height(48), furex.Direction(furex.Row),
		furex.AlignItems(furex.AlignItemCenter),
		furex.Justify(furex.JustifyStart),
		furex.Handler(item),
	)
	view.AddChild(furex.NewView(furex.ID("hover"), furex.Position(furex.PositionAbsolute), furex.Top(10), furex.Left(30), furex.Width(275), furex.Height(40), furex.Handler(&Sprite{Texture: texture.NewTextureFromFile("asset/partylist_hover.png")})))
	view.AddChild(furex.NewView(furex.ID("selected"), furex.Position(furex.PositionAbsolute), furex.Top(10), furex.Left(30), furex.Width(275), furex.Height(40), furex.Handler(&Sprite{Texture: texture.NewTextureFromFile("asset/partylist_selected.png")})))

	// add job icon
	view.AddChild(furex.NewView(furex.Width(38), furex.Height(38), furex.Handler(&Sprite{Texture: player.RoleTexture()})))
	statusView := furex.NewView(furex.MarginLeft(5), furex.MarginTop(10), furex.Direction(furex.Column))
	// add casting view
	castView := furex.NewView(furex.ID("cast"), furex.MarginTop(5), furex.Direction(furex.Column), furex.AlignItems(furex.AlignItemEnd))
	castView.AddChild(furex.NewView(furex.Width(210), furex.Height(12), furex.Handler(&Bar{
		Progress: func() float64 {
			cast := component.Sprite.Get(playerEntry).Instances[0].GetCast()
			if cast == nil {
				return 0
			}
			return float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast)
		},
		BG: castAtlas.GetNineSlice("casting_frame.png"),
		FG: castAtlas.GetNineSlice("casting_fg.png"),
	})))
	castView.AddChild(furex.NewView(furex.Height(12), furex.Width(100), furex.Handler(&Text{
		Align: furex.AlignItemStart,
		Content: func() string {
			cast := component.Sprite.Get(playerEntry).Instances[0].GetCast()
			if cast == nil {
				return ""
			}
			return cast.Name
		},
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  color.NRGBA{240, 152, 0, 128},
	})))

	statusView.AddChild(castView)
	// add name
	statusView.AddChild(furex.NewView(furex.ID("name"), furex.MarginTop(-12), furex.Width(100), furex.Height(13), furex.Handler(&Text{
		Align:        furex.AlignItemStart,
		Content:      player.Name,
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{22, 45, 87, 128},
	})))

	// view for hp and mp
	hm := furex.NewView(furex.Direction(furex.Row), furex.Justify(furex.JustifySpaceBetween), furex.Width(210))
	hm.AddChild(furex.NewView(furex.Direction(furex.Column), furex.AlignItems(furex.AlignItemEnd)).AddChild(furex.NewView(furex.MarginTop(3), furex.Width(125), furex.Height(8), furex.Handler(&Bar{
		Progress: func() float64 {
			return float64(player.HP) / float64(player.MaxHP)
		},
		FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
		BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
	}))).AddChild(furex.NewView(furex.Height(14), furex.MarginTop(-3), furex.Handler(&Text{
		Align: furex.AlignItemEnd,
		Content: func() string {
			return strconv.Itoa(player.HP)
		},
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{22, 45, 87, 128},
	}))))

	hm.AddChild(furex.NewView(furex.Direction(furex.Column), furex.AlignItems(furex.AlignItemEnd)).AddChild(furex.NewView(furex.MarginTop(3), furex.Width(75), furex.Height(8), furex.Handler(&Bar{
		Progress: func() float64 {
			return float64(player.Mana) / float64(player.MaxMana)
		},
		FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
		BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
	}))).AddChild(furex.NewView(furex.Height(14), furex.MarginTop(-3), furex.Handler(&Text{
		Align: furex.AlignItemEnd,
		Content: func() string {
			return strconv.Itoa(player.Mana)
		},
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{22, 45, 87, 128},
	}))))

	statusView.AddChild(hm)
	view.AddChild(statusView)
	bufflist := BuffListView(player.BuffList)
	bufflist.SetMarginTop(20)
	bufflist.SetMarginLeft(5)
	view.AddChild(bufflist)
	return view
}
