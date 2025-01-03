package ui

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

var ecsInstance *ecs.ECS
var barAtlas = texture.NewTextureAtlasFromFile("asset/ui/bar.xml")
var castAtlas = texture.NewTextureAtlasFromFile("asset/ui/casting.xml")

type FFUI struct {
	view *furex.View
	once sync.Once
}

func NewFFUI(ecs *ecs.ECS) *FFUI {
	ecsInstance = ecs
	view := &furex.View{
		Direction: furex.Row,
		Wrap:      furex.Wrap,
	}
	view.AddChild(&furex.View{
		ID:           "left",
		Direction:    furex.Column,
		AlignItems:   furex.AlignItemStart,
		AlignContent: furex.AlignContentSpaceBetween,
		MarginTop:    20,
		MarginLeft:   20,
		MarginBottom: 20,
		Grow:         0.5,
	})
	return &FFUI{
		view: view,
	}
}

func (f *FFUI) Update(w, h int) {
	global := entry.GetGlobal(ecsInstance)
	if !global.Loaded.Load() {
		return
	}
	f.once.Do(func() {
		lview := f.view.MustGetByID("left")
		tlview := &furex.View{
			Direction:  furex.Column,
			AlignItems: furex.AlignItemStart,
		}
		tlview.AddChild(&furex.View{
			Handler: &LimitBreak{
				Value:     &global.LimitBreak,
				BarNumber: &global.Bar,
			},
		})
		// TODO Considering party member changes (remove/add)
		memberList := []*donburi.Entry{}
		tag.PartyMember.Each(ecsInstance.World, func(e *donburi.Entry) {
			status := component.Status.Get(e)
			if status.Role == model.Pet {
				return
			}
			memberList = append(memberList, e)
		})
		tlview.AddChild(NewPartyList(memberList))
		lview.AddChild(tlview)

		enemyBarView := EnemyBarsView()
		f.view.AddChild(enemyBarView)

		playProgressView := ProgressBarView()
		f.view.AddChild(playProgressView)
	})
	furex.GlobalScale = ebiten.Monitor().DeviceScaleFactor()
	f.view.UpdateWithSize(w, h)
}

func (f *FFUI) Draw(screen *ebiten.Image) {
	if f.view != nil {
		f.view.Draw(screen)
	}
}
