package ui

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/ebitenui/ebitenui/widget"
)

func EUIBuffListView(buffs []*model.Buff, scale float64) *widget.Container {
	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
	)

	for _, b := range buffs {
		view.AddChild(EUIBuffView(b, scale))
	}

	return view
}
