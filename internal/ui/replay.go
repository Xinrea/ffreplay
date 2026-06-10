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
	euiRoot := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	return &FFUI{
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
		memberList := []*donburi.Entry{}
		tag.PartyMember.Each(ecsInstance.World, func(e *donburi.Entry) {
			status := component.Status.Get(e)
			if status.Role == role.Pet {
				return
			}

			memberList = append(memberList, e)
		})

		f.euiRoot.AddChild(NewEUIReplayPartyList(memberList, scale))
		f.euiRoot.AddChild(NewEUIDamageHistoryView(len(memberList), scale))
		f.euiRoot.AddChild(EUIEnemyBarsView(scale))
		f.euiRoot.AddChild(NewEUILimitBreak(&global.LimitBreak, &global.Bar, scale))
		f.euiRoot.AddChild(EUIProgressBarView(scale))
	})

	f.eui.Update()
}

func (f *FFUI) Draw(screen *ebiten.Image) {
	if f.eui != nil {
		f.eui.Draw(screen)
	}
}
