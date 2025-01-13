package ui

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

func DamageHistoryView() *furex.View {
	global := entry.GetGlobal(ecsInstance)

	view := furex.NewView(
		furex.TagName("damage-history"),
		furex.Direction(furex.Column),
		furex.Height(160),
		furex.MarginLeft(10),
		furex.MarginTop(40),
		furex.Width(300),
	)

	view.Handler.Update = func(v *furex.View) {
		v.RemoveAll()

		if !global.Loaded.Load() || global.TargetPlayer == nil {
			return
		}

		view.AddChild(DamageHistoryHeaderView())

		instance := component.Sprite.Get(global.TargetPlayer).Instances[0]

		damageHistory := instance.GetHistoryDamageTaken(5)
		for _, damage := range damageHistory {
			view.AddChild(DamageHistoryItemView(damage))
		}
	}

	view.Handler.Draw = func(screen *ebiten.Image, frame image.Rectangle, v *furex.View) {
		if global.TargetPlayer == nil {
			return
		}

		bg := texture.NewNineSlice(
			texture.NewTextureFromFile("asset/partylist_bg.png"),
			PartyListBGNineSliceConfig[0],
			PartyListBGNineSliceConfig[1],
			PartyListBGNineSliceConfig[2],
			PartyListBGNineSliceConfig[3])
		frame.Min.X -= 40
		frame.Min.Y -= 20
		frame.Max.Y += 10
		bg.Draw(screen, frame, nil)
	}

	return view
}

func DamageHistoryHeaderView() *furex.View {
	view := furex.NewView(
		furex.TagName("damage-history-header"),
		furex.Direction(furex.Row),
		furex.Height(20),
	)

	view.AddChild(createHeaderView("damage-history-header-time", "时间点", 80))
	view.AddChild(createHeaderView("damage-history-header-name", "伤害名", 100))
	view.AddChild(createHeaderView("damage-history-header-damage", "最终伤害量", 80))
	view.AddChild(createHeaderView("damage-history-header-multiplier", "减伤", 80))

	return view
}

func createHeaderView(tagName, content string, width int) *furex.View {
	return furex.NewView(
		furex.TagName(tagName),
		furex.Height(14),
		furex.Width(width),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      content,
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	)
}

func DamageHistoryItemView(damage model.DamageTaken) *furex.View {
	view := furex.NewView(
		furex.TagName("damage-history-item"),
		furex.Height(28),
		furex.Direction(furex.Row),
		furex.AlignItems(furex.AlignItemCenter),
	)

	view.AddChild(createDamageHistoryItemTimeView(damage))
	view.AddChild(createDamageHistoryItemNameView(damage))
	view.AddChild(createDamageHistoryItemDamageView(damage))
	view.AddChild(createDamageHistoryItemMultiplierView(damage))

	buffs := BuffListView(damage.RelatedBuffs)

	if buffs != nil {
		buffs.Attrs.MarginTop = 10
		view.AddChild(buffs)
	}

	return view
}

func createDamageHistoryItemTimeView(damage model.DamageTaken) *furex.View {
	return furex.NewView(
		furex.TagName("damage-history-item-time"),
		furex.Height(14),
		furex.Width(80),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      formatDuration(float64(util.TickToMS(damage.Tick)) / 1000),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	)
}

func createDamageHistoryItemNameView(damage model.DamageTaken) *furex.View {
	view := furex.NewView(
		furex.TagName("damage-history-item-name"),
		furex.Direction(furex.Row),
	)

	damageIcon := "asset/ui/d_physical.png"

	switch damage.Type {
	case model.Physical:
		damageIcon = "asset/ui/d_physical.png"
	case model.Magical:
		damageIcon = "asset/ui/d_magical.png"
	case model.Special:
		damageIcon = "asset/ui/d_special.png"
	default:
		log.Println("Unknown damage type:", damage.Type)
	}

	view.AddChild(furex.NewView(
		furex.Height(14),
		furex.Width(14),
		furex.Handler(&Sprite{
			Texture: texture.NewTextureFromFile(damageIcon),
		})))

	view.AddChild(furex.NewView(
		furex.Height(14),
		furex.Width(100),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      shortName(damage.Ability.Name, 10),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	))

	return view
}

func createDamageHistoryItemDamageView(damage model.DamageTaken) *furex.View {
	return furex.NewView(
		furex.TagName("damage-history-item-damage"),
		furex.Height(14),
		furex.Width(80),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      fmt.Sprintf("%d", damage.Amount),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	)
}

func createDamageHistoryItemMultiplierView(damage model.DamageTaken) *furex.View {
	return furex.NewView(
		furex.TagName("damage-history-item-multiplier"),
		furex.Height(14),
		furex.Width(40),
		furex.Handler(&Text{
			Align:        furex.AlignItemStart,
			Content:      fmt.Sprintf("%.0f%%", (1-damage.Multiplier)*100),
			Color:        color.White,
			Shadow:       true,
			ShadowOffset: 2,
			ShadowColor:  color.NRGBA{22, 45, 87, 128},
		}),
	)
}

func shortName(name string, limit int) string {
	runes := []rune(name)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}

	return name
}
