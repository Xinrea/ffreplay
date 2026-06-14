package ui

import (
	"fmt"
	"image/color"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

const propBuffListMaxHeight = 132

// EUIBuffListManager edits a model.BuffList inside inspector panels.
type EUIBuffListManager struct {
	buffList      *model.BuffList
	picker        *EUIBuffPicker
	durationInput *widget.TextInput
	stacksInput   *widget.TextInput
	rows          *widget.Container
	root          *widget.Container
	scale         float64
	lastRowCount  int
}

// NewEUIBuffListManager builds the buff add/remove section for a buff list.
func NewEUIBuffListManager(st propInspectorStyle, ui *ebitenui.UI, buffList *model.BuffList) *EUIBuffListManager {
	m := &EUIBuffListManager{
		buffList: buffList,
		picker:   NewEUIBuffPicker(st, ui),
		scale:    st.scale,
	}

	durW := int(72 * st.scale)
	stackW := int(48 * st.scale)
	m.durationInput = propCompactTextInput(st, durW, "10000")
	m.stacksInput = propCompactTextInput(st, stackW, "1")

	m.rows = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(int(4 * st.scale)),
		)),
	)

	scroll := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: euiimage.NewNineSliceColor(color.NRGBA{22, 24, 34, 255}),
			Mask: euiimage.NewNineSliceColor(color.NRGBA{0, 0, 0, 255}),
		}),
		widget.ScrollContainerOpts.Content(m.rows),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.panelW-st.padding*2, int(float64(propBuffListMaxHeight)*st.scale)),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	addRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	addRow.AddChild(propFieldLabel(st, "时长"))
	addRow.AddChild(m.durationInput)
	addRow.AddChild(propFieldLabel(st, "层数"))
	addRow.AddChild(m.stacksInput)
	addRow.AddChild(propActionButton(st, "添加", m.addSelectedBuff))

	m.root = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	m.root.AddChild(m.picker.Container(st))
	m.root.AddChild(addRow)
	m.root.AddChild(scroll)
	m.Refresh()

	return m
}

// Container returns the root widget tree for this manager.
func (m *EUIBuffListManager) Container() *widget.Container {
	return m.root
}

// ComboOpen reports whether the buff catalog dropdown is expanded.
func (m *EUIBuffListManager) ComboOpen() bool {
	return m.picker != nil && m.picker.ComboOpen()
}

// Close dismisses the floating buff picker panel.
func (m *EUIBuffListManager) Close() {
	if m.picker != nil {
		m.picker.Close()
	}
}

// Focused reports whether any buff manager input is focused.
func (m *EUIBuffListManager) Focused() bool {
	if m.picker.ComboFocused() || m.picker.IDInputFocused() {
		return true
	}
	if m.durationInput != nil && m.durationInput.IsFocused() {
		return true
	}
	if m.stacksInput != nil && m.stacksInput.IsFocused() {
		return true
	}
	return false
}

// RefreshIfChanged rebuilds rows when the buff count changes (e.g. script edits).
func (m *EUIBuffListManager) RefreshIfChanged() {
	if m.buffList == nil {
		return
	}
	count := len(m.buffList.Buffs())
	if count == m.lastRowCount {
		return
	}
	m.Refresh()
}

// Refresh rebuilds the active buff rows from the bound list.
func (m *EUIBuffListManager) Refresh() {
	if m.rows == nil || m.buffList == nil {
		return
	}
	m.rows.RemoveChildren()

	buffs := m.buffList.Buffs()
	m.lastRowCount = len(buffs)
	if len(buffs) == 0 {
		face := newEUIFace(propEUIFontSize * m.scale)
		m.rows.AddChild(widget.NewText(
			widget.TextOpts.Text("（无 Buff）", &face, color.NRGBA{120, 124, 140, 255}),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
		))
		return
	}

	st := newPropInspectorStyle(m.scale)
	for _, buff := range buffs {
		if buff == nil {
			continue
		}
		m.rows.AddChild(m.buffRow(st, buff))
	}
}

func (m *EUIBuffListManager) addSelectedBuff() {
	if m.buffList == nil || ecsInstance == nil {
		return
	}
	buffID, ok := m.picker.SelectedBuffID()
	if !ok {
		return
	}
	duration := int64(10000)
	if v, ok := ParsePositiveInt(m.durationInput); ok {
		duration = int64(v)
	}
	stacks := 1
	if v, ok := ParsePositiveInt(m.stacksInput); ok {
		stacks = v
	}
	applyBuffToList(m.buffList, buffID, duration, stacks, entry.GetTick(ecsInstance))
	InvalidateBuffListSync()
	m.Refresh()
}

func (m *EUIBuffListManager) buffRow(st propInspectorStyle, buff *model.Buff) *widget.Container {
	label := buff.Name
	if label == "" {
		label = "Unknown"
	}
	detail := ""
	if buff.Stacks > 1 {
		detail = fmt.Sprintf(" x%d", buff.Stacks)
	}
	if buff.Duration > 0 {
		detail += fmt.Sprintf(" %ds", buff.Duration/1000)
	}
	text := fmt.Sprintf("%s (%s)%s", formatPropNumber(float64(buff.ID), "%.0f"), label, detail)

	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.panelW-st.padding*2, st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	face := newEUIFace(st.fontSize)
	row.AddChild(widget.NewText(
		widget.TextOpts.Text(text, &face, color.NRGBA{200, 204, 220, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true})),
	))

	removeID := buff.ID
	row.AddChild(propActionButton(st, "×", func() {
		m.buffList.Remove(&model.Buff{ID: removeID})
		InvalidateBuffListSync()
		m.Refresh()
	}))
	return row
}

func applyBuffToList(buffList *model.BuffList, buffID int64, durationMs int64, stacks int, applyTick int64) {
	if buffList == nil {
		return
	}
	name := "Unknown"
	icon := ""
	if info := model.GetBuffInfo(buffID); info != nil {
		if info.Name != "" {
			name = info.Name
		}
		if info.Icon != "" {
			icon = info.Icon
		}
	}
	if entry := BuffCatalogLookup(buffID); entry != nil {
		if name == "Unknown" && entry.Name != "" {
			name = entry.Name
		}
		if icon == "" {
			icon = entry.Icon
		}
	}

	buffList.Add(&model.Buff{
		Type:      model.NormalBuff,
		ID:        buffID,
		Name:      name,
		Icon:      icon,
		Stacks:    stacks,
		Duration:  durationMs,
		ApplyTick: applyTick,
	})
}
