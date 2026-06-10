package ui

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
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
	view    *furex.View
	eui     *ebitenui.UI
	euiRoot *widget.Container
	once    sync.Once
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
	euiRoot := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	return &FFUI{
		view:    view,
		euiRoot: euiRoot,
		eui:     &ebitenui.UI{Container: euiRoot},
	}
}

func (f *FFUI) Update(w, h int) {
	global := entry.GetGlobal(ecsInstance)
	if !global.Loaded.Load() {
		return
	}

	scale := ebiten.Monitor().DeviceScaleFactor()
	if scale <= 0 {
		scale = 1
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

		f.view.AddChild(rview)
		f.euiRoot.AddChild(EUIProgressBarView(scale))
	})

	furex.GlobalScale = scale

	f.view.UpdateWithSize(w, h)
	f.eui.Update()
}

func (f *FFUI) Draw(screen *ebiten.Image) {
	if f.view != nil {
		f.view.Draw(screen)
	}
	if f.eui != nil {
		f.eui.Draw(screen)
	}
}
