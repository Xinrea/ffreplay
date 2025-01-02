package ui

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/furex/v2"
)

// BuffListView is a view for displaying a list of buffs.
// You can provide a model.BuffList to update buffs dynamically,
// or just a list of model.Buff to display static buffs.
func BuffListView(buffs any) *furex.View {
	view := &furex.View{}
	if buffs, ok := buffs.([]*model.Buff); ok {
		view.ID = "bufflist:static"
		for _, b := range buffs {
			view.AddChild(BuffView(b))
		}
	}
	if buffs, ok := buffs.(*model.BuffList); ok {
		view.ID = "bufflist:dynamic"
		view.Handler = &BuffListHandler{
			Buffs: buffs,
		}
	}
	return view
}

type BuffListHandler struct {
	Buffs *model.BuffList
}

func (bl *BuffListHandler) Update(v *furex.View) {
	if bl.Buffs == nil {
		return
	}
	v.RemoveAll()
	for _, b := range bl.Buffs.Buffs() {
		v.AddChild(BuffView(b))
	}
}
