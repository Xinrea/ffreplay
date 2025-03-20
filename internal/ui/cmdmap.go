package ui

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

func (c *CommandHandler) mapHandler(cmds []string) {
	if len(cmds) == 0 {
		c.AddError("Invalid map command")

		return
	}

	switch cmds[0] {
	case "list":
		c.mapList()
	case "phase":
		c.mapPhase(cmds)
	case "set":
		c.mapSet(cmds)
	default:
		c.AddError("Invalid map command")
	}
}

func (c *CommandHandler) mapList() {
	mapids := []string{}
	for k, m := range model.MapCache {
		mapids = append(mapids, fmt.Sprintf("%d-%s", k, m.Path))
	}

	sort.Strings(mapids)

	for _, m := range mapids {
		c.AddResult(m)
	}
}

func (c *CommandHandler) mapPhase(cmds []string) {
	id, _ := strconv.Atoi(cmds[1])
	if m, ok := model.MapCache[id]; ok {
		for i := range m.Phases {
			c.AddResult(fmt.Sprintf("Phase %d", i))
		}
	} else {
		c.AddError("Invalid map id")
	}
}

func (c *CommandHandler) mapSet(cmds []string) {
	mapData := component.Map.Get(component.Map.MustFirst(ecsInstance.World))
	cameraData := entry.GetCamera()
	id, _ := strconv.Atoi(cmds[1])

	if m, ok := model.MapCache[id]; ok {
		mapData.Config = m.Load()
		current := mapData.Config.Maps[mapData.Config.CurrentMap]
		cameraData.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)

		c.AddResult("Setup map " + cmds[1])

		if len(cmds) > 2 {
			phase, _ := strconv.Atoi(cmds[2])
			if phase < len(mapData.Config.Phases) && phase >= 0 {
				mapData.Config.CurrentPhase = phase

				c.AddResult("Setup phase " + cmds[2])
			}
		}
	} else {
		c.AddError("Invalid map id")
	}
}
