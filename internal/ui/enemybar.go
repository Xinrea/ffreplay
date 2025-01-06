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
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

func EnemyBarsView() *furex.View {
	return furex.NewView(furex.Position(furex.PositionAbsolute), furex.Handler(furex.ViewHandler{
		Update: func(v *furex.View) {
			v.SetRight(0)
			v.SetTop(0)
			v.RemoveAll()
			cnt := 0
			for e := range tag.Enemy.Iter(ecsInstance.World) {
				sprite := component.Sprite.Get(e)
				if !sprite.Initialized {
					continue
				}
				enemy := component.Status.Get(e)
				if enemy.Role != role.Boss || !sprite.Instances[0].IsActive(entry.GetTick(ecsInstance)) {
					continue
				}
				v.AddChild(CreateEnemyBarView(cnt, e))
				cnt++
			}
		},
	}))
}

func CreateEnemyBarView(i int, enemy *donburi.Entry) *furex.View {
	sprite := component.Sprite.Get(enemy)
	status := component.Status.Get(enemy)

	view := furex.NewView(
		furex.Position(furex.PositionAbsolute),
		furex.Direction(furex.Column),
		furex.AlignItems(furex.AlignItemStart),
		furex.Right(520),
		furex.Top(20+100*i),
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
