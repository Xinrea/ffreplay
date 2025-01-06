package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/furex/v2"
)

func BuffView(buff *model.Buff) *furex.View {
	view := &furex.View{
		Attrs: furex.ViewAttrs{
			Direction: furex.Column,
		},
	}
	view.AddChild(furex.NewView(furex.Width(24), furex.Height(32), furex.Handler(&Sprite{Texture: buff.Texture()})))
	if buff.Stacks > 1 {
		view.AddChild(furex.NewView(furex.Position(furex.PositionAbsolute), furex.Width(13), furex.Height(13), furex.Top(2), furex.Left(9), furex.Handler(&Text{
			Align:        furex.AlignItemEnd,
			Content:      strconv.Itoa(buff.Stacks),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 4,
			ShadowColor:  color.NRGBA{0, 0, 0, 200},
		})))
	}
	view.AddChild(furex.NewView(furex.MarginTop(-6), furex.Height(12), furex.Handler(&Text{
		Align:        furex.AlignItemCenter,
		Content:      formatSeconds(buff.Remain),
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  color.NRGBA{0, 0, 0, 128},
	})))
	return view
}

func formatSeconds(seconds int64) string {
	minutes := seconds / 60
	hours := minutes / 60
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	if seconds > 0 {
		return fmt.Sprintf("%d", seconds)
	}
	return ""
}
