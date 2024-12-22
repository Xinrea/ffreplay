package data

import (
	"sort"
	"sync/atomic"

	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
)

// FetchLogEvents fetch events from fflogs by report code with fight id.
func FetchLogEvents(c *fflogs.FFLogsClient, code string, fight fflogs.ReportFight) []fflogs.FFLogsEvent {
	events := c.QueryFightEvents(code, fight)
	startTime := int64(fight.StartTime)
	// preprocess events, convert timestamp to tick
	for i := range events {
		events[i].LocalTick = int64(events[i].Timestamp-startTime) / 1000 * 60
	}
	// sort events by tick
	sort.Slice(events, func(i, j int) bool {
		return events[i].LocalTick < events[j].LocalTick
	})
	return events
}

func PreloadAbilityInfo(events []fflogs.FFLogsEvent, counter *atomic.Int32) {
	for i := range events {
		counter.Add(1)
		if events[i].Ability != nil {
			_ = texture.NewAbilityTexture(events[i].Ability.AbilityIcon)
			_ = model.IsGCD(events[i].Ability.Guid)
		}
	}
}
