package ui

import (
	"image"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

const (
	PartyListWidth   = 300
	PartyListBGExtra = 20
)

var PartyListBGNineSliceConfig = [4]int{5, 14, 0, 0}

func NewPartyList(players []*donburi.Entry) *furex.View {
	view := furex.NewView(
		furex.TagName("PartyList"),
		furex.MarginTop(10),
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemStart),
		furex.Justify(furex.JustifyStart),
		furex.Handler(furex.ViewHandler{
			Update: func(v *furex.View) {
				playerItemLen := v.Len() - 1
				bg := v.First()
				expectHeight := playerItemLen*PlayerItemHeight + PartyListBGExtra
				// TODO view.SetHeight should be able to handle this
				if bg.Attrs.Height != expectHeight {
					bg.SetHeight(expectHeight)
				}
			},
		}),
	)

	view.AddChild(
		furex.NewView(
			furex.TagName("PartyListBG"),
			furex.Width(PartyListWidth),
			furex.Position(furex.PositionAbsolute),
			furex.Handler(&Sprite{
				NineSliceTexture: texture.NewNineSlice(
					texture.NewTextureFromFile("asset/partylist_bg.png"),
					PartyListBGNineSliceConfig[0],
					PartyListBGNineSliceConfig[1],
					PartyListBGNineSliceConfig[2],
					PartyListBGNineSliceConfig[3]),
			}),
		),
	)

	for _, p := range players {
		view.AddChild(NewPlayerItem(p))
	}

	view.Layout()

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
	if component.Status.Get(p.Player).Instances[0].GetCast() != nil {
		v.MustGetByID("name").Attrs.Hidden = true
		v.MustGetByID("cast").Attrs.Hidden = false
	} else {
		v.MustGetByID("name").Attrs.Hidden = false
		v.MustGetByID("cast").Attrs.Hidden = true
	}
}

func (p *PlayerItem) HandleJustPressedMouseButtonLeft(_ image.Rectangle, x, y int) bool {
	entry.GetGlobal(ecsInstance).TargetPlayer = p.Player

	return false
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

	CastBarWidth = 100
	BarHeight    = 12

	HPBarWidth   = 125
	MPBarWidth   = 75
	HMPBarHeight = 8

	BuffListOffsetY = 20
)

func NewPlayerItem(playerEntry *donburi.Entry) *furex.View {
	item := &PlayerItem{
		Player: playerEntry,
	}
	player := component.Status.Get(playerEntry)
	view := furex.NewView(
		furex.ID(strconv.Itoa(int(player.ID))),
		furex.Height(PlayerItemHeight), furex.Direction(furex.Row),
		furex.AlignItems(furex.AlignItemCenter),
		furex.Justify(furex.JustifyStart),
		furex.Handler(item),
	)

	// add hover/select sprite
	addHoverSprite(view)

	// add job icon
	view.AddChild(
		furex.NewView(
			furex.Width(JobIconSize),
			furex.Height(JobIconSize),
			furex.Handler(&Sprite{Texture: player.RoleTexture()})))

	// statusView contains name, hp, mp, cast
	statusView := furex.NewView(furex.MarginLeft(5), furex.MarginTop(10), furex.Direction(furex.Column))
	// add casting view
	statusView.AddChild(createCastingView(playerEntry))
	// add name
	statusView.AddChild(
		furex.NewView(
			furex.ID("name"),
			furex.MarginTop(-12),
			furex.Height(NameTextSize),
			furex.Handler(&Text{
				Align:        furex.AlignItemStart,
				Content:      player.Name,
				Color:        color.White,
				Shadow:       true,
				ShadowOffset: 2,
				ShadowColor:  color.NRGBA{22, 45, 87, 128},
			})))

	// view for hp and mp
	statusView.AddChild(createHPMPBar(player))

	view.AddChild(statusView)

	bufflist := BuffListView(player.BuffList)
	bufflist.SetMarginTop(BuffListOffsetY)
	bufflist.SetMarginLeft(5)

	view.AddChild(bufflist)

	return view
}

func addHoverSprite(view *furex.View) {
	view.AddChild(
		furex.NewView(
			furex.ID("hover"),
			furex.Position(furex.PositionAbsolute),
			furex.Top(HoverSpriteTop),
			furex.Left(HoverSpriteLeft),
			furex.Width(HoverSpriteWidth),
			furex.Height(HoverSpriteHeight),
			furex.Handler(&Sprite{
				Texture: texture.NewTextureFromFile("asset/partylist_hover.png"),
			})))
	view.AddChild(
		furex.NewView(
			furex.ID("selected"),
			furex.Position(furex.PositionAbsolute),
			furex.Top(HoverSpriteTop),
			furex.Left(HoverSpriteLeft),
			furex.Width(HoverSpriteWidth),
			furex.Height(HoverSpriteHeight),
			furex.Handler(&Sprite{Texture: texture.NewTextureFromFile("asset/partylist_selected.png")})))
}

func createCastingView(e *donburi.Entry) *furex.View {
	castView := furex.NewView(
		furex.ID("cast"),
		furex.MarginTop(5),
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemEnd))
	castView.AddChild(
		furex.NewView(
			furex.Width(StatusPartWidth),
			furex.Height(BarHeight),
			furex.Handler(&Bar{
				Progress: func() float64 {
					cast := component.Status.Get(e).Instances[0].GetCast()
					if cast == nil {
						return 0
					}

					return float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast)
				},
				BG: castAtlas.GetNineSlice("casting_frame.png"),
				FG: castAtlas.GetNineSlice("casting_fg.png"),
			})))
	castView.AddChild(furex.NewView(furex.Height(CastNameTextSize), furex.Handler(&Text{
		Align: furex.AlignItemStart,
		Content: func() string {
			cast := component.Status.Get(e).Instances[0].GetCast()
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

	return castView
}

func createHPMPBar(player *model.StatusData) *furex.View {
	hm := furex.NewView(
		furex.Direction(furex.Row),
		furex.Justify(furex.JustifySpaceBetween),
		furex.Width(StatusPartWidth),
	)

	createBarView := func(bar *Bar, text *Text, w, h int) *furex.View {
		const MarginTop = 3

		return furex.NewView(
			furex.Direction(furex.Column),
			furex.AlignItems(furex.AlignItemEnd),
		).AddChild(
			furex.NewView(
				furex.MarginTop(MarginTop),
				furex.Width(w),
				furex.Height(h),
				furex.Handler(bar))).AddChild(
			furex.NewView(
				furex.Height(HMPTextSize),
				furex.MarginTop(-3),
				furex.Handler(text)))
	}

	createBarView(&Bar{
		Progress: func() float64 {
			return float64(player.HP) / float64(player.MaxHP)
		},
		FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
		BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
	}, &Text{
		Align: furex.AlignItemEnd,
		Content: func() string {
			return strconv.Itoa(player.HP)
		},
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{22, 45, 87, 128},
	}, HPBarWidth, HMPBarHeight).AddTo(hm)

	createBarView(&Bar{
		Progress: func() float64 {
			return float64(player.Mana) / float64(player.MaxMana)
		},
		FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
		BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
	}, &Text{
		Align: furex.AlignItemEnd,
		Content: func() string {
			return strconv.Itoa(player.Mana)
		},
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{22, 45, 87, 128},
	}, MPBarWidth, HMPBarHeight).AddTo(hm)

	return hm
}
