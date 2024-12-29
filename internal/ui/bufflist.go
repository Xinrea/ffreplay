package ui

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/furex/v2"
)

type BuffList struct {
	Buffs *model.BuffList
}

func (bl *BuffList) Update(v *furex.View) {
	v.RemoveAll()
	for _, b := range bl.Buffs.Buffs() {
		v.AddChild(BuffView(b))
	}
}
