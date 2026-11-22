package planner

import (
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/model/role"
)

// These are local artwork IDs, not FFLogs aura IDs. Several game statuses
// can use the same icon; their names are searchable aliases of one thumbnail.
//
//go:embed buff_icons.tsv
var buffIndex string

const MaxBuffIcons = 8

type BuffIcon struct {
	ID                  int
	Name, Asset, Search string
}

var buffIcons, buffByID = func() ([]BuffIcon, map[int]BuffIcon) {
	var all []BuffIcon

	byID := make(map[int]BuffIcon)

	for line := range strings.SplitSeq(strings.TrimSpace(buffIndex), "\n") {
		key, names, _ := strings.Cut(line, "\t")

		id, err := strconv.Atoi(key)
		if err != nil {
			panic("invalid generated buff icon ID: " + key)
		}

		name, _, _ := strings.Cut(names, "\t")
		if name == "" {
			name = "图标 " + key
		}

		icon := BuffIcon{
			id,
			name,
			fmt.Sprintf("%06d-%06d.png", id/1000*1000, id),
			strings.ToLower(line),
		}
		all = append(all, icon)
		byID[id] = icon
	}

	return all, byID
}()

func FindBuffIcons(query string) []BuffIcon {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return buffIcons
	}

	var matches []BuffIcon

	for _, icon := range buffIcons {
		if strings.Contains(icon.Search, query) {
			matches = append(matches, icon)
		}
	}

	return matches
}

func LookupBuffIcon(id int) (BuffIcon, bool) {
	icon, ok := buffByID[id]

	return icon, ok
}

func (o Object) SupportsBuffs() bool {
	r := role.StringToRole(o.Kind)
	if r >= role.MT && r < role.Boss {
		return true
	}

	for _, g := range gameIcons {
		if g.tool.Key == o.Kind {
			return g.group == "职能"
		}
	}

	return false
}

func (o Object) validBuffs() bool {
	if len(o.Buffs) == 0 {
		return true
	}

	if !o.SupportsBuffs() || len(o.Buffs) > MaxBuffIcons {
		return false
	}

	for i, id := range o.Buffs {
		if _, ok := LookupBuffIcon(id); !ok || slices.Contains(o.Buffs[:i], id) {
			return false
		}
	}

	return true
}

// ToggleBuff edits only the current step and records one undoable change.
func (e *Editor) ToggleBuff(index, id int) bool {
	objects := e.Step().Objects
	if index < 0 || index >= len(objects) || !objects[index].SupportsBuffs() {
		return false
	}

	if _, ok := LookupBuffIcon(id); !ok {
		return false
	}

	o := &objects[index]
	if i := slices.Index(o.Buffs, id); i >= 0 {
		e.Checkpoint()

		o.Buffs = slices.Delete(o.Buffs, i, i+1)
	} else {
		if len(o.Buffs) >= MaxBuffIcons {
			return false
		}

		e.Checkpoint()

		o.Buffs = append(o.Buffs, id)
	}

	return true
}
