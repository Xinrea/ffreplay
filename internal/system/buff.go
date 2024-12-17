package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/game/skills"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type RemoveCallbackDB map[int64]func(*model.Buff, *ecs.ECS, *donburi.Entry, *donburi.Entry)

var buffRemoveDB RemoveCallbackDB

func init() {
	buffRemoveDB = make(RemoveCallbackDB)
	// register all callback info in db
	buffRemoveDB[1004004] = func(b *model.Buff, ecs *ecs.ECS, source *donburi.Entry, target *donburi.Entry) {
		entry.CastSkill(ecs, 1000, 1000, skills.NewSkillTestCircle(ecs, source, target, 1.5))
	}
}

func (s *System) BuffUpdate(ecs *ecs.ECS, tick int64) {
	for e := range component.Status.Iter(ecs.World) {
		status := component.Status.Get(e)
		if status.Casting != nil {
			// check cast end
			if status.Casting.Duration <= util.TickToMS(tick-status.Casting.ApplyTick) {
				status.Casting = nil
			}
		}
	}
	if s.InReplay {
		return
	}
	for e := range tag.Buffable.Iter(ecs.World) {
		component.Status.Get(e).BuffList.UpdateExpire(tick)
	}
}
