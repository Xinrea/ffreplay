package ui

import (
	"fmt"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"golang.org/x/image/math/f64"
)

func (c *CommandHandler) playerHandler(cmds []string) {
	if len(cmds) == 0 {
		c.AddError("Invalid player command")

		return
	}

	switch cmds[0] {
	case "add":
		c.playerAdd(cmds)
	case "remove":
		c.playerRemove(cmds)
	case "headmarker":
		c.playerHeadMarker(cmds)
	default:
		c.AddError("Invalid player command")
	}
}

func (c *CommandHandler) playerAdd(cmds []string) {
	if len(cmds) < 3 {
		c.AddError("Invalid player add command")

		return
	}

	r := role.StringToRole(cmds[2])
	if r == -1 {
		c.AddError("Invalid player role")

		return
	}

	initialPos := f64.Vec2{0, 0}

	mapData := component.Map.Get(component.Map.MustFirst(ecsInstance.World))
	if mapData.Config != nil {
		current := mapData.Config.Maps[mapData.Config.CurrentMap]
		initialPos = f64.Vec2{current.Offset.X * 25, current.Offset.Y * 25}
	}

	p := entry.NewPlayer(ecsInstance, r, initialPos, &fflogs.PlayerDetail{
		ID:     c.player.idcnt,
		Name:   fmt.Sprintf("[%d]%s", c.player.idcnt, cmds[1]),
		Server: "ffreplay",
	})
	root.FilterByTagName("PartyList")[0].AddChild(NewPlayerItem(p))
	c.AddResult("Player " + cmds[1] + " added")
	c.player.idcnt += 1
}

func (c *CommandHandler) playerRemove(cmds []string) {
	global := entry.GetGlobal(ecsInstance)

	if len(cmds) < 2 {
		c.AddError("Invalid player remove command")

		return
	}

	for _, v := range root.FilterByTagName("PartyList") {
		for _, p := range v.GetChildren() {
			if p.Attrs.ID == cmds[1] {
				v.RemoveChild(p)
			}
		}
	}

	for p := range tag.Player.Iter(ecsInstance.World) {
		status := component.Status.Get(p)
		if strconv.Itoa(int(status.ID)) == cmds[1] {
			if global.TargetPlayer == p {
				global.TargetPlayer = nil
			}

			p.Remove()
			c.AddResult("Player " + status.Name + " removed")

			return
		}
	}
}

func (c *CommandHandler) playerHeadMarker(cmds []string) {
	if len(cmds) < 3 {
		c.AddError("Invalid player headmarker command")

		return
	}

	id, err := strconv.Atoi(cmds[2])
	if err != nil {
		c.AddError("Invalid player id")

		return
	}

	for p := range tag.Player.Iter(ecsInstance.World) {
		status := component.Status.Get(p)
		if int(status.ID) == id {
			switch cmds[1] {
			case "add":
				status.AddHeadMarker(model.HeadMarkerType1)
			case "remove":
				status.ClearHeadMarker()
			default:
				c.AddError("Invalid player headmarker command")
			}

			return
		}
	}
}
