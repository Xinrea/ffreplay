package ui

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/furex/v2"
)

// BuffListView is a view for displaying a list of buffs.
// You can provide a model.BuffList to update buffs dynamically,
// or just a list of model.Buff to display static buffs.
func BuffListView(buffs any) *furex.View {
	if buffs, ok := buffs.([]*model.Buff); ok {
		view := furex.NewView(furex.ID("bufflist:static"))
		for _, b := range buffs {
			view.AddChild(BuffView(b))
		}

		return view
	}

	if buffs, ok := buffs.([]*model.BasicBuffInfo); ok {
		view := furex.NewView(furex.ID("bufflist:static"))
		for _, b := range buffs {
			view.AddChild(BuffView(&model.Buff{
				ID:   b.ID,
				Name: b.Name,
				Icon: b.Icon,
			}))
		}

		return view
	}

	if buffs, ok := buffs.(*model.BuffList); ok {
		return furex.NewView(furex.ID("bufflist:dynamic"), furex.Handler(&BuffListHandler{Buffs: buffs}))
	}

	return nil
}

type BuffListHandler struct {
	Buffs *model.BuffList

	handler furex.ViewHandler
}

func (bl *BuffListHandler) Handler() furex.ViewHandler {
	bl.handler.Extra = bl
	bl.handler.Update = bl.update

	return bl.handler
}

func (bl *BuffListHandler) update(v *furex.View) {
	if bl.Buffs == nil {
		return
	}

	v.RemoveAll()

	for _, b := range bl.Buffs.Buffs() {
		v.AddChild(BuffView(b))
	}

	v.Layout()
}
