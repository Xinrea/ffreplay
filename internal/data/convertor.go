package data

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

type StatusEvent struct {
	Tick     int64
	Face     float64
	HP       int
	MaxHP    int
	MP       int
	MaxMP    int
	Position vector.Vector
}

type InstanceStatus map[int][]StatusEvent

func extractTargetStatusFromEvent(e fflogs.FFLogsEvent) StatusEvent {
	return StatusEvent{
		Tick:     e.LocalTick,
		Face:     float64(e.TargetResources.Facing) / 100,
		HP:       int(e.TargetResources.HitPoints),
		MaxHP:    int(e.TargetResources.MaxHitPoints),
		MP:       int(e.TargetResources.Mp),
		MaxMP:    int(e.TargetResources.MaxMP),
		Position: vector.Vector{float64(e.TargetResources.X) / 100 * 25, float64(e.TargetResources.Y) / 100 * 25},
	}
}

func extractSourceStatusFromEvent(e fflogs.FFLogsEvent) StatusEvent {
	return StatusEvent{
		Tick:     e.LocalTick,
		Face:     float64(e.SourceResources.Facing) / 100,
		HP:       int(e.SourceResources.HitPoints),
		MaxHP:    int(e.SourceResources.MaxHitPoints),
		MP:       int(e.SourceResources.Mp),
		MaxMP:    int(e.SourceResources.MaxMP),
		Position: vector.Vector{float64(e.SourceResources.X) / 100 * 25, float64(e.SourceResources.Y) / 100 * 25},
	}
}

// FetchLogEvents fetch events from fflogs by report code with fight id.
func FetchLogEvents(
	c *fflogs.FFLogsClient,
	code string,
	fight fflogs.ReportFight,
) (map[int64]InstanceStatus, []fflogs.FFLogsEvent) {
	status := make(map[int64]InstanceStatus)
	events := c.QueryFightEvents(fflogs.RawQueryFightEvents, code, fight)
	startTime := int64(fight.StartTime)

	preloadCalculatedDamageEvents(c.QueryFightEvents(fflogs.RawQueryDamageTakenEvents, code, fight))
	preprocessDamageEvents(events)

	// preprocess events, convert timestamp to tick
	for i := range events {
		events[i].LocalTick = int64(events[i].Timestamp-startTime) / 1000 * 60
		processEvent(events[i], status)
	}

	return status, events
}

func processEvent(event fflogs.FFLogsEvent, status map[int64]InstanceStatus) {
	processSourceEvent(event, status)
	processTargetEvent(event, status)
}

func processSourceEvent(event fflogs.FFLogsEvent, status map[int64]InstanceStatus) {
	if event.SourceID != nil && event.SourceResources != nil {
		sourceID := *event.SourceID
		if _, ok := status[sourceID]; !ok {
			status[sourceID] = make(InstanceStatus)
		}

		instanceID := 1

		if event.SourceInstance != nil {
			instanceID = int(*event.SourceInstance)
		}

		newStatus := extractSourceStatusFromEvent(event)

		if len(status[sourceID][instanceID]) == 0 ||
			status[sourceID][instanceID][len(status[sourceID][instanceID])-1].Tick != event.LocalTick {
			status[sourceID][instanceID] = append(status[sourceID][instanceID], newStatus)
		} else {
			status[sourceID][instanceID][len(status[sourceID][instanceID])-1] = newStatus
		}
	}
}

func processTargetEvent(event fflogs.FFLogsEvent, status map[int64]InstanceStatus) {
	if event.TargetID != nil && event.TargetResources != nil {
		targetID := *event.TargetID
		if _, ok := status[targetID]; !ok {
			status[targetID] = make(InstanceStatus)
		}

		instanceID := 1

		if event.TargetInstance != nil {
			instanceID = int(*event.TargetInstance)
		}

		newStatus := extractTargetStatusFromEvent(event)

		if len(status[targetID][instanceID]) == 0 ||
			status[targetID][instanceID][len(status[targetID][instanceID])-1].Tick != event.LocalTick {
			status[targetID][instanceID] = append(status[targetID][instanceID], newStatus)
		} else {
			status[targetID][instanceID][len(status[targetID][instanceID])-1] = newStatus
		}
	}
}

func PreloadAbilityInfo(events []fflogs.FFLogsEvent, counter *atomic.Int32) {
	for i := range events {
		counter.Add(1)

		if events[i].Ability != nil {
			_ = texture.NewAbilityTexture(events[i].Ability.AbilityIcon)
			_ = model.GetAction(events[i].Ability.Guid)
			model.SetBuffInfoCache(events[i].Ability.Guid, &model.BasicBuffInfo{
				ID:   events[i].Ability.Guid,
				Name: events[i].Ability.Name,
				Icon: events[i].Ability.AbilityIcon,
			})
		}
	}
}

var packetMap = make(map[string]fflogs.FFLogsEvent)

func preloadCalculatedDamageEvents(events []fflogs.FFLogsEvent) {
	for i := range events {
		identifier := getDamageEventIdentifier(events[i])
		packetMap[identifier] = events[i]
	}
}

func preprocessDamageEvents(events []fflogs.FFLogsEvent) {
	for i := range events {
		if events[i].Type == fflogs.TDamage && events[i].PacketID != nil {
			identifier := getDamageEventIdentifier(events[i])

			if event, ok := packetMap[identifier]; ok {
				events[i].Buffs = event.Buffs
			}
		}
	}
}

func getDamageEventIdentifier(event fflogs.FFLogsEvent) string {
	if event.SourceID == nil || event.TargetID == nil || event.PacketID == nil {
		j, err := json.Marshal(event)
		log.Fatal("missing source/target id or packet id", string(j), err)
	}

	return fmt.Sprintf("%d-%d-%d", *event.SourceID, *event.TargetID, *event.PacketID)
}
