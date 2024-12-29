package ui

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/yohamta/furex/v2"
)

func ProgressBarView() *furex.View {
	global := entry.GetGlobal(ecsInstance)
	newHandler := furex.NewHandler(furex.HandlerOpts{
		Update: func(v *furex.View) {
			current := float64(entry.GetTick(ecsInstance)) / 60
			progress := 0.0
			if global.FightDuration.Load() > 0 {
				progress = current / (float64(global.FightDuration.Load()) / 1000)
			}
			v.MustGetByID("bar").Handler.(*Bar).Progress = progress
			v.MustGetByID("progress").Handler.(*Text).Content = fmt.Sprintf("%s / %s", formatDuration(current), formatDuration(float64(global.FightDuration.Load())/1000))
			v.MustGetByID("speed").Handler.(*Text).Content = fmt.Sprintf("当前速度：%.1f", float64(global.Speed)/10)
		},
	})
	view := &furex.View{
		Position:   furex.PositionAbsolute,
		Direction:  furex.Column,
		AlignItems: furex.AlignItemEnd,
		Justify:    furex.JustifyEnd,
		Handler:    newHandler,
	}
	view.SetBottom(20)
	view.SetRight(20)
	view.AddChild(&furex.View{
		ID:     "progress",
		Height: 13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Color:        color.White,
			Content:      "00:00 / 00:00",
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})
	view.AddChild(&furex.View{
		ID:           "bar",
		Width:        250,
		Height:       13,
		MarginTop:    5,
		MarginBottom: 5,
		Handler: &Bar{
			FG: barAtlas.GetNineSlice("normal_bar_fg.png"),
			BG: barAtlas.GetNineSlice("normal_bar_bg.png"),
		},
	})
	view.AddChild(&furex.View{
		ID:     "speed",
		Height: 13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Color:        color.White,
			Content:      "当前速度：1.0",
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})

	view.AddChild(&furex.View{
		MarginTop: 10,
		Height:    13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Content:      "快退: 方向键左 | 快进: 方向键右 | 点击进度条跳转",
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})
	view.AddChild(&furex.View{
		MarginTop: 5,
		Height:    13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Content:      "移动视角 W/A/S/D | 旋转视角: E/Q | 调试模式：`",
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})
	view.AddChild(&furex.View{
		MarginTop: 5,
		Height:    13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Content:      "暂停: SPACE | 播放速度: 方向键（上下）| 重置: R",
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})
	view.AddChild(&furex.View{
		MarginTop: 5,
		Height:    13,
		Handler: &Text{
			Align:        furex.AlignItemEnd,
			Content:      "锁定玩家: 1-8 | 点击小队列表锁定 | 解除锁定: ESC",
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		},
	})
	return view
}

func formatDuration(s float64) string {
	minutes := int(s) / 60
	seconds := int(s) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
