package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// iconComboItem is one selectable entry in an icon picker panel.
type iconComboItem struct {
	Label  string
	Search string
	Icon   *ebiten.Image
	Entry  any
}

// iconPickerConfig tunes the emoji-style picker grid.
type iconPickerConfig struct {
	Title    string
	GridCols int
	GridRows int
}

func (cfg iconPickerConfig) pageSize() int {
	if cfg.GridCols <= 0 {
		cfg.GridCols = 6
	}
	if cfg.GridRows <= 0 {
		cfg.GridRows = 4
	}
	return cfg.GridCols * cfg.GridRows
}

// EUIIconCombo opens a floating icon grid panel (search + pagination, no scroll).
type EUIIconCombo struct {
	ui           *ebitenui.UI
	root         *widget.Container
	trigger      *widget.Button
	removePicker widget.RemoveWindowFunc
	open         bool
	pendingOpen  bool

	allItems    []iconComboItem
	filtered    []iconComboItem
	page        int
	cfg         iconPickerConfig
	selected    any
	onSelect    func(any)
	st          propInspectorStyle
	iconSize    int
	cellSize    int
	searchInput *widget.TextInput
	gridBody    *widget.Container
	pageText    *widget.Text
	prevBtn     *widget.Button
	nextBtn     *widget.Button
}

func newEUIIconCombo(
	ui *ebitenui.UI,
	st propInspectorStyle,
	cfg iconPickerConfig,
	items []iconComboItem,
	initial any,
	onSelect func(any),
) *EUIIconCombo {
	if cfg.GridCols <= 0 {
		cfg.GridCols = 6
	}
	if cfg.GridRows <= 0 {
		cfg.GridRows = 4
	}
	if cfg.Title == "" {
		cfg.Title = "选择"
	}

	iconSize := int(24 * st.scale)
	if iconSize < 16 {
		iconSize = 16
	}
	cellSize := iconSize + int(12*st.scale)

	c := &EUIIconCombo{
		ui:       ui,
		allItems: items,
		filtered: append([]iconComboItem(nil), items...),
		onSelect: onSelect,
		st:       st,
		cfg:      cfg,
		iconSize: iconSize,
		cellSize: cellSize,
	}

	face := newEUIFace(st.fontSize)
	btnTextColor := &widget.ButtonTextColor{
		Idle:     color.NRGBA{220, 222, 235, 255},
		Disabled: color.NRGBA{120, 122, 140, 255},
	}

	var initialItem *iconComboItem
	for i := range items {
		if items[i].Entry == initial {
			initialItem = &items[i]
			break
		}
	}
	if initialItem == nil && len(items) > 0 {
		initialItem = &items[0]
		c.selected = items[0].Entry
	} else if initialItem != nil {
		c.selected = initial
	}

	triggerLabel := ""
	var triggerIcon *ebiten.Image
	if initialItem != nil {
		triggerLabel = initialItem.Label
		triggerIcon = initialItem.Icon
	}

	c.trigger = widget.NewButton(
		widget.ButtonOpts.Image(propButtonImage()),
		widget.ButtonOpts.TextAndImage(triggerLabel, &face, propIconGraphic(triggerIcon), btnTextColor),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(int(6 * st.scale))),
		widget.ButtonOpts.GraphicPadding(*widget.NewInsetsSimple(int(4 * st.scale))),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(int(120*st.scale), st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.TrackHover(true),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if c.open {
				c.Close()
				return
			}
			// Defer until next UI update so the opening click does not
			// immediately satisfy CLICK_OUT on the new window.
			c.pendingOpen = true
		}),
	)

	c.root = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(int(120*st.scale), st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.OnUpdate(func(w widget.HasWidget) {
				if c.pendingOpen {
					c.pendingOpen = false
					c.openPicker()
				}
			}),
		),
	)
	c.root.AddChild(c.trigger)

	return c
}

func propIconGraphic(img *ebiten.Image) *widget.GraphicImage {
	if img == nil {
		return nil
	}
	return &widget.GraphicImage{Idle: img}
}

func iconPickerToolTip(st propInspectorStyle, label string) *widget.ToolTip {
	face := newEUIFace(st.fontSize * 0.92)
	pad := int(6 * st.scale)
	content := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(color.NRGBA{32, 34, 48, 245})),
		widget.ContainerOpts.Layout(widget.NewRowLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(1, 1),
		),
	)
	content.AddChild(widget.NewText(
		widget.TextOpts.Text(label, &face, color.NRGBA{220, 224, 240, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
			widget.WidgetOpts.MinSize(1, int(float64(st.rowH)*0.85)),
		),
		widget.TextOpts.Padding(&widget.Insets{
			Top: pad, Bottom: pad, Left: pad, Right: pad,
		}),
	))
	return widget.NewToolTip(
		widget.ToolTipOpts.Content(content),
		widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
	)
}

func (c *EUIIconCombo) openPicker() {
	if c.ui == nil || c.trigger == nil {
		return
	}
	c.Close()

	c.page = 0
	c.filtered = append(c.filtered[:0], c.allItems...)

	panel, panelW, panelH := c.buildPickerPanel()
	win := widget.NewWindow(
		widget.WindowOpts.Contents(panel),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),
		widget.WindowOpts.DrawLayer(200),
		widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
			c.open = false
			c.removePicker = nil
		}),
	)
	win.Ephemeral = true
	c.layoutPickerWindow(win, panelW, panelH)

	// AddWindowQuietly(..., false): AddWindow closes ephemeral windows by default,
	// which would remove this picker immediately after adding it.
	c.removePicker = c.ui.AddWindowQuietly(win, false)
	c.open = true
	if c.searchInput != nil {
		c.searchInput.SetText("")
	}
	c.refreshGrid()
}

func (c *EUIIconCombo) buildPickerPanel() (*widget.Container, int, int) {
	pad := int(8 * c.st.scale)
	gap := int(4 * c.st.scale)
	panelW := pad*2 + c.cfg.GridCols*c.cellSize + (c.cfg.GridCols-1)*gap
	gridH := c.cfg.GridRows*c.cellSize + (c.cfg.GridRows-1)*gap

	face := newEUIFace(c.st.fontSize)
	searchW := panelW - pad*2

	c.searchInput = widget.NewTextInput(
		widget.TextInputOpts.Face(&face),
		widget.TextInputOpts.Image(propTextInputImage()),
		widget.TextInputOpts.Color(propTextInputColor()),
		widget.TextInputOpts.Padding(c.st.tiPad),
		widget.TextInputOpts.ChangedHandler(func(args *widget.TextInputChangedEventArgs) {
			c.applySearch(args.InputText)
		}),
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(searchW, c.st.rowH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	c.gridBody = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(c.cfg.GridCols),
			widget.GridLayoutOpts.Spacing(gap, gap),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(c.cfg.GridCols*c.cellSize+(c.cfg.GridCols-1)*gap, gridH),
		),
	)

	pageFace := newEUIFace(c.st.fontSize * 0.95)
	c.pageText = widget.NewText(
		widget.TextOpts.Text("1 / 1", &pageFace, color.NRGBA{170, 175, 195, 255}),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.MinSize(int(60*c.st.scale), c.st.rowH),
		),
	)

	c.prevBtn = propActionButton(c.st, "‹", func() {
		if c.page > 0 {
			c.page--
			c.refreshGrid()
		}
	})
	c.nextBtn = propActionButton(c.st, "›", func() {
		if c.page < c.totalPages()-1 {
			c.page++
			c.refreshGrid()
		}
	})

	footer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(gap),
		)),
	)
	footer.AddChild(c.prevBtn)
	footer.AddChild(c.pageText)
	footer.AddChild(c.nextBtn)

	titleFace := newEUIFace(c.st.fontSize * 1.05)
	title := widget.NewText(
		widget.TextOpts.Text(c.cfg.Title, &titleFace, color.NRGBA{200, 204, 220, 255}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(searchW, int(20*c.st.scale)),
		),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(color.NRGBA{24, 26, 38, 250})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top: pad, Bottom: pad, Left: pad, Right: pad,
			}),
			widget.RowLayoutOpts.Spacing(gap),
		)),
	)
	panel.AddChild(title)
	panel.AddChild(c.searchInput)
	panel.AddChild(c.gridBody)
	panel.AddChild(footer)

	panelH := pad*2 + int(20*c.st.scale) + gap + c.st.rowH + gap + gridH + gap + c.st.rowH
	return panel, panelW, panelH
}

func (c *EUIIconCombo) layoutPickerWindow(win *widget.Window, panelW, panelH int) {
	trig := c.trigger.GetWidget().Rect
	sw, sh := ebiten.WindowSize()

	x := trig.Min.X
	y := trig.Max.Y + int(4*c.st.scale)
	if x+panelW > sw {
		x = sw - panelW - int(8*c.st.scale)
	}
	if y+panelH > sh {
		y = trig.Min.Y - panelH - int(4*c.st.scale)
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	win.SetLocation(image.Rect(x, y, x+panelW, y+panelH))
}

func (c *EUIIconCombo) applySearch(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	c.filtered = c.filtered[:0]
	if q == "" {
		c.filtered = append(c.filtered, c.allItems...)
	} else {
		for _, item := range c.allItems {
			if strings.Contains(item.Search, q) {
				c.filtered = append(c.filtered, item)
			}
		}
	}
	c.page = 0
	c.refreshGrid()
}

func (c *EUIIconCombo) totalPages() int {
	n := len(c.filtered)
	ps := c.cfg.pageSize()
	if n == 0 {
		return 1
	}
	return (n + ps - 1) / ps
}

func (c *EUIIconCombo) refreshGrid() {
	if c.gridBody == nil {
		return
	}
	c.gridBody.RemoveChildren()

	totalPages := c.totalPages()
	if c.page >= totalPages {
		c.page = totalPages - 1
	}
	if c.page < 0 {
		c.page = 0
	}

	if c.pageText != nil {
		if len(c.filtered) == 0 {
			c.pageText.Label = "无结果"
		} else {
			c.pageText.Label = fmt.Sprintf("%d / %d", c.page+1, totalPages)
		}
	}
	if c.prevBtn != nil {
		c.prevBtn.GetWidget().Disabled = c.page <= 0 || len(c.filtered) == 0
	}
	if c.nextBtn != nil {
		c.nextBtn.GetWidget().Disabled = c.page >= totalPages-1 || len(c.filtered) == 0
	}

	ps := c.cfg.pageSize()
	start := c.page * ps
	end := start + ps
	if start > len(c.filtered) {
		start = len(c.filtered)
	}
	if end > len(c.filtered) {
		end = len(c.filtered)
	}
	pageItems := c.filtered[start:end]

	// Pad grid to fixed rows*cols so layout stays stable.
	for len(pageItems) < ps {
		pageItems = append(pageItems, iconComboItem{})
	}

	for _, item := range pageItems {
		if item.Entry == nil {
			spacer := widget.NewContainer(
				widget.ContainerOpts.WidgetOpts(
					widget.WidgetOpts.MinSize(c.cellSize, c.cellSize),
				),
			)
			c.gridBody.AddChild(spacer)
			continue
		}

		entry := item.Entry
		btn := widget.NewButton(
			widget.ButtonOpts.Image(propButtonImage()),
			widget.ButtonOpts.Graphic(propIconGraphic(item.Icon)),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(c.cellSize, c.cellSize),
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
				widget.WidgetOpts.ToolTip(iconPickerToolTip(c.st, item.Label)),
			),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				c.selectEntry(entry)
				if c.onSelect != nil {
					c.onSelect(entry)
				}
				c.Close()
			}),
		)
		c.gridBody.AddChild(btn)
	}
}

func (c *EUIIconCombo) selectEntry(entry any) {
	c.selected = entry
	for i := range c.allItems {
		if c.allItems[i].Entry == entry {
			c.trigger.SetText(c.allItems[i].Label)
			c.trigger.SetGraphicImage(propIconGraphic(c.allItems[i].Icon))
			return
		}
	}
}

// Close dismisses the floating picker panel.
func (c *EUIIconCombo) Close() {
	c.pendingOpen = false
	if c.removePicker != nil {
		c.removePicker()
		c.removePicker = nil
	}
	c.open = false
}

// Widget returns the trigger row container.
func (c *EUIIconCombo) Widget() *widget.Container {
	return c.root
}

// SelectedEntry returns the currently selected entry value.
func (c *EUIIconCombo) SelectedEntry() any {
	return c.selected
}

// ComboOpen reports whether the picker panel is visible.
func (c *EUIIconCombo) ComboOpen() bool {
	return c.open
}

// Focused reports whether the trigger or picker search field owns focus.
func (c *EUIIconCombo) Focused() bool {
	if c.trigger != nil && c.trigger.IsFocused() {
		return true
	}
	if c.open && c.searchInput != nil && c.searchInput.IsFocused() {
		return true
	}
	return false
}
