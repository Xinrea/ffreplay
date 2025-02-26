package buffs

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type RemoveCallbackDB map[int64]func(*model.Buff, *ecs.ECS, *donburi.Entry, *donburi.Entry)

var BuffRemoveCallBackDB RemoveCallbackDB = map[int64]func(*model.Buff, *ecs.ECS, *donburi.Entry, *donburi.Entry){
	1004158: general_clear_tether,
	1002255: general_clear_tether,
	1003587: general_clear_tether,
}

func general_clear_tether(buff *model.Buff, ecs *ecs.ECS, from *donburi.Entry, on *donburi.Entry) {
	status := component.Status.Get(on)
	status.ClearTether()
}
