package ui

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/yohamta/donburi"
)

// objectListItem represents one entry in the scene objects list.
type objectListItem struct {
	entry     *donburi.Entry
	label     string
	iconLabel string // single-char prefix: P(Player), E(Enemy), M(Marker)
}

// buildSceneObjectsSection returns a collapsible section listing all game
// objects currently in the world. Clicking an item selects it.
func buildSceneObjectsSection(st propInspectorStyle, collapseState map[string]bool) (*widget.Container, *widget.Container) {
	section, body := newPropSection("场景对象", st, collapseState)

	items := collectSceneObjects()
	if len(items) == 0 {
		face := newEUIFace(st.fontSize * 0.95)
		body.AddChild(widget.NewText(
			widget.TextOpts.Text("（暂无对象）", &face, color.NRGBA{120, 124, 140, 255}),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		))
		return section, body
	}

	for _, it := range items {
		row := buildObjectRow(st, it)
		body.AddChild(row)
	}

	return section, body
}

// collectSceneObjects gathers all selectable game objects from the ECS world.
func collectSceneObjects() []objectListItem {
	var items []objectListItem

	// Players
	for e := range tag.Player.Iter(ecsInstance.World) {
		if !e.Valid() {
			continue
		}
		status := component.Status.Get(e)
		label := "玩家"
		if status != nil && status.Name != "" {
			label = status.Name
		}
		items = append(items, objectListItem{
			entry:     e,
			label:     label,
			iconLabel: "P",
		})
	}

	// Enemies
	for e := range tag.Enemy.Iter(ecsInstance.World) {
		if !e.Valid() {
			continue
		}
		status := component.Status.Get(e)
		label := "敌人"
		if status != nil && status.Name != "" {
			label = status.Name
		}
		items = append(items, objectListItem{
			entry:     e,
			label:     label,
			iconLabel: "E",
		})
	}

	// WorldMarkers
	for e := range tag.WorldMarker.Iter(ecsInstance.World) {
		if !e.Valid() {
			continue
		}
		marker := component.WorldMarker.Get(e)
		label := markerTypeName(marker.Type)
		items = append(items, objectListItem{
			entry:     e,
			label:     fmt.Sprintf("%s (%.0f, %.0f)", label, marker.Position[0], marker.Position[1]),
			iconLabel: "M",
		})
	}

	// Sort: Players first, then Enemies, then Markers
	sort.Slice(items, func(i, j int) bool {
		order := map[byte]int{'P': 0, 'E': 1, 'M': 2}
		oi := order[items[i].iconLabel[0]]
		oj := order[items[j].iconLabel[0]]
		if oi != oj {
			return oi < oj
		}
		return items[i].label < items[j].label
	})

	return items
}

// buildObjectRow creates a single clickable row for an object list item.
func buildObjectRow(st propInspectorStyle, it objectListItem) *widget.Container {
	iconColors := map[byte]color.NRGBA{
		'P': {100, 200, 255, 255}, // blue for players
		'E': {255, 140, 100, 255}, // orange for enemies
		'M': {180, 220, 100, 255}, // green for markers
	}

	clr := iconColors[it.iconLabel[0]]
	if clr == (color.NRGBA{}) {
		clr = color.NRGBA{180, 180, 180, 255}
	}

	entryRef := it.entry // capture

	rowH := int(float64(propEUIRowHeight) * st.scale)
	rowFace := newEUIFace(st.fontSize * 0.9)

	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(4 * st.scale)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(st.panelW-st.padding*2, rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	// Icon label
	row.AddChild(widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("[%s]", it.iconLabel), &rowFace, clr),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(int(20*st.scale), rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{MaxWidth: int(24 * st.scale)}),
		),
	))

	// Object name (clickable to select)
	selectBtn := widget.NewButton(
		widget.ButtonOpts.Text(it.label, &rowFace, &widget.ButtonTextColor{
			Idle:     color.NRGBA{200, 204, 220, 255},
			Hover:    color.NRGBA{255, 255, 255, 255},
			Disabled: color.NRGBA{120, 124, 140, 255},
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:         euiimage.NewNineSliceColor(color.NRGBA{0, 0, 0, 0}),
			Hover:        euiimage.NewNineSliceColor(color.NRGBA{60, 65, 100, 180}),
			Pressed:      euiimage.NewNineSliceColor(color.NRGBA{45, 50, 80, 200}),
			PressedHover: euiimage.NewNineSliceColor(color.NRGBA{45, 50, 80, 200}),
		}),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(int(4 * st.scale))),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if entryRef.Valid() {
				g := entry.GetGlobal(ecsInstance)
				g.Selected = entryRef
				g.SelectedInstance = 0
			}
		}),
	)
	row.AddChild(selectBtn)

	return row
}
