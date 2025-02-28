package system

import (
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/game/buffs"
	"github.com/Xinrea/ffreplay/internal/game/skills"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/fogleman/ease"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

// Update only do update-work every 30 ticks, which means 0.5 second in default 60 TPS.
func (s *System) LogUpdate(ecs *ecs.ECS, tick int64) {
	if entry.GetGlobal(s.ecs).ReplayMode {
		s.replayUpdate(ecs, tick)
	}
}

func (s *System) replayUpdate(ecs *ecs.ECS, tick int64) {
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	if !global.Loaded.Load() {
		return
	}

	// limitbreak event
	index := sort.Search(len(s.LimitbreakEvents), func(i int) bool {
		return s.LimitbreakEvents[i].LocalTick > tick
	})
	if index > 0 {
		global.Bar = int(*s.LimitbreakEvents[index-1].Bars)
		global.LimitBreak = int(*s.LimitbreakEvents[index-1].Value)
	}

	// map event
	gamemap := component.Map.Get(component.Map.MustFirst(ecs.World))
	if gamemap.Config != nil {
		index = sort.Search(len(s.MapChangeEvents), func(i int) bool {
			return s.MapChangeEvents[i].LocalTick > tick
		})
		if index > 0 {
			gamemap.Config.CurrentMap = *s.MapChangeEvents[index-1].MapID
		}
	}

	// marker event
	for s.WorldMarkerEvents.Cursor < len(s.WorldMarkerEvents.Events) &&
		s.WorldMarkerEvents.Events[s.WorldMarkerEvents.Cursor].LocalTick <= tick {
		event := s.WorldMarkerEvents.Events[s.WorldMarkerEvents.Cursor]
		s.applyLog(ecs, nil, event)

		s.WorldMarkerEvents.Cursor++
	}

	s.handleFFlogsEvents(ecs, tick)
}

func (s *System) handleFFlogsEvents(ecs *ecs.ECS, tick int64) {
	lineMap := make(map[*donburi.Entry]*EventLine)

	for e := range tag.GameObject.Iter(ecs.World) {
		id := component.Status.Get(e).ID

		line := s.EventLines[id]
		if line == nil || line.Cursor >= len(line.Events) {
			continue
		}

		lineMap[e] = line

		s.updateInstances(e, line, tick)
	}

	s.consumeEvents(tick, lineMap)
}

func (s *System) consumeEvents(tick int64, lineMap map[*donburi.Entry]*EventLine) {
	for {
		var topTarget *donburi.Entry = nil

		var topLine *EventLine = nil

		var topTick int64 = math.MaxInt64

		for e, line := range lineMap {
			if line.Cursor < len(line.Events) && line.Events[line.Cursor].LocalTick <= tick {
				if topLine == nil || line.Events[line.Cursor].LocalTick < topTick {
					topTarget = e
					topLine = line
					topTick = line.Events[line.Cursor].LocalTick
				}
			}
		}

		if topLine == nil {
			break
		}

		s.applyLog(s.ecs, topTarget, topLine.Events[topLine.Cursor])

		topLine.Cursor++
	}
}

func (s *System) updateInstances(e *donburi.Entry, line *EventLine, tick int64) {
	for i, sprite := range component.Status.Get(e).Instances {
		isNPC := component.Status.Get(e).Role == role.NPC
		instanceID := i + 1
		index := sort.Search(len(line.Status[instanceID]), func(i int) bool {
			return line.Status[instanceID][i].Tick >= tick
		})

		if index == len(line.Status[instanceID]) {
			continue
		}

		status := line.Status[instanceID][index]
		if index == 0 || component.Status.Get(e).IsDead() {
			s.normalUpdate(e, sprite, status)
		} else {
			previous := line.Status[instanceID][index-1]

			if isNPC {
				s.normalUpdate(e, sprite, status)
			} else {
				s.lerpUpdate(e, sprite, previous, status, tick)
			}
		}
	}
}

func (s *System) normalUpdate(e *donburi.Entry, sprite *model.Instance, status data.StatusEvent) {
	facing := status.Face + math.Pi/2
	sprite.Face = facing
	sprite.Object.UpdatePosition(status.Position)
	component.Status.Get(e).HP = status.HP
	component.Status.Get(e).MaxHP = status.MaxHP
	component.Status.Get(e).Mana = status.MP
	component.Status.Get(e).MaxMana = status.MaxMP
}

func (s *System) lerpUpdate(e *donburi.Entry, sprite *model.Instance, previous, status data.StatusEvent, tick int64) {
	t := float64(tick - previous.Tick)
	d := float64(status.Tick - previous.Tick)

	if d == 0 {
		d = t
	}

	pos := previous.Position.Lerp(status.Position, ease.InOutSine(t/d))
	facing := math.Pi/2 + util.LerpRadians(previous.Face, status.Face, ease.InOutSine(t/d))
	sprite.Face = facing
	sprite.Object.UpdatePosition(pos)

	component.Status.Get(e).HP = status.HP
	component.Status.Get(e).MaxHP = status.MaxHP
	component.Status.Get(e).Mana = status.MP
	component.Status.Get(e).MaxMana = status.MaxMP
}

type EventHandler func(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent)

var EventHandlerMap = map[fflogs.EventType]EventHandler{
	fflogs.Combatantinfo:      handleCombatantinfo,
	fflogs.Applybuff:          handleApplyBuff,
	fflogs.Applydebuff:        handleApplyDebuff,
	fflogs.Refreshbuff:        handleRefreshBuff,
	fflogs.RefreshDebuff:      handleRefreshDebuff,
	fflogs.Removebuff:         handleRemoveBuff,
	fflogs.RemoveDebuff:       handleRemoveDebuff,
	fflogs.Begincast:          handleBeginCast,
	fflogs.Cast:               handleCast,
	fflogs.Death:              handleDeath,
	fflogs.WorldMarkerRemoved: handleWorldMarkerRemoved,
	fflogs.WorldMarkerPlaced:  handleWorldMarkerPlaced,
	fflogs.Applybuffstack:     handleApplyBuffStack,
	fflogs.Removebuffstack:    handleRemoveBuffStack,
	fflogs.TDamage:            handleDamage,
	fflogs.Tether:             handleTether,
}

func (s *System) applyLog(ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	if event.SourceID != nil && s.EntryMap[*event.SourceID] != nil {
		s.updateEventSourceStatus(event)
	}

	if event.TargetID != nil && s.EntryMap[*event.TargetID] != nil {
		s.updateEventTargetStatus(event)
	}

	if handler, ok := EventHandlerMap[event.Type]; ok {
		handler(s, ecs, eventSource, event)

		return
	}
}

var TETHER_SUPPORTED_MAPS = map[int]bool{
	77: true, // FRU
}

func handleTether(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	currentMapID := component.Map.Get(component.Map.MustFirst(ecs.World)).Config.CurrentMap
	if _, ok := TETHER_SUPPORTED_MAPS[currentMapID]; !ok {
		return
	}

	source := s.EntryMap[*event.SourceID]
	target := s.EntryMap[*event.TargetID]

	if source == nil || target == nil {
		return
	}

	sourceStatus := component.Status.Get(source)
	targetStatus := component.Status.Get(target)

	sourceStatus.AddTether(event.LocalTick, targetStatus)

	return
}

func handleCombatantinfo(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	status := component.Status.Get(eventSource)
	status.BuffList.SetBuffs(aurasToBuffs(event.Auras))

	return
}

func handleRefreshBuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	ability := (*event.Ability).ToBuff()
	ability.ApplyTick = event.LocalTick
	ability.Duration = *event.Duration
	status.BuffList.Add(ability)

	return
}

func handleRefreshDebuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	ability := (*event.Ability).ToBuff()
	ability.ApplyTick = event.LocalTick
	ability.Duration = *event.Duration
	status.BuffList.Add(ability)

	return
}

func handleRemoveBuffStack(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	status.BuffList.UpdateStack(event.Ability.Guid, int(*event.Stack))

	return
}

func handleApplyBuffStack(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	status.BuffList.UpdateStack(event.Ability.Guid, int(*event.Stack))

	return
}

func handleRemoveBuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	ability := (*event.Ability).ToBuff()
	status.BuffList.Remove(ability)

	return
}

func handleRemoveDebuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)

	ability := (*event.Ability).ToBuff()
	status.BuffList.Remove(ability)

	if entry.GetGlobal(ecs).Debug {
		util.PrintJson(ability)
	}

	return
}

func handleDeath(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	if event.TargetID == nil {
		return
	}

	eventTarget := s.EntryMap[*event.TargetID]
	if eventTarget == nil {
		return
	}

	status := component.Status.Get(eventTarget)
	status.SetDeath(true)

	return
}

func handleWorldMarkerRemoved(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	var targetMarker *donburi.Entry = nil

	for m := range component.WorldMarker.Iter(ecs.World) {
		if component.WorldMarker.Get(m).Type == model.WorldMarkerType(*event.Icon) {
			targetMarker = m

			break
		}
	}

	if targetMarker == nil {
		return
	}

	ecs.World.Remove(targetMarker.Entity())

	return
}

func handleWorldMarkerPlaced(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	found := false

	for m := range component.WorldMarker.Iter(ecs.World) {
		marker := component.WorldMarker.Get(m)
		if marker.Type == model.WorldMarkerType(*event.Icon-1) {
			marker.Position[0] = float64(*event.X) / 100 * 25
			marker.Position[1] = float64(*event.Y) / 100 * 25
			found = true

			break
		}
	}

	if !found {
		entry.NewWorldMarker(ecs, model.WorldMarkerType(*event.Icon-1), f64.Vec2{
			float64(*event.X) / 100 * 25,
			float64(*event.Y) / 100 * 25,
		})
	}
}

func handleBeginCast(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	caster := s.EntryMap[*event.SourceID]
	if caster == nil {
		return
	}

	target := s.EntryMap[*event.TargetID]
	if target == nil {
		return
	}

	casterInstanceID := 0
	targetInstanceID := 0

	if event.SourceInstance != nil {
		casterInstanceID = int(*event.SourceInstance) - 1
	}

	if event.TargetInstance != nil {
		targetInstanceID = int(*event.TargetInstance) - 1
	}

	casterInstance := component.Status.Get(caster).Instances[casterInstanceID]
	targetInstance := component.Status.Get(target).Instances[targetInstanceID]

	skill := skills.QueryCastingSkill(event.Ability.ToSkill(*event.Duration))
	skill.StartTick = event.LocalTick

	if entry.GetGlobal(ecs).Debug && component.Status.Get(caster).Role == role.NPC {
		log.Println("NPC begin cast", skill.ID, skill.Name, *event.Duration)
	}

	s.Cast(ecs, casterInstance, targetInstance, skill, event.LocalTick)

	return
}

func handleApplyBuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	ability := (*event.Ability).ToBuff()
	ability.ApplyTick = event.LocalTick

	if event.Duration != nil {
		ability.Duration = *event.Duration
	}

	status.BuffList.Add(ability)
	// TODO implement buff effect(removeCallback) to do this work
	if ability.ID == 1000418 {
		status.SetDeath(false)
	}

	return
}

func handleApplyDebuff(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	sourceTarget := s.EntryMap[*event.SourceID]

	buffTarget := s.EntryMap[*event.TargetID]
	if buffTarget == nil {
		return
	}

	status := component.Status.Get(buffTarget)
	ability := (*event.Ability).ToBuff()
	ability.ECS = ecs
	ability.Source = sourceTarget
	ability.Target = buffTarget
	ability.RemoveCallback = buffs.BuffRemoveCallBackDB[ability.ID]
	ability.Type = model.Debuff
	ability.ApplyTick = event.LocalTick
	ability.Duration = *event.Duration
	status.BuffList.Add(ability)

	return
}

func (s *System) updateEventTargetStatus(event fflogs.FFLogsEvent) {
	instanceID := 0
	if event.TargetInstance != nil {
		instanceID = int(*event.TargetInstance) - 1
	}

	status := component.Status.Get(s.EntryMap[*event.TargetID])
	target := component.Status.Get(s.EntryMap[*event.TargetID])
	target.Instances[instanceID].LastActive = event.LocalTick

	if event.TargetMarker != nil {
		status.Marker = *event.TargetMarker
	} else {
		status.Marker = 0
	}
}

func (s *System) updateEventSourceStatus(event fflogs.FFLogsEvent) {
	instanceID := 0
	if event.SourceInstance != nil {
		instanceID = int(*event.SourceInstance) - 1
	}

	source := component.Status.Get(s.EntryMap[*event.SourceID])
	status := component.Status.Get(s.EntryMap[*event.SourceID])
	source.Instances[instanceID].LastActive = event.LocalTick

	if event.SourceMarker != nil {
		status.Marker = *event.SourceMarker
	} else {
		status.Marker = 0
	}
}

func aurasToBuffs(auras []fflogs.Aura) []*model.Buff {
	buffs := make([]*model.Buff, len(auras))
	for i, aura := range auras {
		buffs[i] = &model.Buff{
			ID:       aura.Ability,
			Name:     aura.Name,
			Icon:     aura.Icon,
			Stacks:   int(aura.Stacks),
			Duration: 0,
		}
	}

	return buffs
}

func handleCast(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	caster := s.EntryMap[*event.SourceID]
	if caster == nil {
		return
	}

	casterInstanceID := 0
	targetInstanceID := 0

	if event.SourceInstance != nil {
		casterInstanceID = int(*event.SourceInstance) - 1
	}

	if event.TargetInstance != nil {
		targetInstanceID = int(*event.TargetInstance) - 1
	}

	if event.Ability == nil {
		return
	}

	if event.TargetID == nil || s.EntryMap[*event.TargetID] == nil {
		return
	}

	casterInstance := component.Status.Get(caster).Instances[casterInstanceID]
	targetInstance := component.Status.Get(s.EntryMap[*event.TargetID]).Instances[targetInstanceID]

	skill := skills.QueryCastingSkill(event.Ability.ToSkill(0))
	skill.StartTick = event.LocalTick

	if entry.GetGlobal(ecs).Debug && component.Status.Get(caster).Role == role.NPC {
		log.Println("NPC inst-cast", *event.SourceID, casterInstanceID, skill.ID, skill.Name)
	}

	s.Cast(ecs, casterInstance, targetInstance, skill, event.LocalTick)
}

func handleDamage(s *System, ecs *ecs.ECS, eventSource *donburi.Entry, event fflogs.FFLogsEvent) {
	// source := s.EntryMap[*event.SourceID]
	target, ok := s.EntryMap[*event.TargetID]
	if !ok {
		return
	}

	targetStatus := component.Status.Get(target)
	targetInstance := targetStatus.Instances[0]

	relatedBuffs := make([]*model.BasicBuffInfo, 0)

	buffs := event.Buffs

	buffStrs := strings.Split(buffs, ".")
	for _, buffStr := range buffStrs {
		if buffStr == "" {
			continue
		}

		buffID, err := strconv.Atoi(buffStr)
		if err != nil {
			log.Println("failed to parse buff id", buffStr)

			continue
		}

		buffInfo := model.GetBuffInfo(int64(buffID))
		if buffInfo != nil {
			relatedBuffs = append(relatedBuffs, buffInfo)
		} else {
			log.Println("failed to find buff info for", buffID)
		}
	}

	var amount int64 = 0
	if event.Amount != nil {
		amount = *event.Amount
	}

	var multiplier float64 = 1
	if event.Multiplier != nil {
		multiplier = *event.Multiplier
	}

	damageTakenEntry := model.DamageTaken{
		Tick:         event.LocalTick,
		Type:         model.DamageType(event.Ability.Type),
		SourceID:     *event.SourceID,
		Ability:      event.Ability.ToSkill(0),
		Amount:       amount,
		Multiplier:   multiplier,
		RelatedBuffs: relatedBuffs,
	}

	targetInstance.AddDamageTaken(damageTakenEntry)
}
