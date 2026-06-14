package ui

import (
	"strings"

	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// jobPickerEntry is stored inside icon combo entries.
type jobPickerEntry struct {
	Option role.JobOption
}

// EUIJobPicker is a reusable job dropdown for editor panels.
type EUIJobPicker struct {
	root  *widget.Container
	combo *EUIIconCombo
}

// NewEUIJobPicker builds a labeled job dropdown for editor panels.
func NewEUIJobPicker(st propInspectorStyle, ui *ebitenui.UI, current role.RoleType, onChange func(role.RoleType)) *EUIJobPicker {
	options := role.PlayableJobOptions()
	iconSize := int(24 * st.scale)
	if iconSize < 16 {
		iconSize = 16
	}

	items := make([]iconComboItem, 0, len(options))
	var initial any
	for _, opt := range options {
		entry := jobPickerEntry{Option: opt}
		items = append(items, iconComboItem{
			Label:  opt.Label,
			Search: strings.ToLower(opt.Label),
			Icon:   rolePickerIcon(opt.Role, iconSize),
			Entry:  entry,
		})
		if opt.Role == current {
			initial = entry
		}
	}
	if initial == nil && len(items) > 0 {
		initial = items[0].Entry
	}

	picker := &EUIJobPicker{}
	picker.combo = newEUIIconCombo(ui, st, iconPickerConfig{
		Title: "选择职业", GridCols: 6, GridRows: 4,
	}, items, initial, func(entry any) {
		if item, ok := entry.(jobPickerEntry); ok && onChange != nil {
			onChange(item.Option.Role)
		}
	})
	picker.root = propLabeledControlRow(st, "职业", picker.combo.Widget())
	return picker
}

func rolePickerIcon(r role.RoleType, size int) *ebiten.Image {
	img := texture.NewTextureFromFile("asset/role/" + r.String() + ".png")
	return scaleImage(img, size, size)
}

// Container returns the labeled row widget.
func (p *EUIJobPicker) Container() *widget.Container {
	return p.root
}

// ComboOpen reports whether the dropdown list is expanded.
func (p *EUIJobPicker) ComboOpen() bool {
	return p.combo != nil && p.combo.ComboOpen()
}

// Focused reports whether the picker owns keyboard focus.
func (p *EUIJobPicker) Focused() bool {
	return p.combo != nil && p.combo.Focused()
}

// Close dismisses the floating picker panel.
func (p *EUIJobPicker) Close() {
	if p.combo != nil {
		p.combo.Close()
	}
}
