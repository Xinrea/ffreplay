package ui

import (
	"fmt"
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

type PlaygroundUI struct {
	view    *furex.View
	once    sync.Once
	maplist map[string]*bool
}

var _ UI = (*PlaygroundUI)(nil)

func NewPlaygroundUI(ecs *ecs.ECS) *PlaygroundUI {
	ecsInstance = ecs
	view := &furex.View{
		Position:  furex.PositionAbsolute,
		Direction: furex.Column,
		Top:       0,
		Left:      0,
	}
	return &PlaygroundUI{
		view:    view,
		maplist: make(map[string]*bool),
	}
}

func (p *PlaygroundUI) Update(w, h int) {
	global := entry.GetGlobal(ecsInstance)
	if global.Loaded.Load() {
		p.once.Do(func() {
			gamemap := component.Map.Get(component.Map.MustFirst(ecsInstance.World))
			camera := component.Camera.Get(component.Camera.MustFirst(ecsInstance.World))
			for _, m := range model.MapCache {
				if len(m.Phases) == 0 {
					v := false
					p.maplist[m.Path] = &v
					p.view.AddChild(CheckBoxView(16, false, &v, m.Path, func(checked bool) {
						if checked {
							gamemap.Config = m.Load()
							camera.Position = vector.NewVector(m.Offset.X*25, m.Offset.Y*25)
							for k, v := range p.maplist {
								if k != m.Path {
									*v = false
								}
							}
						} else {
							v = true
						}
					}))
					continue
				}
				for i, phase := range m.Phases {
					id := fmt.Sprintf("%s:%d", phase.Path, i)
					v := false
					p.maplist[id] = &v
					p.view.AddChild(CheckBoxView(16, false, &v, id, func(checked bool) {
						if checked {
							gamemap.Config = m.Load()
							gamemap.Config.CurrentPhase = i
							camera.Position = vector.NewVector(m.Offset.X*25, m.Offset.Y*25)
							for k, v := range p.maplist {
								if k != id {
									*v = false
								}
							}
						} else {
							v = true
						}
					}))
				}
			}
		})
	}
	s := ebiten.Monitor().DeviceScaleFactor()
	furex.GlobalScale = s
	if p.view != nil {
		p.view.UpdateWithSize(w, h)
	}
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {
	if p.view != nil {
		p.view.Draw(screen)
	}
}
