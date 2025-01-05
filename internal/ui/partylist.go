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
	view := &furex.View{
		TagName:    "PartyList",
		MarginTop:  10,
		Direction:  furex.Column,
		AlignItems: furex.AlignItemStart,
		Justify:    furex.JustifyStart,
	}
	view.AddChild(&furex.View{
		TagName:  "PartyListBG",
		Width:    300,
		Height:   48*len(players) + 5,
		Position: furex.PositionAbsolute,
		Handler:  &Sprite{NineSliceTexture: texture.NewNineSlice(texture.NewTextureFromFile("asset/partylist_bg.png"), 5, 14, 0, 0)},
	})
	for _, p := range players {
		view.AddChild(NewPlayerItem(p))
	}
	return view
}

type PlayerItem struct {
	Player   *donburi.Entry
	Hovered  bool
	Selected bool
}

var _ furex.MouseEnterLeaveHandler = (*PlayerItem)(nil)
var _ furex.MouseLeftButtonHandler = (*PlayerItem)(nil)

func (p *PlayerItem) Update(v *furex.View) {
	// status := component.Status.Get(p.Player)
	if p.Hovered {
		v.MustGetByID("hover").Display = furex.DisplayFlex
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		v.MustGetByID("hover").Display = furex.DisplayNone
	}
	targetPlayer := entry.GetGlobal(ecsInstance).TargetPlayer
	if targetPlayer == p.Player {
		v.MustGetByID("selected").Display = furex.DisplayFlex
	} else {
		v.MustGetByID("selected").Display = furex.DisplayNone
	}

	// if player is casting, hide name
	if component.Sprite.Get(p.Player).Instances[0].GetCast() != nil {
		v.MustGetByID("name").Display = furex.DisplayNone
		v.MustGetByID("cast").Display = furex.DisplayFlex
	} else {
		v.MustGetByID("name").Display = furex.DisplayFlex
		v.MustGetByID("cast").Display = furex.DisplayNone
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
	view := &furex.View{
		ID:         strconv.Itoa(int(player.ID)),
		Height:     48,
		Direction:  furex.Row,
		AlignItems: furex.AlignItemCenter,
		Justify:    furex.JustifyStart,
		Handler:    item,
	}
	view.AddChild(&furex.View{
		ID:       "hover",
		Position: furex.PositionAbsolute,
		Top:      10,
		Left:     30,
		Width:    275,
		Height:   40,
		Handler:  &Sprite{Texture: texture.NewTextureFromFile("asset/partylist_hover.png")},
	})
	view.AddChild(&furex.View{
		ID:       "selected",
		Position: furex.PositionAbsolute,
		Top:      10,
		Left:     30,
		Width:    275,
		Height:   40,
		Handler:  &Sprite{Texture: texture.NewTextureFromFile("asset/partylist_selected.png")},
	})
	// add job icon
	view.AddChild(&furex.View{
		Width:   38,
		Height:  38,
		Handler: &Sprite{Texture: player.RoleTexture()},
	})
	statusView := &furex.View{
		MarginLeft: 5,
		Direction:  furex.Column,
	}
	// add casting view
	castView := &furex.View{
		ID:         "cast",
		MarginTop:  5,
		Direction:  furex.Column,
		AlignItems: furex.AlignItemEnd,
	}
	castView.AddChild(&furex.View{
		Width:  210,
		Height: 12,
		Handler: &Bar{
			Progress: func() float64 {
				cast := component.Sprite.Get(playerEntry).Instances[0].GetCast()
				if cast == nil {
					return 0
				}
				return float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast)
			},
			BG: castAtlas.GetNineSlice("casting_frame.png"),
			FG: castAtlas.GetNineSlice("casting_fg.png"),
		},
	})
	castView.AddChild(&furex.View{
		MarginTop: -5,
		Width:     100,
		Height:    12,
		Handler: &Text{
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
		}})
	statusView.AddChild(castView)
	// add name
	statusView.AddChild(&furex.View{
		ID:        "name",
		MarginTop: 10,
		Width:     100,
		Height:    13,
		Handler: &Text{
			Align:        furex.AlignItemStart,
			Content:      player.Name,
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}})
	// view for hp and mp
	hm := &furex.View{
		Direction: furex.Row,
		Justify:   furex.JustifySpaceBetween,
		Width:     210,
	}
	hm.AddChild((&furex.View{
		Direction:  furex.Column,
		AlignItems: furex.AlignItemEnd,
	}).AddChild(&furex.View{
		MarginTop: 3,
		Width:     125,
		Height:    8,
		Handler: &Bar{
			Progress: func() float64 {
				return float64(player.HP) / float64(player.MaxHP)
			},
			FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
			BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
		}}).AddChild(&furex.View{
		Height:    14,
		MarginTop: -3,
		Handler: &Text{
			Align: furex.AlignItemEnd,
			Content: func() string {
				return strconv.Itoa(player.HP)
			},
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	}))
	hm.AddChild((&furex.View{
		Direction:  furex.Column,
		AlignItems: furex.AlignItemEnd,
	}).AddChild(&furex.View{
		MarginTop: 3,
		Width:     75,
		Height:    8,
		Handler: &Bar{
			Progress: func() float64 {
				return float64(player.Mana) / float64(player.MaxMana)
			},
			FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
			BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
		}}).AddChild(&furex.View{
		Height:    14,
		MarginTop: -3,
		Handler: &Text{
			Align: furex.AlignItemEnd,
			Content: func() string {
				return strconv.Itoa(player.Mana)
			},
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}}))
	statusView.AddChild(hm)
	view.AddChild(statusView)
	bufflist := BuffListView(player.BuffList)
	bufflist.SetMarginTop(20)
	bufflist.SetMarginLeft(5)
	view.AddChild(bufflist)
	return view
}
