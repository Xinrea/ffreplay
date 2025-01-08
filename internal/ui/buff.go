package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/furex/v2"
)

const (
	BuffWidth          = 24
	BuffHeight         = 32
	BuffStackFontSize  = 13
	BuffStackTop       = 2
	BuffStackLeft      = 9
	BuffStackShadow    = 4
	BuffRemainTop      = -6
	BuffRemainFontSize = 12
)

func BuffView(buff *model.Buff) *furex.View {
	view := &furex.View{
		Attrs: furex.ViewAttrs{
			Direction: furex.Column,
		},
	}
	view.AddChild(
		furex.NewView(
			furex.Width(BuffWidth),
			furex.Height(BuffHeight),
			furex.Handler(&Sprite{Texture: buff.Texture()}),
		),
	)

	if buff.Stacks > 1 {
		view.AddChild(
			furex.NewView(
				furex.Position(furex.PositionAbsolute),
				furex.Width(BuffStackFontSize),
				furex.Height(BuffStackFontSize),
				furex.Top(BuffStackTop),
				furex.Left(BuffStackLeft),
				furex.Handler(&Text{
					Align:        furex.AlignItemEnd,
					Content:      strconv.Itoa(buff.Stacks),
					Color:        color.White,
					Shadow:       true,
					ShadowOffset: BuffStackShadow,
					ShadowColor:  color.NRGBA{0, 0, 0, 200},
				})))
	}

	view.AddChild(furex.NewView(furex.MarginTop(BuffRemainTop), furex.Height(BuffRemainFontSize), furex.Handler(&Text{
		Align:        furex.AlignItemCenter,
		Content:      formatSeconds(buff.Remain),
		Color:        color.White,
		Shadow:       true,
		ShadowOffset: 1,
		ShadowColor:  color.NRGBA{0, 0, 0, 128},
	})))

	return view
}

const secondsInMinute = 60

func formatSeconds(seconds int64) string {
	minutes := seconds / secondsInMinute
	hours := minutes / secondsInMinute

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
