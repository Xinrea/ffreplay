package system

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

// Remember that system is updated in TPS (Ticks Per Second) rate, in ebiten, it's 60 TPS.
type System struct {
	lock           sync.Mutex
	ECS            *ecs.ECS
	ViewPort       f64.Vec2
	MainPlayerRole model.RoleType
	EventLines     map[int64]*EventLine
	EntryMap       map[int64]*donburi.Entry
	InReplay       bool
	reset          bool
	Pause          bool
}

type EventLine struct {
	Cursor int
	Events []fflogs.FFLogsEvent
	Status map[int][]StatusEvent
}

type StatusEvent struct {
	Tick     int64
	Face     float64
	HP       int
	MaxHP    int
	MP       int
	MaxMP    int
	Position vector.Vector
}

// NewSystem create a system that controls camera, players and all game objects also status.
// if replay is true, all skill and buff status are disabled, player status only change by fflogs event
func NewSystem(replay bool) *System {
	return &System{
		lock:           sync.Mutex{},
		MainPlayerRole: model.MT,
		InReplay:       replay,
		EntryMap:       make(map[int64]*donburi.Entry),
		EventLines:     make(map[int64]*EventLine),
		reset:          false,
	}
}

func (s *System) Reset() {
	s.reset = true
}

func (s *System) Layout(w, h int) {
	s.ViewPort = f64.Vec2{float64(w), float64(h)}
}

func (s *System) AddEntry(id int64, player *donburi.Entry) {
	s.EntryMap[id] = player
	role := component.Status.Get(player).Role
	if role != model.Boss && role != model.NPC {
		s.MainPlayerRole = role
	}
}

func (s *System) AddEventLine(id int64, events []fflogs.FFLogsEvent) {
	s.lock.Lock()
	defer s.lock.Unlock()
	status := make(map[int][]StatusEvent)
	for _, e := range events {
		// make sure every tick has at most one status event
		if e.SourceID != nil && *e.SourceID == id && e.SourceResources != nil {
			instanceID := 1
			if e.SourceInstance != nil {
				instanceID = int(*e.SourceInstance)
			}
			if len(status[instanceID]) == 0 || status[instanceID][len(status[instanceID])-1].Tick != e.LocalTick {
				status[instanceID] = append(status[instanceID], StatusEvent{
					Tick:     e.LocalTick,
					Face:     float64(e.SourceResources.Facing) / 100,
					HP:       int(e.SourceResources.HitPoints),
					MaxHP:    int(e.SourceResources.MaxHitPoints),
					MP:       int(e.SourceResources.Mp),
					MaxMP:    int(e.SourceResources.MaxMP),
					Position: vector.Vector{float64(e.SourceResources.X-10000) / 100 * 25, float64(e.SourceResources.Y-10000) / 100 * 25},
				})
			} else {
				status[instanceID][len(status[instanceID])-1] = StatusEvent{
					Tick:     e.LocalTick,
					Face:     float64(e.SourceResources.Facing) / 100,
					HP:       int(e.SourceResources.HitPoints),
					MaxHP:    int(e.SourceResources.MaxHitPoints),
					MP:       int(e.SourceResources.Mp),
					MaxMP:    int(e.SourceResources.MaxMP),
					Position: vector.Vector{float64(e.SourceResources.X-10000) / 100 * 25, float64(e.SourceResources.Y-10000) / 100 * 25},
				}
			}
		}
		if e.TargetID != nil && *e.TargetID == id && e.TargetResources != nil {
			instanceID := 1
			if e.TargetInstance != nil {
				instanceID = int(*e.TargetInstance)
			}
			if len(status[instanceID]) == 0 || status[instanceID][len(status[instanceID])-1].Tick != e.LocalTick {
				status[instanceID] = append(status[instanceID], StatusEvent{
					Tick:     e.LocalTick,
					Face:     float64(e.TargetResources.Facing) / 100,
					HP:       int(e.TargetResources.HitPoints),
					MaxHP:    int(e.TargetResources.MaxHitPoints),
					MP:       int(e.TargetResources.Mp),
					MaxMP:    int(e.TargetResources.MaxMP),
					Position: vector.Vector{float64(e.TargetResources.X-10000) / 100 * 25, float64(e.TargetResources.Y-10000) / 100 * 25},
				})
			} else {
				status[instanceID][len(status[instanceID])-1] = StatusEvent{
					Tick:     e.LocalTick,
					Face:     float64(e.TargetResources.Facing) / 100,
					HP:       int(e.TargetResources.HitPoints),
					MaxHP:    int(e.TargetResources.MaxHitPoints),
					MP:       int(e.TargetResources.Mp),
					MaxMP:    int(e.TargetResources.MaxMP),
					Position: vector.Vector{float64(e.TargetResources.X-10000) / 100 * 25, float64(e.TargetResources.Y-10000) / 100 * 25},
				}
			}
		}
	}
	filteredEvents := make([]fflogs.FFLogsEvent, 0)
	for _, e := range events {
		if e.SourceID != nil && *e.SourceID == id {
			filteredEvents = append(filteredEvents, e)
		}
	}
	s.EventLines[id] = &EventLine{
		Cursor: 0,
		Events: filteredEvents,
		Status: status,
	}
}

func (s *System) Update(ecs *ecs.ECS) {
	globalData := component.Global.Get(tag.Global.MustFirst(ecs.World))
	if s.reset {
		s.doReset(ecs)
		globalData.Tick = 0
		s.reset = false
	}
	s.LogUpdate(ecs, globalData.Tick/10)
	s.ControlUpdate(ecs)
	s.TimelineUpdate(ecs)
	s.BuffUpdate(ecs, globalData.Tick/10)
	s.SkillUpdate(ecs)
	s.MarkerUpdate(ecs)
	if globalData.Loaded.Load() && !s.Pause && util.MSToTick(globalData.FightDuration.Load())*10 > globalData.Tick {
		globalData.Tick += globalData.Speed
		globalData.Tick = min(globalData.Tick, util.MSToTick(globalData.FightDuration.Load())*10)
	}
}

func (s *System) doReset(ecs *ecs.ECS) {
	// reset event lines
	for _, line := range s.EventLines {
		line.Cursor = 0
	}
	// clean all buffs and casting
	for e := range component.Status.Iter(ecs.World) {
		status := component.Status.Get(e)
		status.BuffList.Clear()
		for _, instance := range component.Sprite.Get(e).Instances {
			instance.Casting = nil
		}
	}
}
