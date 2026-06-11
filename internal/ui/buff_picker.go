package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// buffPickerEntry is stored inside icon combo entries.
type buffPickerEntry struct {
	Catalog *BuffCatalogEntry
}

func (e buffPickerEntry) Label() string {
	if e.Catalog == nil {
		return "自定义 ID"
	}
	return fmt.Sprintf("%s (%s)", formatPropNumber(float64(e.Catalog.ID), "%.0f"), e.Catalog.Name)
}

// EUIBuffPicker is a reusable buff catalog dropdown.
type EUIBuffPicker struct {
	combo       *EUIIconCombo
	idInput     *widget.TextInput
	manualEntry buffPickerEntry
}

// NewEUIBuffPicker builds a buff picker with catalog dropdown and optional manual ID input.
func NewEUIBuffPicker(st propInspectorStyle, ui *ebitenui.UI) *EUIBuffPicker {
	ensureBuffCatalogFromAssets()
	catalog := BuffCatalog()
	iconSize := int(24 * st.scale)
	if iconSize < 16 {
		iconSize = 16
	}

	items := make([]iconComboItem, 0, len(catalog)+1)
	var initial any
	for i := range catalog {
		entry := buffPickerEntry{Catalog: &catalog[i]}
		items = append(items, iconComboItem{
			Label:  entry.Label(),
			Search: buffPickerSearchKey(entry.Catalog),
			Icon:   buffPickerIcon(entry.Catalog.Icon, iconSize),
			Entry:  entry,
		})
		if initial == nil {
			initial = entry
		}
	}
	manual := buffPickerEntry{Catalog: nil}
	items = append(items, iconComboItem{
		Label:  manual.Label(),
		Search: "自定义 manual id",
		Icon:   buffPickerIcon("", iconSize),
		Entry:  manual,
	})

	picker := &EUIBuffPicker{manualEntry: manual}
	picker.combo = newEUIIconCombo(ui, st, iconPickerConfig{
		Title: "选择 Buff", GridCols: 8, GridRows: 4,
	}, items, initial, func(entry any) {
		if item, ok := entry.(buffPickerEntry); ok && item.Catalog == nil && picker.idInput != nil {
			picker.idInput.Focus(true)
		}
	})

	idW := int(90 * st.scale)
	picker.idInput = propCompactTextInput(st, idW, "")

	return picker
}

func buffPickerSearchKey(catalog *BuffCatalogEntry) string {
	if catalog == nil {
		return ""
	}
	return strings.ToLower(fmt.Sprintf("%s %s %d", catalog.Name, catalog.Icon, catalog.ID))
}

func buffPickerIcon(iconName string, size int) *ebiten.Image {
	var img *ebiten.Image
	if iconName == "" {
		img = texture.NewAbilityTexture("")
	} else {
		img = texture.NewAbilityTexture(iconName)
	}
	if img == nil {
		return nil
	}
	return scaleImage(img, size, size)
}

// Container returns the picker UI: catalog row + manual ID row.
func (p *EUIBuffPicker) Container(st propInspectorStyle) *widget.Container {
	block := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(int(4 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	block.AddChild(propLabeledControlRow(st, "Buff", p.combo.Widget()))

	manualRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	manualRow.AddChild(propFieldLabel(st, "或 ID"))
	manualRow.AddChild(p.idInput)
	block.AddChild(manualRow)

	return block
}

// SelectedBuffID resolves the buff id from the catalog selection or manual input.
func (p *EUIBuffPicker) SelectedBuffID() (int64, bool) {
	if entry := p.combo.SelectedEntry(); entry != nil {
		if item, ok := entry.(buffPickerEntry); ok && item.Catalog != nil {
			return item.Catalog.ID, true
		}
	}
	if p.idInput == nil {
		return 0, false
	}
	id, err := parsePropNumber(p.idInput.GetText())
	if err != nil || id <= 0 {
		return 0, false
	}
	return int64(id), true
}

// ComboFocused reports whether the catalog dropdown owns focus.
func (p *EUIBuffPicker) ComboFocused() bool {
	return p.combo != nil && p.combo.Focused()
}

// ComboOpen reports whether the catalog dropdown list is expanded.
func (p *EUIBuffPicker) ComboOpen() bool {
	return p.combo != nil && p.combo.ComboOpen()
}

// IDInputFocused reports whether the manual ID field owns focus.
func (p *EUIBuffPicker) IDInputFocused() bool {
	return p.idInput != nil && p.idInput.IsFocused()
}

// ParsePositiveInt reads a positive integer from a property text input.
func ParsePositiveInt(ti *widget.TextInput) (int, bool) {
	if ti == nil {
		return 0, false
	}
	v, err := parsePropNumber(ti.GetText())
	if err != nil || v <= 0 {
		return 0, false
	}
	return int(v), true
}

// SetManualID sets the manual buff id field (digits only formatting applied on sync).
func (p *EUIBuffPicker) SetManualID(id int64) {
	if p.idInput != nil {
		p.idInput.SetText(strconv.FormatInt(id, 10))
	}
}

// Close dismisses the floating buff picker panel.
func (p *EUIBuffPicker) Close() {
	if p.combo != nil {
		p.combo.Close()
	}
}
