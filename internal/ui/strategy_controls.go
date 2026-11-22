package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func boardAt(x, y, w, h int) []widget.WidgetOpt {
	return []widget.WidgetOpt{widget.WidgetOpts.MinSize(w, h),
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{Padding: &widget.Insets{Left: x, Top: y}})}
}

func (b *StrategyBoard) label(x, y int, label string, size float64, ink color.Color) {
	face := newEUIFace(size)
	b.ui.Container.AddChild(widget.NewText(widget.TextOpts.Text(label, &face, ink),
		widget.TextOpts.WidgetOpts(boardAt(x, y, 0, 0)...)))
}

func (b *StrategyBoard) button(x, y, w int, label string, active bool, fn func(), extras ...widget.ButtonOpt) *widget.Button {
	face := newEUIFace(14)
	img, ink := boardButton, boardInk
	if active {
		img, ink = boardActive, boardGold
	}
	opts := []widget.ButtonOpt{widget.ButtonOpts.Text(label, &face, &widget.ButtonTextColor{Idle: ink, Disabled: boardDisabled}),
		widget.ButtonOpts.Image(img), widget.ButtonOpts.WidgetOpts(boardAt(x, y, w, 30)...),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { fn() })}
	btn := widget.NewButton(append(opts, extras...)...)
	b.ui.Container.AddChild(btn)
	return btn
}

type boardField struct {
	value string
	apply func(string) bool
}

func (b *StrategyBoard) finishField(input *widget.TextInput) {
	f := b.fields[input]
	if f == nil || input.GetText() == f.value {
		return
	}
	v := strings.TrimSpace(input.GetText())
	if f.apply(v) {
		f.value = v
		b.refreshOnIdle = true
	}
	input.SetText(f.value)
}

func (b *StrategyBoard) field(x, y, w int, value string, apply func(string) bool) {
	face := newEUIFace(14)
	input := widget.NewTextInput(widget.TextInputOpts.Face(&face), widget.TextInputOpts.Image(boardInput),
		widget.TextInputOpts.Color(boardInputInk), widget.TextInputOpts.Padding(&widget.Insets{Left: 12, Right: 12, Top: 6, Bottom: 6}),
		widget.TextInputOpts.WidgetOpts(boardAt(x, y, w, 30)...), widget.TextInputOpts.SubmitOnEnter(true),
		widget.TextInputOpts.IgnoreEmptySubmit(false),
		widget.TextInputOpts.SubmitHandler(func(a *widget.TextInputChangedEventArgs) {
			b.finishField(a.TextInput)
			b.ui.ClearFocus()
			b.rebuild = true
		}))
	input.SetText(value)
	b.fields[input] = &boardField{value, apply}
	b.ui.Container.AddChild(input)
}

func (b *StrategyBoard) buildControls() {
	b.ui.ClearFocus()
	b.ui.Container.RemoveChildren()
	b.layerDrag = nil
	b.fields = make(map[*widget.TextInput]*boardField)
	b.label(20, 6, "战术板", 23, boardGold)
	b.label(21, 34, "FFReplay", 13, boardMuted)
	if !b.compact {
		b.label(leftRail+17, 8, fmt.Sprintf("当前步骤   %02d / %02d", b.editor.Current+1, len(b.editor.Steps)), 12, boardGold)
		b.label(leftRail+16, 26, shortLabel(b.editor.Step().Name, 18), 20, boardInk)
	} else {
		for i, name := range []string{"工具", "对象", "方案"} {
			b.button(b.width-207+i*65, 15, 60, name, b.panel == name, func() {
				if b.panel == name {
					b.panel = ""
				} else {
					b.panel = name
				}
				b.pending, b.linkTool, b.linkStart, b.rebuild = "", "", nil, true
			})
		}
	}
	b.stepStrip()
	if !b.compact || b.panel == "方案" {
		b.fileActions()
	}
	if !b.compact || b.panel == "工具" {
		b.toolbox()
	}
	if !b.compact || b.panel == "对象" {
		b.buildInspector()
	}
}

func (b *StrategyBoard) drawerRect() image.Rectangle {
	switch b.panel {
	case "对象":
		return image.Rect(b.width-rightRail, topBar, b.width, b.height-bottomBar)
	case "工具":
		return image.Rect(0, topBar, leftRail, b.height-bottomBar)
	case "方案":
		return image.Rect(b.width-388, topBar, b.width, topBar+146)
	default:
		return image.Rectangle{}
	}
}

func (b *StrategyBoard) fileActionsY() int {
	if b.compact {
		return topBar + 12
	}
	return 15
}

func (b *StrategyBoard) fileActions() {
	x, y := b.width-367, b.fileActionsY()
	b.button(x, y, 55, "撤销", false, func() { b.history(false) }).GetWidget().Disabled = !b.editor.CanUndo()
	b.button(x+60, y, 55, "重做", false, func() { b.history(true) }).GetWidget().Disabled = !b.editor.CanRedo()
	b.button(x+125, y, 68, "保存草稿", true, b.saveDraft)
	b.button(x+198, y, 68, "导出文件", false, func() { b.store(exportBoard) })
	b.platformControls(x + 271)
	if b.compact {
		b.button(x, y+48, 76, "新建方案", false, b.newPlan)
		b.label(x+91, y+54, "拖入 .ffplan 打开方案", 14, boardMuted)
		b.gameCodeButtons(x, y+90)
	}
}

func (b *StrategyBoard) toolbox() {
	b.label(16, 80, "场地设置", 18, boardGold)
	b.button(16, 112, 82, "圆形", !b.editor.Square, func() { b.setArena(false) })
	b.button(104, 112, 82, "方形", b.editor.Square, func() { b.setArena(true) })
	b.hudToggle(16, 148, "5 m 网格", b.editor.Grid, func(v bool) { b.editor.Checkpoint(); b.editor.Grid = v; b.rebuild = true })
	b.hudToggle(104, 148, "1 m 吸附", b.snap, func(v bool) { b.snap = v; b.rebuild = true })
	b.floorPicker()
	b.label(16, 227, "放置对象", 18, boardGold)
	for i, name := range append(slices.Clone(planner.PaletteNames), "连线") {
		b.button(16+(i%4)*44, 256+(i/4)*34, 42, name, b.group == i, func() {
			b.group, b.palettePage = i, 0
			b.pending, b.linkTool, b.linkStart = "", "", nil
			b.rebuild = true
		})
	}
	if b.group == len(b.palette) {
		b.linkPalette()
	} else {
		b.buildPalette()
	}
	if b.group == 0 {
		missing := b.editor.Step().MissingParty()
		b.button(16, 428, 170, "补齐八人队伍", false, func() {
			if len(missing) == 0 {
				return
			}
			b.editor.Checkpoint()
			b.editor.Step().Objects = append(b.editor.Step().Objects, missing...)
			b.resetSelection()
			b.panel = ""
			b.status = fmt.Sprintf("已添加 %d 名队员，原有站位未变", len(missing))
		}).GetWidget().Disabled = len(missing) == 0 || len(b.editor.Step().Objects)+len(missing) > planner.MaxObjects
	}
	if !b.compact {
		b.gameCodeButtons(16, b.height-158)
		b.label(16, b.height-112, "方案文件", 15, boardInk)
		b.button(106, b.height-120, 80, "新建方案", false, b.newPlan)
		b.label(16, b.height-84, "拖入 .ffplan 即可打开", 14, boardMuted)
	}
}

func shortLabel(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// A sliding window keeps the active step visible even for a full 64-step plan.
func stepWindow(current, total, slots int) (int, int) {
	slots = min(total, max(1, slots))
	start := min(max(0, current-slots/2), total-slots)
	return start, start + slots
}

func (b *StrategyBoard) stepStrip() {
	if b.compact && (b.panel == "工具" || b.panel == "对象") {
		return
	}
	r := b.canvas()
	x, y := r.Min.X+16, r.Max.Y+6
	b.label(x, y+4, "机制步骤", 17, boardGold)
	end := r.Max.X - 16
	for i, a := range []struct {
		label string
		copy  bool
	}{{"＋ 空白", false}, {"复制当前", true}} {
		btn := b.button(end-218+i*74, y, 70, a.label, false, func() {
			if b.editor.AddStep(a.copy) {
				b.resetSelection()
			}
		})
		btn.GetWidget().Disabled = len(b.editor.Steps) >= planner.MaxSteps
	}
	b.button(end-70, y, 70, "删除步骤", false, func() {
		if b.editor.DeleteStep() {
			b.resetSelection()
		}
	}).GetWidget().Disabled = len(b.editor.Steps) == 1
	y += 40
	current, total := b.editor.Current, len(b.editor.Steps)
	b.button(x, y, 28, "‹", false, func() { b.switchStep(current - 1) }).GetWidget().Disabled = current == 0
	b.button(end-28, y, 28, "›", false, func() { b.switchStep(current + 1) }).GetWidget().Disabled = current == total-1
	width := end - x - 72
	start, stop := stepWindow(current, total, width/128)
	chip := min(166, width/(stop-start))
	for i := start; i < stop; i++ {
		label := fmt.Sprintf("%02d   %s", i+1, shortLabel(b.editor.Steps[i].Name, (chip-40)/14))
		b.button(x+36+(i-start)*chip, y, chip-4, label, i == current, func() { b.switchStep(i) })
	}
}

func (b *StrategyBoard) newPlan() {
	b.editor.Checkpoint()
	b.editor.Plan, b.editor.Current = planner.New(), 0
	b.resetSelection()
	b.panel = ""
	b.status = "已新建方案；可撤销，原草稿尚未覆盖"
}

func (b *StrategyBoard) setArena(square bool) {
	if square != b.editor.Square {
		b.editor.Checkpoint()
		b.editor.Square = square
		b.rebuild = true
	}
}

func (b *StrategyBoard) buildPalette() {
	group := b.palette[b.group]
	cols, w := 4, 38
	if b.group == 2 {
		cols, w = 2, 82
	}
	perPage := cols * 4
	pages := (len(group) + perPage - 1) / perPage
	start := b.palettePage * perPage
	if pages > 1 {
		b.button(16, 507, 38, "‹", false, func() { b.palettePage--; b.rebuild = true }).GetWidget().Disabled = b.palettePage == 0
		b.label(76, 515, fmt.Sprintf("%d / %d", b.palettePage+1, pages), 13, boardMuted)
		b.button(148, 507, 38, "›", false, func() { b.palettePage++; b.rebuild = true }).GetWidget().Disabled = b.palettePage+1 >= pages
	}
	for i, t := range group[start:min(start+perPage, len(group))] {
		x, y := 16+(i%cols)*(w+6), 330+(i/cols)*44
		img := boardSlot
		if t.Key == b.pending {
			img = boardSlotActive
		}
		opts := []widget.ButtonOpt{widget.ButtonOpts.Image(img),
			widget.ButtonOpts.WidgetOpts(boardAt(x, y, w, 38)...),
			widget.ButtonOpts.PressedHandler(func(*widget.ButtonPressedEventArgs) {
				b.linkStart, b.linkTool = nil, ""
				b.pending = t.Key
				b.armedAt = image.Pt(input.CursorPosition())
				b.toolPressed = input.MouseButtonPressed(ebiten.MouseButtonLeft)
				if b.compact {
					b.panel = ""
				}
				b.rebuild = true
			})}
		face := newEUIFace(14)
		opts = append(opts, widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.ToolTip(widget.NewTextToolTip(t.Label, &face, boardInk, hudTooltip))))
		if t.Icon != "" && b.group != 2 {
			opts = append(opts, widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: boardIcon(t.Icon)}))
		} else {
			opts = append(opts, widget.ButtonOpts.TextAndImage(shortLabel(t.Label, 3), &face, &widget.GraphicImage{Idle: toolGlyph(t)}, &widget.ButtonTextColor{Idle: boardInk}))
		}
		b.ui.Container.AddChild(widget.NewButton(opts...))
	}
}

func (b *StrategyBoard) buildInspector() {
	x := b.width - rightRail + 16
	s := b.editor.Step()
	idx, single := b.single()
	if b.buffEditor != nil && single && s.Objects[idx].SupportsBuffs() {
		b.buildBuffPanel(x, idx)
		return
	}
	b.label(x, 76, "步骤名称", 14, boardMuted)
	b.nameField(x, 99, s.Name, func(v string) { s.Name = v })
	if b.selectedLink > 0 {
		b.linkInspector(x)
		return
	}
	propY := b.buildLayers(x)
	b.label(x, propY, "对象属性", 17, boardGold)
	if single && s.Objects[idx].SupportsBuffs() {
		b.button(x+100, propY-7, 116, fmt.Sprintf("头顶 Buff · %d", len(s.Objects[idx].Buffs)), false, func() {
			b.buffEditor = &boardBuffPanel{}
			b.rebuild = true
		})
	}
	if len(b.selected) == 0 {
		b.label(x, propY+37, "点击场地对象或图层即可编辑", 14, boardMuted)
		return
	}
	if i, ok := b.single(); ok {
		b.nameField(x, propY+33, s.Objects[i].Label, func(v string) { s.Objects[i].Label = v })
	} else {
		b.label(x, propY+38, fmt.Sprintf("已选 %d 个对象", len(b.selected)), 15, boardGold)
	}
	fields := []struct {
		label  string
		value  func(*planner.Object) *float64
		lo, hi float64
	}{
		{"X (m)", func(o *planner.Object) *float64 { return &o.X }, -planner.Extent, planner.Extent},
		{"Y (m)", func(o *planner.Object) *float64 { return &o.Y }, -planner.Extent, planner.Extent},
		{"半径 (m)", func(o *planner.Object) *float64 { return &o.Radius }, .5, planner.Extent},
		{"角度 (°)", func(o *planner.Object) *float64 { return &o.Angle }, -360, 360},
	}
	for n, f := range fields {
		col, row := x+(n%2)*112, propY+72+(n/2)*52
		b.label(col, row, f.label, 12, boardMuted)
		value := ""
		first := true
		for i := range b.selected {
			v := fmt.Sprintf("%.2f", *f.value(&s.Objects[i]))
			if !first && v != value {
				value = ""
				break
			}
			value, first = v, false
		}
		b.field(col, row+18, 104, value, func(v string) bool {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil || math.IsNaN(n) || n < f.lo || n > f.hi {
				b.status = fmt.Sprintf("请输入 %g 到 %g 之间的数值", f.lo, f.hi)
				return false
			}
			changed := false
			for i := range b.selected {
				changed = changed || *f.value(&s.Objects[i]) != n
			}
			if !changed {
				return true
			}
			b.editor.Checkpoint()
			for i := range b.selected {
				*f.value(&s.Objects[i]) = n
			}
			return true
		})
	}
	for i, name := range []string{"琥珀", "水蓝", "紫晶", "翠绿"} {
		active := true
		for idx := range b.selected {
			active = active && s.Objects[idx].Color == i
		}
		label := name
		if active {
			label = "✓" + name
		}
		b.button(x+i*55, propY+185, 50, label, active, func() {
			if active {
				return
			}
			b.editor.Checkpoint()
			for idx := range b.selected {
				s.Objects[idx].Color = i
			}
			b.rebuild = true
		}, widget.ButtonOpts.TextColor(&widget.ButtonTextColor{Idle: boardInks[i]}))
	}
	b.label(x, propY+225, "Enter 应用输入，Esc 取消", 13, boardMuted)
}

func (b *StrategyBoard) nameField(x, y int, initial string, set func(string)) {
	b.field(x, y, 216, initial, func(v string) bool {
		if v == "" || len([]rune(v)) > 60 {
			b.status = "名称需要 1 至 60 个字符，已保留原值"
			return false
		}
		b.editor.Checkpoint()
		set(v)
		return true
	})
}
