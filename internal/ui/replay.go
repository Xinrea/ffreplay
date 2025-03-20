package ui

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

var (
	ecsInstance *ecs.ECS
	barAtlas    = texture.NewTextureAtlasFromFile("asset/ui/bar.xml")
	castAtlas   = texture.NewTextureAtlasFromFile("asset/ui/casting.xml")
)

const (
	UIHalf    = 0.5
	UIPadding = 20
)

type FFUI struct {
	view *furex.View
	once sync.Once
}

type UI interface {
	Update(w int, h int)
	Draw(screen *ebiten.Image)
}

var _ UI = (*FFUI)(nil)

func NewReplayUI(ecs *ecs.ECS) *FFUI {
	ecsInstance = ecs
	view := furex.NewView(
		furex.Direction(furex.Row),
		furex.Justify(furex.JustifySpaceBetween),
	)

	return &FFUI{
		view: view,
	}
}

func (f *FFUI) Update(w, h int) {
	global := entry.GetGlobal()
	if !global.Loaded.Load() {
		return
	}

	f.once.Do(func() {
		lview := furex.NewView(
			furex.ID("left"),
			furex.MarginTop(UIPadding),
			furex.MarginLeft(UIPadding),
			furex.Direction(furex.Column),
		)

		lview.AddChild(furex.NewView(furex.Handler(&LimitBreak{
			Value:     &global.LimitBreak,
			BarNumber: &global.Bar,
		})))

		memberList := []*donburi.Entry{}

		tag.PartyMember.Each(ecsInstance.World, func(e *donburi.Entry) {
			status := component.Status.Get(e)
			if status.Role == role.Pet {
				return
			}

			memberList = append(memberList, e)
		})

		lview.AddChild(NewPartyList(memberList))
		lview.AddChild(DamageHistoryView())

		f.view.AddChild(lview)

		rview := furex.NewView(
			furex.ID("right"),
			furex.Width(600),
			furex.MarginTop(UIPadding),
			furex.MarginRight(UIPadding),
			furex.MarginBottom(UIPadding),
			furex.Direction(furex.Column),
			furex.Justify(furex.JustifySpaceBetween),
			furex.AlignItems(furex.AlignItemEnd),
		)
		rview.AddChild(EnemyBarsView())
		rview.AddChild(ProgressBarView())

		f.view.AddChild(rview)
	})

	furex.GlobalScale = ebiten.Monitor().DeviceScaleFactor()

	f.view.UpdateWithSize(w, h)
}

func (f *FFUI) Draw(screen *ebiten.Image) {
	if f.view != nil {
		f.view.Draw(screen)
	}
}
