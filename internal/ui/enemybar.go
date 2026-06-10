package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

func EnemyBarsView() *furex.View {
	return furex.NewView(
		furex.MarginRight(UIPadding),
		furex.MarginTop(UIPadding),
		furex.Direction(furex.Column),
		furex.Handler(furex.ViewHandler{
			Update: func(v *furex.View) {
				v.RemoveAll()
				cnt := 0
				for e := range tag.Enemy.Iter(ecsInstance.World) {
					sprite := component.Sprite.Get(e)
					if !sprite.Initialized {
						continue
					}
					enemy := component.Status.Get(e)
					if (enemy.Role != role.Boss && enemy.Role != role.Special) ||
						!sprite.Instances[0].IsActive(entry.GetTick(ecsInstance)) {
						continue
					}
					v.AddChild(CreateEnemyBarView(cnt, e))
					cnt++
				}
				v.Layout()
			},
		}))
}

func CreateEnemyBarView(i int, enemy *donburi.Entry) *furex.View {
	sprite := component.Sprite.Get(enemy)
	status := component.Status.Get(enemy)

	view := furex.NewView(
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemStart),
	)
	nameView := furex.NewView(furex.Height(13), furex.Handler(&Text{
		Content:      status.Name,
		Color:        color.NRGBA{252, 183, 190, 255},
		Align:        furex.AlignItemStart,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{0, 0, 0, 128},
	}))

	nameCast := furex.NewView(
		furex.Width(500),
		furex.Direction(furex.Row),
		furex.Justify(furex.JustifySpaceBetween),
		furex.AlignItems(furex.AlignItemEnd),
	)
	nameCast.AddChild(nameView)
	nameCast.AddChild(createEnemyCastingView(sprite))
	view.AddChild(nameCast)

	view.AddChild(
		furex.NewView(
			furex.ID("bar"),
			furex.Width(500),
			furex.Height(10),
			furex.MarginTop(5),
			furex.Handler(&Bar{
				Progress: float64(status.HP) / float64(status.MaxHP),
				FG:       barAtlas.GetNineSlice("red_bar_fg.png"),
				BG:       barAtlas.GetNineSlice("red_bar_bg.png"),
			}),
		))

	view.AddChild(createEnemyHPTextView(status))

	bufflist := BuffListView(status.BuffList.Buffs())
	bufflist.Attrs.MarginTop = 5

	view.AddChild(bufflist)

	return view
}

func createEnemyCastingView(sprite *model.SpriteData) *furex.View {
	castView := furex.NewView(furex.Height(24), furex.Direction(furex.Column), furex.AlignItems(furex.AlignItemEnd))

	if sprite.Instances[0].GetCast() != nil {
		cast := sprite.Instances[0].GetCast()
		castView.AddChild(furex.NewView(furex.Width(210), furex.Height(12), furex.Handler(&Bar{
			Progress: float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast),
			BG:       castAtlas.GetNineSlice("casting_frame.png"),
			FG:       castAtlas.GetNineSlice("casting_fg.png"),
		})))
		castView.AddChild(furex.NewView(furex.Height(12), furex.MarginTop(-5), furex.Handler(&Text{
			Align:        furex.AlignItemEnd,
			Content:      cast.Name,
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 1,
			ShadowColor:  color.NRGBA{240, 152, 0, 128},
		})))
	}

	return castView
}

func createEnemyHPTextView(status *model.StatusData) *furex.View {
	hpView := furex.NewView(furex.Width(500), furex.Direction(furex.Row), furex.Justify(furex.JustifySpaceBetween))
	hpView.AddChild(furex.NewView(furex.MarginTop(5), furex.Height(13), furex.Handler(&Text{
		Content:      formatInt(status.HP) + " / " + formatInt(status.MaxHP),
		Color:        color.NRGBA{252, 183, 190, 255},
		Align:        furex.AlignItemStart,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{0, 0, 0, 128},
	}),
	))
	hpView.AddChild(furex.NewView(furex.MarginTop(5), furex.Height(13), furex.Handler(&Text{
		Content:      fmt.Sprintf("%.2f%%", float64(status.HP)/float64(status.MaxHP)*100),
		Color:        color.NRGBA{252, 183, 190, 255},
		Align:        furex.AlignItemEnd,
		Shadow:       true,
		ShadowOffset: 2,
		ShadowColor:  color.NRGBA{0, 0, 0, 128},
	})))

	return hpView
}

func formatInt(n int) string {
	// 将 int64 转换为字符串
	str := strconv.FormatInt(int64(n), 10)

	// 计算整数的长度
	length := len(str)
	if length <= 3 {
		return str // 如果长度小于等于3，直接返回
	}

	// 使用 strings.Builder 来构建结果字符串
	var builder strings.Builder

	for i, digit := range str {
		// 每三位添加一个逗号
		if i != 0 && (length-i)%3 == 0 {
			builder.WriteRune(',')
		}

		builder.WriteRune(digit)
	}

	return builder.String()
}

func EUIEnemyBarsView(scale float64) *widget.Container {
	if scale <= 0 {
		scale = 1
	}

	pad := int(float64(UIPadding) * scale)
	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:   pad,
				Right: pad,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
	)

	view.GetWidget().OnUpdate = func(w widget.HasWidget) {
		view.RemoveChildren()
		cnt := 0
		for e := range tag.Enemy.Iter(ecsInstance.World) {
			sprite := component.Sprite.Get(e)
			if !sprite.Initialized {
				continue
			}
			enemy := component.Status.Get(e)
			if (enemy.Role != role.Boss && enemy.Role != role.Special) ||
				!sprite.Instances[0].IsActive(entry.GetTick(ecsInstance)) {
				continue
			}
			view.AddChild(CreateEUIEnemyBarView(cnt, e, scale))
			cnt++
		}
	}

	return view
}

func CreateEUIEnemyBarView(i int, enemy *donburi.Entry, scale float64) *widget.Container {
	sprite := component.Sprite.Get(enemy)
	status := component.Status.Get(enemy)
	face := newEUIFace(13 * scale)
	nameColor := color.NRGBA{252, 183, 190, 255}

	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(int(5*scale)),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
		})),
	)

	nameCast := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(int(500*scale), int(24*scale))),
	)
	nameCast.AddChild(widget.NewText(
		widget.TextOpts.Text(status.Name, &face, nameColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	))
	nameCast.AddChild(createEUIEnemyCastingView(sprite, scale))
	view.AddChild(nameCast)

	hpProgress := 0.0
	if status.MaxHP > 0 {
		hpProgress = float64(status.HP) / float64(status.MaxHP)
	}
	view.AddChild(NewEUIBar(
		int(500*scale),
		int(10*scale),
		barAtlas.GetNineSlice("red_bar_bg.png"),
		barAtlas.GetNineSlice("red_bar_fg.png"),
		hpProgress,
		nil,
		widget.RowLayoutData{Position: widget.RowLayoutPositionEnd},
	))

	view.AddChild(createEUIEnemyHPTextView(status, scale))

	buffs := EUIBuffListView(status.BuffList.Buffs(), scale)
	buffs.GetWidget().LayoutData = widget.RowLayoutData{Position: widget.RowLayoutPositionEnd}
	view.AddChild(buffs)

	return view
}

func createEUIEnemyCastingView(sprite *model.SpriteData, scale float64) *widget.Container {
	face := newEUIFace(12 * scale)
	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(int(210*scale), int(24*scale))),
	)

	cast := sprite.Instances[0].GetCast()
	if cast == nil {
		return view
	}

	view.AddChild(NewEUIBar(
		int(210*scale),
		int(12*scale),
		castAtlas.GetNineSlice("casting_frame.png"),
		castAtlas.GetNineSlice("casting_fg.png"),
		float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick))/float64(cast.Cast),
		nil,
		widget.RowLayoutData{Position: widget.RowLayoutPositionEnd},
	))
	view.AddChild(widget.NewText(
		widget.TextOpts.Text(cast.Name, &face, color.White),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionEnd,
		})),
	))

	return view
}

func createEUIEnemyHPTextView(status *model.StatusData, scale float64) *widget.Container {
	face := newEUIFace(13 * scale)
	nameColor := color.NRGBA{252, 183, 190, 255}
	view := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(int(500*scale), int(13*scale))),
	)

	view.AddChild(widget.NewText(
		widget.TextOpts.Text(formatInt(status.HP)+" / "+formatInt(status.MaxHP), &face, nameColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
	))

	percent := 0.0
	if status.MaxHP > 0 {
		percent = float64(status.HP) / float64(status.MaxHP) * 100
	}
	view.AddChild(widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%.2f%%", percent), &face, nameColor),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
	))

	return view
}
