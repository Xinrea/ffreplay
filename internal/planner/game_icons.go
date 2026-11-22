package planner

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed game_icons.tsv
var iconTable string

// Radius is the footprint at 100% game scale on a 40 m wide board.
type gameIcon struct {
	id    uint16
	group string
	tool  Tool
}

var gameIcons = func() []gameIcon {
	var icons []gameIcon
	for line := range strings.SplitSeq(strings.TrimSpace(iconTable), "\n") {
		f := strings.Split(line, "\t")
		id, e := strconv.ParseUint(f[0], 10, 16)
		r, err := strconv.ParseFloat(f[3], 64)
		if e != nil || err != nil {
			panic("invalid game icon catalogue: " + line)
		}
		icons = append(icons, gameIcon{uint16(id), f[1], Tool{"game-" + f[0], f[2], "asset/strategy/" + f[0] + ".png", r}})
	}
	return icons
}()

var PaletteNames = []string{"队员", "标点", "机制", "职业", "指示器", "敌方", "职能"}

func extraIcons(group string) []Tool {
	var out []Tool
	for _, g := range gameIcons {
		if g.group == group {
			out = append(out, g.tool)
		}
	}
	return out
}
