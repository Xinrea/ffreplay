package ui

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/furex/v2"
)

func ProgressBarView() *furex.View {
	global := entry.GetGlobal(ecsInstance)
	newHandler := furex.ViewHandler{
		Update: func(v *furex.View) {
			current := float64(entry.GetTick(ecsInstance)) / 60
			progress := 0.0
			if global.FightDuration.Load() > 0 {
				progress = current / (float64(global.FightDuration.Load()) / 1000)
			}
			if bar, ok := v.MustGetByID("bar").Handler.Extra.(*Bar); ok {
				bar.Progress = progress
			}
			if text, ok := v.MustGetByID("progress").Handler.Extra.(*Text); ok {
				text.Content = fmt.Sprintf(
					"%s / %s",
					formatDuration(current),
					formatDuration(float64(global.FightDuration.Load())/1000),
				)
			}
			if text, ok := v.MustGetByID("speed").Handler.Extra.(*Text); ok {
				text.Content = fmt.Sprintf("当前速度：%.1f", float64(global.Speed)/10)
			}
		},
	}

	view := createProgressBar(newHandler)

	addTextLine := func(top int, content string) {
		view.AddChild(furex.NewView(
			furex.MarginTop(top),
			furex.Height(13),
			furex.Handler(&Text{
				Align:        furex.AlignItemEnd,
				Content:      content,
				Color:        color.White,
				Shadow:       true,
				ShadowOffset: 2,
				ShadowColor:  color.NRGBA{22, 45, 87, 128},
			}),
		))
	}

	addTextLine(10, "快退: 方向键左 | 快进: 方向键右 | 点击进度条跳转")
	addTextLine(5, "移动视角 W/A/S/D | 旋转视角: E/Q | 调试模式：`")
	addTextLine(5, "暂停: SPACE | 播放速度: 方向键（上下）| 重置: R")
	addTextLine(5, "锁定玩家: 1-8 | 点击小队列表锁定 | 解除锁定: ESC")

	return view
}

func createProgressBar(newHandler furex.ViewHandler) *furex.View {
	global := entry.GetGlobal(ecsInstance)
	view := furex.NewView(
		furex.Position(furex.PositionAbsolute),
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemEnd),
		furex.Justify(furex.JustifyEnd),
		furex.Handler(newHandler),
		furex.Bottom(20),
		furex.Right(20),
	)

	view.AddChild(furex.NewView(
		furex.ID("progress"),
		furex.Height(13),
		furex.Handler(&Text{
			Align:        furex.AlignItemEnd,
			Color:        color.White,
			Content:      "00:00 / 00:00",
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	))

	segments := fetchProgressSegments(global)

	view.AddChild(furex.NewView(
		furex.ID("bar"),
		furex.Width(250),
		furex.Height(13),
		furex.MarginTop(5),
		furex.MarginBottom(5),
		furex.Handler(&Bar{
			FG:           barAtlas.GetNineSlice("normal_bar_fg.png"),
			BG:           barAtlas.GetNineSlice("normal_bar_bg.png"),
			Segments:     segments,
			Interactable: true,
			ClickAt: func(c, p float64) {
				global.Tick = util.MSToTick(int64(float64(global.FightDuration.Load())*p)) * 10
				if p < c {
					global.Reset.Store(true)
				}
			},
		}),
	))

	view.AddChild(furex.NewView(
		furex.ID("speed"),
		furex.Height(13),
		furex.Handler(&Text{
			Align:        furex.AlignItemEnd,
			Color:        color.White,
			Content:      "当前速度：1.0",
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	))

	return view
}

func fetchProgressSegments(global *model.GlobalData) []float64 {
	// progress segments from phases
	segments := []float64{}

	for _, phase := range global.Phases {
		if phase == 0 {
			continue
		}

		segments = append(segments, float64(phase)/float64(util.MSToTick(global.FightDuration.Load())))
	}

	if len(segments) > 0 && segments[len(segments)-1] < 1 {
		segments = append(segments, 1)
	}

	return segments
}

func formatDuration(s float64) string {
	minutes := int(s) / 60
	seconds := int(s) % 60

	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
