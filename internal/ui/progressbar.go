package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
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

	view.Layout()

	return view
}

func createProgressBar(newHandler furex.ViewHandler) *furex.View {
	global := entry.GetGlobal(ecsInstance)
	view := furex.NewView(
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemEnd),
		furex.Justify(furex.JustifyEnd),
		furex.Handler(newHandler),
		furex.MarginBottom(UIPadding),
		furex.MarginRight(UIPadding),
	)

	cv := CheckBoxView(
		13,
		true,
		&global.RangeDisplay,
		"显示技能范围(测试)",
		nil,
	)

	cv.SetMarginBottom(12)
	view.AddChild(cv)

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

type euiProgressBar struct {
	widget       *widget.Widget
	width        int
	height       int
	progress     func() float64
	segments     []float64
	clickAt      func(current float64, clicked float64)
	lastProgress float64
}

func newEUIProgressBar(width, height int, segments []float64, progress func() float64, clickAt func(current, clicked float64)) *euiProgressBar {
	bar := &euiProgressBar{
		width:    width,
		height:   height,
		segments: segments,
		progress: progress,
		clickAt:  clickAt,
	}
	bar.widget = widget.NewWidget(
		widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position:  widget.RowLayoutPositionEnd,
			MaxWidth:  width,
			MaxHeight: height,
		}),
		widget.WidgetOpts.CursorHovered(input.CURSOR_POINTER),
		widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
			if args.Button != ebiten.MouseButtonLeft || bar.clickAt == nil {
				return
			}
			p := float64(args.OffsetX) / float64(max(1, bar.widget.Rect.Dx()))
			if p < 0 {
				p = 0
			}
			if p > 1 {
				p = 1
			}
			bar.clickAt(bar.currentProgress(), p)
		}),
	)
	return bar
}

func (b *euiProgressBar) GetWidget() *widget.Widget {
	return b.widget
}

func (b *euiProgressBar) SetLocation(rect image.Rectangle) {
	b.widget.Rect = rect
}

func (b *euiProgressBar) PreferredSize() (int, int) {
	return b.width, b.height
}

func (b *euiProgressBar) Validate() {}

func (b *euiProgressBar) Update(updObj *widget.UpdateObject) {
	b.widget.Update(updObj)
}

func (b *euiProgressBar) Render(screen *ebiten.Image) {
	b.widget.Render(screen)
	frame := b.widget.Rect
	progress := b.currentProgress()

	if len(b.segments) == 0 {
		barAtlas.GetNineSlice("normal_bar_bg.png").Draw(screen, frame, nil)
	} else {
		start := 0.0
		for _, end := range b.segments {
			barAtlas.GetNineSlice("normal_bar_bg.png").Draw(screen, image.Rect(
				frame.Min.X+int(start*float64(frame.Dx())),
				frame.Min.Y,
				frame.Min.X+int(end*float64(frame.Dx())),
				frame.Max.Y,
			), nil)
			start = end
		}
	}

	fgFrame := frame
	fgFrame.Max.X = fgFrame.Min.X + int(progress*float64(frame.Dx()))
	barAtlas.GetNineSlice("normal_bar_fg.png").Draw(screen, fgFrame, nil)
}

func (b *euiProgressBar) currentProgress() float64 {
	progress := 1.0
	if b.progress != nil {
		progress = b.progress()
	}
	if progress > 1 {
		progress = 1
	}
	if progress < 0 {
		progress = 0
	}
	b.lastProgress = progress
	return progress
}

func EUIProgressBarView(scale float64) *widget.Container {
	global := entry.GetGlobal(ecsInstance)
	if scale <= 0 {
		scale = 1
	}

	pad := int(float64(UIPadding) * scale)
	face := newEUIFace(13 * scale)
	textColor := color.White

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Right:  pad,
				Bottom: pad,
			}),
			widget.RowLayoutOpts.Spacing(int(5*scale)),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionEnd,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
		})),
	)

	alignEnd := func(w widget.HasWidget) {
		w.GetWidget().LayoutData = widget.RowLayoutData{Position: widget.RowLayoutPositionEnd}
	}

	rangeCheckbox := NewEUICheckbox(13, true, &global.RangeDisplay, "显示技能范围(测试)", nil, scale)
	alignEnd(rangeCheckbox)
	container.AddChild(rangeCheckbox)

	var progressText *widget.Text
	progressText = widget.NewText(
		widget.TextOpts.Text("00:00 / 00:00", &face, textColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionEnd}),
			widget.WidgetOpts.OnUpdate(func(w widget.HasWidget) {
				current := float64(entry.GetTick(ecsInstance)) / 60
				progressText.Label = fmt.Sprintf(
					"%s / %s",
					formatDuration(current),
					formatDuration(float64(global.FightDuration.Load())/1000),
				)
			}),
		),
	)
	container.AddChild(progressText)

	currentProgress := func() float64 {
		current := float64(entry.GetTick(ecsInstance)) / 60
		if global.FightDuration.Load() <= 0 {
			return 0
		}
		return current / (float64(global.FightDuration.Load()) / 1000)
	}

	bar := newEUIProgressBar(
		int(250*scale),
		int(13*scale),
		fetchProgressSegments(global),
		currentProgress,
		func(current, clicked float64) {
			global.Tick = util.MSToTick(int64(float64(global.FightDuration.Load())*clicked)) * 10
			if clicked < current {
				global.Reset.Store(true)
			}
		},
	)
	container.AddChild(bar)

	var speedText *widget.Text
	speedText = widget.NewText(
		widget.TextOpts.Text("当前速度：1.0", &face, textColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionEnd}),
			widget.WidgetOpts.OnUpdate(func(w widget.HasWidget) {
				speedText.Label = fmt.Sprintf("当前速度：%.1f", float64(global.Speed)/10)
			}),
		),
	)
	container.AddChild(speedText)

	addInstruction := func(content string) {
		container.AddChild(widget.NewText(
			widget.TextOpts.Text(content, &face, textColor),
			widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
			})),
		))
	}

	addInstruction("快退: 方向键左 | 快进: 方向键右 | 点击进度条跳转")
	addInstruction("移动视角 W/A/S/D | 旋转视角: E/Q | 调试模式：`")
	addInstruction("暂停: SPACE | 播放速度: 方向键（上下）| 重置: R")
	addInstruction("锁定玩家: 1-8 | 点击小队列表锁定 | 解除锁定: ESC")

	return container
}
