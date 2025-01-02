package data

import (
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

// FetchLogEvents fetch events from fflogs by report code with fight id.
func FetchLogEvents(c *fflogs.FFLogsClient, code string, fight fflogs.ReportFight) (map[int64]InstanceStatus, []fflogs.FFLogsEvent) {
	status := make(map[int64]InstanceStatus)
	events := c.QueryFightEvents(code, fight)
	startTime := int64(fight.StartTime)

	extractTargetStatusFromEvent := func(e fflogs.FFLogsEvent) StatusEvent {
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
	extractSourceStatusFromEvent := func(e fflogs.FFLogsEvent) StatusEvent {
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
	// preprocess events, convert timestamp to tick
	for i := range events {
		events[i].LocalTick = int64(events[i].Timestamp-startTime) / 1000 * 60
		// make sure every tick has at most one status event
		if events[i].SourceID != nil && events[i].SourceResources != nil {
			sourceID := *events[i].SourceID
			if _, ok := status[sourceID]; !ok {
				status[sourceID] = make(InstanceStatus)
			}
			instanceID := 1
			if events[i].SourceInstance != nil {
				instanceID = int(*events[i].SourceInstance)
			}
			newStatus := extractSourceStatusFromEvent(events[i])
			if len(status[sourceID][instanceID]) == 0 || status[sourceID][instanceID][len(status[sourceID][instanceID])-1].Tick != events[i].LocalTick {
				status[sourceID][instanceID] = append(status[sourceID][instanceID], newStatus)
			} else {
				status[sourceID][instanceID][len(status[sourceID][instanceID])-1] = newStatus
			}
		}
		if events[i].TargetID != nil && events[i].TargetResources != nil {
			targetID := *events[i].TargetID
			if _, ok := status[targetID]; !ok {
				status[targetID] = make(InstanceStatus)
			}
			instanceID := 1
			if events[i].TargetInstance != nil {
				instanceID = int(*events[i].TargetInstance)
			}
			newStatus := extractTargetStatusFromEvent(events[i])
			if len(status[targetID][instanceID]) == 0 || status[targetID][instanceID][len(status[targetID][instanceID])-1].Tick != events[i].LocalTick {
				status[targetID][instanceID] = append(status[targetID][instanceID], newStatus)
			} else {
				status[targetID][instanceID][len(status[targetID][instanceID])-1] = newStatus
			}
		}
	}
	return status, events
}

func PreloadAbilityInfo(events []fflogs.FFLogsEvent, counter *atomic.Int32) {
	for i := range events {
		counter.Add(1)
		if events[i].Ability != nil {
			_ = texture.NewAbilityTexture(events[i].Ability.AbilityIcon)
			_ = model.GetAction(events[i].Ability.Guid)
		}
	}
}
