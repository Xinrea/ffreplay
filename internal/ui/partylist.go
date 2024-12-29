package ui

import (
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

func NewPartyList(players []*donburi.Entry) *furex.View {
	view := &furex.View{
		Direction:  furex.Column,
		AlignItems: furex.AlignItemStart,
		Justify:    furex.JustifyStart,
	}
	view.AddChild(&furex.View{
		Width:    300,
		Height:   50*len(players) + 20,
		Position: furex.PositionAbsolute,
		Handler:  &Sprite{Texture: texture.NewNineSlice(texture.NewTextureFromFile("asset/partylist_bg.png"), 5, 14, 0, 0)},
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
}

func (p *PlayerItem) HandleJustPressedMouseButtonLeft(x, y int) bool {
	entry.GetGlobal(ecsInstance).TargetPlayer = p.Player
	return true
}

func (p *PlayerItem) HandleJustReleasedMouseButtonLeft(x, y int) {
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
		Handler:  &Sprite{Texture: texture.NewNineSlice(texture.NewTextureFromFile("asset/partylist_hover.png"), 0, 0, 0, 0)},
	})
	view.AddChild(&furex.View{
		ID:       "selected",
		Position: furex.PositionAbsolute,
		Top:      10,
		Left:     30,
		Width:    275,
		Height:   40,
		Handler:  &Sprite{Texture: texture.NewNineSlice(texture.NewTextureFromFile("asset/partylist_selected.png"), 0, 0, 0, 0)},
	})
	// add job icon
	view.AddChild(&furex.View{
		Width:   38,
		Height:  38,
		Handler: &Sprite{Texture: texture.NewNineSlice(player.RoleTexture(), 0, 0, 0, 0)},
	})
	statusView := &furex.View{
		MarginLeft: 5,
		MarginTop:  20,
		Direction:  furex.Column,
	}
	// add name
	statusView.AddChild(&furex.View{
		Width:  100,
		Height: 13,
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
	view.AddChild(&furex.View{
		MarginTop:  20,
		MarginLeft: 5,
		Handler: &BuffList{
			Buffs: player.BuffList,
		},
	})
	return view
}
