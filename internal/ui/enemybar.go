package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/furex/v2"
)

func EnemyBarsView() *furex.View {
	view := &furex.View{
		Position: furex.PositionAbsolute,
		Handler: furex.NewHandler(furex.HandlerOpts{
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
					if enemy.Role != model.Boss || !sprite.Instances[0].IsActive(entry.GetTick(ecsInstance)) {
						continue
					}
					v.AddChild(CreateEnemyBarView(cnt, e))
					cnt++
				}
			},
		}),
	}
	return view
}

func CreateEnemyBarView(i int, enemy *donburi.Entry) *furex.View {
	sprite := component.Sprite.Get(enemy)
	status := component.Status.Get(enemy)
	view := &furex.View{
		Position:   furex.PositionAbsolute,
		Direction:  furex.Column,
		AlignItems: furex.AlignItemStart,
	}
	view.SetRight(520)
	view.SetTop(20 + 50*i)
	nameView := &furex.View{
		Height: 13,
		Handler: &Text{
			Content:      status.Name,
			Color:        color.NRGBA{252, 183, 190, 255},
			Align:        furex.AlignItemStart,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		},
	}
	// add casting view
	castView := &furex.View{
		Height:     24,
		Direction:  furex.Column,
		AlignItems: furex.AlignItemEnd,
	}
	if sprite.Instances[0].GetCast() != nil {
		cast := sprite.Instances[0].GetCast()
		castView.AddChild(&furex.View{
			Width:  210,
			Height: 12,
			Handler: &Bar{
				Progress: float64(util.TickToMS(entry.GetTick(ecsInstance)-cast.StartTick)) / float64(cast.Cast),
				BG:       castAtlas.GetNineSlice("casting_frame.png"),
				FG:       castAtlas.GetNineSlice("casting_fg.png"),
			},
		})
		castView.AddChild(&furex.View{
			Height:    12,
			MarginTop: -5,
			Handler: &Text{
				Align:        furex.AlignItemEnd,
				Content:      cast.Name,
				Color:        color.White,
				Shadow:       true,
				ShadowOffset: 1,
				ShadowColor:  color.NRGBA{240, 152, 0, 128},
			}})
	}
	nameCast := &furex.View{
		Width:      500,
		Direction:  furex.Row,
		Justify:    furex.JustifySpaceBetween,
		AlignItems: furex.AlignItemEnd,
	}
	nameCast.AddChild(nameView)
	nameCast.AddChild(castView)
	view.AddChild(nameCast)
	view.AddChild(&furex.View{
		ID:        "bar",
		Width:     500,
		Height:    10,
		MarginTop: 5,
		Handler: &Bar{
			Progress: float64(status.HP) / float64(status.MaxHP),
			FG:       barAtlas.GetNineSlice("red_bar_fg.png"),
			BG:       barAtlas.GetNineSlice("red_bar_bg.png"),
		},
	})
	hpView := &furex.View{
		Width:     500,
		Direction: furex.Row,
		Justify:   furex.JustifySpaceBetween,
	}
	hpView.AddChild(&furex.View{
		MarginTop: 5,
		Height:    13,
		Handler: &Text{
			Content:      formatInt(status.HP) + " / " + formatInt(status.MaxHP),
			Color:        color.NRGBA{252, 183, 190, 255},
			Align:        furex.AlignItemStart,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		},
	})
	hpView.AddChild(&furex.View{
		MarginTop: 5,
		Height:    13,
		Handler: &Text{
			Content:      fmt.Sprintf("%.2f%%", float64(status.HP)/float64(status.MaxHP)*100),
			Color:        color.NRGBA{252, 183, 190, 255},
			Align:        furex.AlignItemEnd,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{0, 0, 0, 128},
		},
	})
	view.AddChild(hpView)
	view.AddChild((&furex.View{
		Handler: furex.NewHandler(furex.HandlerOpts{
			Update: func(v *furex.View) {
				v.SetRight(0)
			},
		}),
	}).AddChild(&furex.View{
		Handler: &BuffList{
			Buffs: status.BuffList,
		},
	}))
	return view
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

func AppendHandler(view *furex.View, f func(v *furex.View)) {
	Original := view.Handler
	view.AddChild(&furex.View{
		Handler: furex.NewHandler(furex.HandlerOpts{
			Update: func(v *furex.View) {
				if Original != nil {
					if u, ok := Original.(furex.Updater); ok {
						u.Update(v)
					}
				}
				f(v)
			},
		}),
	})
}
