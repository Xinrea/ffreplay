package system

import (
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

// Remember that system is updated in TPS (Ticks Per Second) rate, in ebiten, it's 60 TPS.
type System struct {
	lock              sync.Mutex
	ecs               *ecs.ECS
	ViewPort          f64.Vec2
	PlayerList        []*donburi.Entry
	MapChangeEvents   []fflogs.FFLogsEvent
	LimitbreakEvents  []fflogs.FFLogsEvent
	WorldMarkerEvents EventLine
	EventLines        map[int64]*EventLine
	EntryMap          map[int64]*donburi.Entry
	reset             bool
	Pause             bool
}

type EventLine struct {
	Cursor int
	Events []fflogs.FFLogsEvent
	Status map[int][]data.StatusEvent
}

// NewSystem create a system that controls camera, players and all game objects also status.
// if replay is true, all skill and buff status are disabled, player status only change by fflogs event.
func NewSystem() *System {
	return &System{
		lock:       sync.Mutex{},
		EntryMap:   make(map[int64]*donburi.Entry),
		EventLines: make(map[int64]*EventLine),
		reset:      false,
	}
}

func (s *System) Init(ecs *ecs.ECS) {
	s.ecs = ecs
	ecs.AddSystem(s.Update)
}

func (s *System) Reset() {
	s.reset = true
}

func (s *System) Layout(w, h int) {
	s.ViewPort = f64.Vec2{float64(w), float64(h)}
}

func (s *System) AddEntry(id int64, player *donburi.Entry) {
	s.EntryMap[id] = player

	prole := component.Status.Get(player).Role
	if prole != role.Boss && prole != role.NPC {
		s.PlayerList = append(s.PlayerList, player)
	}
}

func (s *System) AddLimitbreakEvents(events []fflogs.FFLogsEvent) {
	s.LimitbreakEvents = events
}

func (s *System) AddMapChangeEvents(events []fflogs.FFLogsEvent) {
	s.MapChangeEvents = events
}

func (s *System) AddWorldMarkerEvents(events []fflogs.FFLogsEvent) {
	s.WorldMarkerEvents.Cursor = 0
	s.WorldMarkerEvents.Events = events
}

func (s *System) AddEventLine(id int64, status map[int][]data.StatusEvent, events []fflogs.FFLogsEvent) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.EventLines[id] = &EventLine{
		Cursor: 0,
		Events: events,
		Status: status,
	}

	// preprocess BTick and ETick
	sprite := component.Sprite.Get(s.EntryMap[id])
	if sprite == nil {
		return
	}

	for i := range sprite.Instances {
		if len(status[i]) == 0 {
			continue
		}

		first_event := status[i][0]
		last_event := status[i][len(status[i])-1]
		sprite.Instances[i].BTick = first_event.Tick
		sprite.Instances[i].ETick = last_event.Tick
	}

	for _, event := range events {
		instanceID := 0
		if event.SourceInstance != nil {
			instanceID = int(*event.SourceInstance) - 1
		}

		if sprite.Instances[instanceID].BTick == -1 {
			sprite.Instances[instanceID].BTick = event.LocalTick
		} else {
			sprite.Instances[instanceID].BTick = min(sprite.Instances[instanceID].BTick, event.LocalTick)
		}

		if sprite.Instances[instanceID].ETick == -1 {
			sprite.Instances[instanceID].ETick = event.LocalTick
		} else {
			sprite.Instances[instanceID].ETick = max(sprite.Instances[instanceID].ETick, event.LocalTick)
		}
	}
}

func (s *System) Update(ecs *ecs.ECS) {
	globalData := component.Global.Get(tag.Global.MustFirst(ecs.World))
	if globalData.Reset.Load() {
		s.doReset(ecs)
		globalData.Reset.Store(false)
	}

	if s.reset {
		s.doReset(ecs)

		globalData.Tick = 0
		s.reset = false
	}

	if globalData.Loaded.Load() {
		s.LogUpdate(ecs, globalData.Tick/10)
	}

	s.ControlUpdate(ecs)
	s.TimelineUpdate(ecs)
	s.BuffUpdate(ecs, globalData.Tick/10)
	s.SkillUpdate(ecs)
	s.WorldMarkerUpdate(ecs)
	s.BackgroundUpdate()

	if globalData.Loaded.Load() && !s.Pause && util.MSToTick(globalData.FightDuration.Load())*10 > globalData.Tick {
		globalData.Tick += globalData.Speed
		globalData.Tick = min(globalData.Tick, util.MSToTick(globalData.FightDuration.Load())*10)
	}
}

func (s *System) doReset(ecs *ecs.ECS) {
	// reset event lines
	s.WorldMarkerEvents.Cursor = 0
	for _, line := range s.EventLines {
		line.Cursor = 0
	}
	// clean all buffs and casting
	for e := range component.Status.Iter(ecs.World) {
		component.Status.Get(e).Reset()

		for _, instance := range component.Sprite.Get(e).Instances {
			instance.Reset()
		}
	}
}
