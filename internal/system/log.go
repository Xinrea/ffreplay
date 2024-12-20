package system

import (
	"log"
	"math"
	"sort"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/game/skills"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/fogleman/ease"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// Update only do update-work every 30 ticks, which means 0.5 second in default 60 TPS.
func (s *System) LogUpdate(ecs *ecs.ECS, tick int64) {
	if s.InReplay {
		s.replayUpdate(ecs, tick)
	}
}

func (s *System) replayUpdate(ecs *ecs.ECS, tick int64) {
	for e := range tag.GameObject.Iter(ecs.World) {
		id := component.Status.Get(e).ID
		line := s.EventLines[id]
		if line == nil {
			continue
		}
		if line.Cursor >= len(line.Events) {
			continue
		}
		// face and object is owned by instances, but they share status
		for i, sprite := range component.Sprite.Get(e).Instances {
			instanceID := i + 1
			// binary find the status event in line.Status
			index := sort.Search(len(line.Status[instanceID]), func(i int) bool {
				return line.Status[instanceID][i].Tick >= tick
			})
			// if not found, skip
			if index == len(line.Status[instanceID]) {
				continue
			}

			normalUpdate := func(status StatusEvent) {
				facing := status.Face + math.Pi/2
				sprite.Face = facing
				sprite.Object.UpdatePosition(status.Position)
				component.Status.Get(e).HP = status.HP
				component.Status.Get(e).MaxHP = status.MaxHP
				component.Status.Get(e).Mana = status.MP
				component.Status.Get(e).MaxMana = status.MaxMP
			}
			lerpUpdate := func(previous, status StatusEvent) {
				// lerping between two status event
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
			// apply status event
			// if is last event, just apply it
			status := line.Status[instanceID][index]
			if index == 0 || component.Status.Get(e).IsDead() {
				normalUpdate(status)
			} else {
				// not lerping for npc (normally invisible object in game)
				previous := line.Status[instanceID][index-1]
				if component.Status.Get(e).Role == model.NPC {
					normalUpdate(previous)
				} else {
					lerpUpdate(previous, status)
				}
			}
		}

		// tick is adjusted backward, reset line, and consume progress will handle this
		if line.Cursor > 0 && line.Events[line.Cursor-1].LocalTick > tick {
			line.Cursor = 0
			component.Status.Get(e).BuffList.Clear()
			for _, instance := range component.Sprite.Get(e).Instances {
				instance.Casting = nil
				instance.LastActive = -1
				instance.HistoryCasting = nil
			}
		}

		// consume all events until event that should not happen at this tick
		for line.Cursor < len(line.Events) && line.Events[line.Cursor].LocalTick <= tick {
			event := line.Events[line.Cursor]
			s.applyLog(ecs, tick, e, event)
			line.Cursor++
		}
	}
}

func (s *System) applyLog(ecs *ecs.ECS, tick int64, target *donburi.Entry, event fflogs.FFLogsEvent) {
	if event.TargetID != nil && s.EntryMap[*event.TargetID] != nil {
		instanceID := 0
		if event.TargetInstance != nil {
			instanceID = int(*event.TargetInstance) - 1
		}
		target := component.Sprite.Get(s.EntryMap[*event.TargetID])
		target.Instances[instanceID].LastActive = event.LocalTick
	}

	if event.SourceID != nil && s.EntryMap[*event.SourceID] != nil {
		instanceID := 0
		if event.SourceInstance != nil {
			instanceID = int(*event.SourceInstance) - 1
		}
		source := component.Sprite.Get(s.EntryMap[*event.SourceID])
		source.Instances[instanceID].LastActive = event.LocalTick
	}
	// {
	// "timestamp": 4134160,
	// "type": "combatantinfo",
	// "fight": 9,
	// "sourceID": 7,
	// "gear": [],
	// "auras": [
	// 	{
	// 		"source": 7,
	// 		"ability": 1000048,
	// 		"stacks": 1,
	// 		"icon": "216000-216202.png",
	// 		"name": "进食"
	// 	}
	// 	"level": 100,
	//  "simulatedCrit": 0.23880764904386953,
	//  "simulatedDirectHit": 0.3048368953880765
	// ]}
	if event.Type == fflogs.Combatantinfo {
		status := component.Status.Get(target)
		status.BuffList.SetBuffs(aurasToBuffs(event.Auras))
		return
	}
	// {
	// 	"timestamp": 4136478,
	// 	"type": "applybuff",
	// 	"sourceID": 7,
	// 	"targetID": 7,
	// 	"abilityGameID": 1003671,
	// 	"fight": 9,
	// 	"extraAbilityGameID": 34647,
	// 	"duration": 30000
	// }
	if event.Type == fflogs.Applybuff {
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

	if event.Type == fflogs.Applydebuff {
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
		ability.RemoveCallback = buffRemoveDB[ability.ID]
		ability.Type = model.Debuff
		ability.ApplyTick = event.LocalTick
		ability.Duration = *event.Duration
		status.BuffList.Add(ability)
		return
	}

	// 	{
	// 	"timestamp": 4142892,
	// 	"type": "refreshbuff",
	// 	"sourceID": 6,
	// 	"targetID": 6,
	// 	"abilityGameID": 1002677,
	// 	"fight": 9,
	// 	"duration": 40000
	// }
	if event.Type == fflogs.Refreshbuff {
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
	// {
	// 	"timestamp": 4139459,
	// 	"type": "removebuff",
	// 	"sourceID": 7,
	// 	"targetID": 7,
	// 	"abilityGameID": 1003658,
	// 	"fight": 9
	// }
	if event.Type == fflogs.Removebuff {
		buffTarget := s.EntryMap[*event.TargetID]
		if buffTarget == nil {
			return
		}
		status := component.Status.Get(buffTarget)
		ability := (*event.Ability).ToBuff()
		status.BuffList.Remove(ability)
		return
	}

	if event.Type == fflogs.RemoveDebuff {
		buffTarget := s.EntryMap[*event.TargetID]
		if buffTarget == nil {
			return
		}
		status := component.Status.Get(buffTarget)
		ability := (*event.Ability).ToBuff()
		status.BuffList.Remove(ability)
		return
	}

	if event.Type == fflogs.Begincast {
		caster := s.EntryMap[*event.SourceID]
		if caster == nil {
			return
		}
		instanceID := 0
		if event.SourceInstance != nil {
			instanceID = int(*event.SourceInstance) - 1
		}
		skill := skills.QuerySkill(event.Ability.ToSkill(*event.Duration))
		skill.StartTick = event.LocalTick

		component.Sprite.Get(caster).Instances[instanceID].Cast(skill)
		if caster.HasComponent(tag.Enemy) {
			log.Printf("[%d]%s[%d] begin cast [%d]%s on [%d]%s\n", component.Status.Get(caster).ID, component.Status.Get(caster).Name, instanceID+1, event.Ability.Guid, event.Ability.Name, component.Status.Get(target).ID, component.Status.Get(target).Name)
		}
		return
	}

	if event.Type == fflogs.Cast {
		caster := s.EntryMap[*event.SourceID]
		if caster == nil {
			return
		}
		instanceID := 0
		if event.SourceInstance != nil {
			instanceID = int(*event.SourceInstance) - 1
		}
		if event.Ability == nil {
			return
		}
		if event.TargetID == nil || s.EntryMap[*event.TargetID] == nil {
			return
		}
		target := s.EntryMap[*event.TargetID]
		skill := skills.QuerySkill(event.Ability.ToSkill(0))
		skill.StartTick = event.LocalTick
		if caster.HasComponent(tag.Enemy) {
			log.Printf("[%d]%s[%d] cast [%d]%s on [%d]%s\n", component.Status.Get(caster).ID, component.Status.Get(caster).Name, instanceID+1, skill.ID, event.Ability.Name, component.Status.Get(target).ID, component.Status.Get(target).Name)
		}
		s.Cast(ecs, caster, instanceID, target, 0, skill)
		return
	}
}

func aurasToBuffs(auras []fflogs.Aura) []model.Buff {
	buffs := make([]model.Buff, len(auras))
	for i, aura := range auras {
		buffs[i] = model.Buff{
			ID:       aura.Ability,
			Name:     aura.Name,
			Icon:     aura.Icon,
			Stacks:   int(aura.Stacks),
			Duration: 0,
		}
	}
	return buffs
}
