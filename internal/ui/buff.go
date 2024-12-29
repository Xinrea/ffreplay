package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/yohamta/furex/v2"
)

func BuffView(buff *model.Buff) *furex.View {
	view := &furex.View{
		Direction: furex.Column,
	}
	view.AddChild(&furex.View{
		Width:  24,
		Height: 32,
		Handler: &Sprite{
			Texture: texture.NewNineSlice(buff.Texture(), 0, 0, 0, 0),
		}})
	if buff.Stacks > 1 {
		view.AddChild(&furex.View{
			Position: furex.PositionAbsolute,
			Width:    24,
			Height:   32,
			Handler: &Sprite{
				Texture: texture.NewNineSlice(model.BuffStackBG, 0, 0, 0, 0),
			}})
		view.AddChild(&furex.View{
			Position: furex.PositionAbsolute,
			Width:    11,
			Height:   11,
			Top:      1,
			Left:     10,
			Handler: &Text{
				Align:   furex.AlignItemEnd,
				Content: strconv.Itoa(buff.Stacks),
				Color:   color.Black,
			}})
	}
	view.AddChild(&furex.View{
		MarginTop: -6,
		Height:    12,
		Handler: &Text{
			Align:        furex.AlignItemCenter,
			Content:      formatSeconds(buff.Remain),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 1,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		}})
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
