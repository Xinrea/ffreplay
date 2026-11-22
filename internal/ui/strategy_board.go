package ui

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/fs"
	"math"
	"slices"
	"strings"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	leftRail  = 204
	rightRail = 248
	topBar    = 60
	bottomBar = 42
	stepRail  = 90
)

// StrategyBoard is a document editor rather than a running encounter. A step
// can therefore be copied, reordered or restored without cloning live ECS entries.
type StrategyBoard struct {
	editor          planner.Editor
	ui              *ebitenui.UI
	width, height   int
	compact         bool
	panel           string
	group, page     int
	palettePage     int
	linkTool        string
	linkStart       *linkPoint
	selectedLink    int // one-based, zero when no link is selected
	selected        map[int]bool
	pending, status string
	armedAt         image.Point
	toolPressed     bool
	rebuild         bool
	drag            *boardDrag
	layerDrag       *layerGesture
	snap            bool
	palette         [][]planner.Tool
	floors          map[string]string // document key -> trusted catalogue asset
	savedRevision   uint64
	hasDraft        bool
	guarded         bool
	fields          map[*widget.TextInput]*boardField
	buffEditor      *boardBuffPanel
	refreshOnIdle   bool
	floorArt        *ebiten.Image
	floorKey        string
	floorSquare     bool
}

type boardDrag struct {
	x, y   float64
	mode   string
	start  map[int]planner.Object
	anchor int
}

func NewStrategyBoard() *StrategyBoard {
	b := &StrategyBoard{
		editor: planner.Editor{Plan: planner.New()}, selected: map[int]bool{},
		ui:      &ebitenui.UI{Container: widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))},
		palette: planner.Palette(), rebuild: true,
		status: "从左侧拖入对象，或先选工具再点击场地。Shift 可多选。",
	}
	b.loadFloors()
	data, err := loadBoard()
	if err == nil && len(data) > 0 {
		var p planner.Plan
		p, err = b.decode(data)
		if err == nil {
			b.editor.Plan = p
			b.hasDraft = !openedSharedBoard()
			b.status = "已打开方案"
		}
	}
	if err != nil {
		b.status = "无法打开方案：" + err.Error()
	}
	return b
}

func (b *StrategyBoard) Layout(w, h int) (int, int) {
	compact := w < 1024
	minWidth := 1024.
	if compact {
		minWidth = 480
	}
	scale := max(1., minWidth/float64(max(1, w)), 700./float64(max(1, h)))
	w, h = int(math.Round(float64(w)*scale)), int(math.Round(float64(h)*scale))
	if w != b.width || h != b.height || compact != b.compact {
		b.width, b.height, b.rebuild = w, h, true
		b.compact = compact
		b.panel = ""
	}
	return w, h
}

func (b *StrategyBoard) canvas() image.Rectangle {
	if b.compact {
		return image.Rect(0, topBar, b.width, b.height-bottomBar-stepRail)
	}
	return image.Rect(leftRail, topBar, b.width-rightRail, b.height-bottomBar-stepRail)
}

func (b *StrategyBoard) transform() (cx, cy, unit float64) {
	r := b.canvas()
	return float64(r.Min.X+r.Max.X) / 2, float64(r.Min.Y+r.Max.Y) / 2, float64(min(r.Dx(), r.Dy())-76) / (2 * planner.Extent)
}

func (b *StrategyBoard) resetSelection() {
	b.linkTool, b.linkStart, b.selectedLink = "", nil, 0
	clear(b.selected)
	b.buffEditor = nil
	b.layerDrag = nil
	b.pending, b.drag, b.page, b.rebuild = "", nil, 0, true
}

func (b *StrategyBoard) selectObject(i int, extend bool) {
	b.linkTool, b.linkStart, b.selectedLink = "", nil, 0
	b.buffEditor = nil
	if !extend {
		clear(b.selected)
	}
	if extend && b.selected[i] {
		delete(b.selected, i)
	} else {
		b.selected[i] = true
	}
	b.pending = ""
	b.page = (len(b.editor.Step().Objects) - 1 - i) / layersPerPage
	b.rebuild = true
}

func (b *StrategyBoard) history(redo bool) {
	if b.editor.History(redo) {
		b.resetSelection()
	}
}

func (b *StrategyBoard) removeSelected() {
	if b.selectedLink > 0 {
		b.editor.Checkpoint()
		s := b.editor.Step()
		i := b.selectedLink - 1
		s.Links = slices.Delete(s.Links, i, i+1)
		b.resetSelection()
		return
	}
	if len(b.selected) == 0 {
		return
	}
	b.editor.Checkpoint()
	s := b.editor.Step()
	for i := len(s.Objects) - 1; i >= 0; i-- {
		if b.selected[i] {
			s.RemoveObject(i)
		}
	}
	b.resetSelection()
}

func (b *StrategyBoard) single() (int, bool) {
	if len(b.selected) == 1 {
		for i := range b.selected {
			return i, true
		}
	}
	return 0, false
}

func (b *StrategyBoard) layer(dir int) {
	i, ok := b.single()
	if ok {
		b.moveLayer(i, i+dir)
	}
}

func (b *StrategyBoard) moveLayer(i, j int) {
	s := b.editor.Step()
	if i == j || j < 0 || j >= len(s.Objects) {
		return
	}
	b.editor.Checkpoint()
	o := s.Objects[i]
	s.Objects = slices.Delete(s.Objects, i, i+1)
	s.Objects = slices.Insert(s.Objects, j, o)
	b.selectObject(j, false)
}

func (b *StrategyBoard) store(action func([]byte) (string, error)) bool {
	data, err := json.Marshal(b.editor.Plan)
	if err == nil && len(data) > planner.MaxFileSize {
		err = fmt.Errorf("方案过大，请减少步骤或对象")
	}
	if err == nil {
		b.status, err = action(data)
	}
	if err != nil {
		b.status = "操作未完成：" + err.Error()
	}
	return err == nil
}

func (b *StrategyBoard) saveDraft() {
	if b.store(saveBoard) {
		b.savedRevision, b.hasDraft = b.editor.Revision(), true
	}
}

func (b *StrategyBoard) switchStep(i int) {
	if i >= 0 && i < len(b.editor.Steps) && i != b.editor.Current {
		b.editor.Current = i
		b.resetSelection()
	}
}

func (b *StrategyBoard) importDrop(files fs.FS) {
	names, err := fs.ReadDir(files, ".")
	if err != nil || len(names) != 1 || names[0].IsDir() {
		b.status = "请一次拖入一个 .ffplan 文件"
		return
	}
	f, err := files.Open(names[0].Name())
	if err == nil {
		defer f.Close()
		var data []byte
		data, err = io.ReadAll(io.LimitReader(f, planner.MaxFileSize+1))
		if err == nil {
			if strings.HasPrefix(strings.TrimSpace(string(data)), "[stgy:") {
				b.importGame(string(data))
				return
			}
			var p planner.Plan
			p, err = b.decode(data)
			if err == nil {
				b.editor.Checkpoint()
				b.editor.Plan, b.editor.Current = p, 0
				b.resetSelection()
				b.status = "已导入 " + names[0].Name() + "；可撤销"
			}
		}
	}
	if err != nil {
		b.status = "未导入：" + err.Error()
	}
}

func (b *StrategyBoard) Update() error {
	if b.width == 0 {
		return nil
	}
	defer b.updateLeaveGuard()
	_, editing := b.ui.GetFocusedWidget().(*widget.TextInput)
	if b.rebuild || (b.refreshOnIdle && !editing && !input.MouseButtonPressed(ebiten.MouseButtonLeft)) {
		if b.rebuild && editing {
			b.finishField(b.ui.GetFocusedWidget().(*widget.TextInput))
		}
		b.buildControls()
		b.rebuild, b.refreshOnIdle = false, false
	}
	// Commit the field before a press can change selection, save, or navigate.
	// Never let a deferred blur callback write through a stale step/object index.
	previous, _ := b.ui.GetFocusedWidget().(*widget.TextInput)
	if previous != nil && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		b.ui.ClearFocus()
		b.buffEditor = nil
		b.rebuild = true
		return nil
	}
	b.ui.Update()
	if previous != nil && !previous.IsFocused() {
		b.finishField(previous)
	}
	if files := ebiten.DroppedFiles(); files != nil {
		b.endDrag(false)
		b.layerDrag = nil
		if input, ok := b.ui.GetFocusedWidget().(*widget.TextInput); ok {
			b.finishField(input)
			b.ui.ClearFocus()
		}
		b.importDrop(files)
		return nil
	}
	if b.layerDrag != nil {
		b.updateLayerGesture()
		return nil
	}
	if _, editing := b.ui.GetFocusedWidget().(*widget.TextInput); editing {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if b.buffEditor != nil {
			b.buffEditor, b.rebuild = nil, true
		} else if b.compact && b.panel != "" {
			b.panel, b.rebuild = "", true
		} else if b.drag != nil {
			b.endDrag(false)
			b.status = "已取消拖动"
		} else {
			b.resetSelection()
		}
		return nil
	}
	if b.drag == nil {
		mod := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyPageUp):
			b.switchStep(b.editor.Current - 1)
			return nil
		case inpututil.IsKeyJustPressed(ebiten.KeyPageDown):
			b.switchStep(b.editor.Current + 1)
			return nil
		case mod && inpututil.IsKeyJustPressed(ebiten.KeyZ):
			b.history(ebiten.IsKeyPressed(ebiten.KeyShift))
			return nil
		case mod && inpututil.IsKeyJustPressed(ebiten.KeyY):
			b.history(true)
			return nil
		case inpututil.IsKeyJustPressed(ebiten.KeyDelete), inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
			b.removeSelected()
			return nil
		}
		if !mod {
			b.nudge()
		}
	}
	b.pointer()
	return nil
}

func (b *StrategyBoard) updateLeaveGuard() {
	dirty := b.unsaved()
	if dirty != b.guarded {
		guardBoardExit(dirty)
		b.guarded = dirty
	}
}

func (b *StrategyBoard) unsaved() bool {
	if b.editor.Revision() != b.savedRevision {
		return true
	}
	if input, ok := b.ui.GetFocusedWidget().(*widget.TextInput); ok {
		f := b.fields[input]
		return f != nil && input.GetText() != f.value
	}
	return false
}

func (b *StrategyBoard) pointer() {
	mx, my := input.CursorPosition()
	cx, cy, unit := b.transform()
	x, y := (float64(mx)-cx)/unit, (float64(my)-cy)/unit
	onCanvas := image.Pt(mx, my).In(b.canvas())
	pressed := input.MouseButtonJustPressed(ebiten.MouseButtonLeft)
	released := input.MouseButtonJustReleased(ebiten.MouseButtonLeft)
	dropTool := b.toolPressed && released && math.Hypot(float64(mx-b.armedAt.X), float64(my-b.armedAt.Y)) > 5
	if released {
		b.toolPressed = false
	}
	s := b.editor.Step()
	if b.compact && b.panel != "" {
		if pressed && onCanvas && !image.Pt(mx, my).In(b.drawerRect()) {
			b.panel, b.rebuild = "", true
		}
		return
	}
	if b.drag != nil {
		if input.MouseButtonPressed(ebiten.MouseButtonLeft) || released {
			b.moveDrag(x, y)
		}
		if !input.MouseButtonPressed(ebiten.MouseButtonLeft) {
			b.endDrag(true)
		}
		return
	}
	if !onCanvas {
		return
	}
	if b.linkTool != "" {
		if pressed && math.Abs(x) <= planner.Extent && math.Abs(y) <= planner.Extent {
			b.placeLink(x, y)
		}
		return
	}
	if b.pending != "" {
		if o, ok := b.placement(x, y); ((pressed && !b.toolPressed) || dropTool) && ok {
			if len(s.Objects) >= planner.MaxObjects {
				b.status = "本步骤的对象已达上限"
				return
			}
			b.editor.Checkpoint()
			s.Objects = append(s.Objects, o)
			b.selectObject(len(s.Objects)-1, false)
			b.status = "已放置 " + o.Label + "；方向键微调，Shift 加大步长"
		}
		return
	}
	if !pressed {
		return
	}
	mode := "move"
	hit := -1
	selected, single := b.single()
	cycle := ebiten.IsKeyPressed(ebiten.KeyAlt)
	if single && !cycle {
		o := s.Objects[selected]
		for _, kind := range boardHandles(o) {
			hx, hy := handlePosition(o, kind, unit)
			if math.Hypot(x-hx, y-hy)*unit <= 9 {
				hit, mode = selected, kind
				break
			}
		}
	}
	if hit < 0 {
		first := len(s.Objects) - 1
		if cycle && single && s.Objects[selected].Hit(x, y) {
			first = selected - 1
		}
		hit = pickLayer(s.Objects, x, y, first)
		if hit < 0 {
			b.resetSelection()
			if i := b.pickLink(x, y, unit); i >= 0 {
				b.selectedLink = i + 1
			}
			return
		}
		if !cycle {
			if i := b.pickLink(x, y, unit); i >= 0 && s.Links[i].Above > hit {
				b.resetSelection()
				b.selectedLink = i + 1
				return
			}
		}
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		if shift || !b.selected[hit] {
			b.selectObject(hit, shift)
		}
		if !b.selected[hit] {
			return
		}
	}
	start := make(map[int]planner.Object, len(b.selected))
	for i := range b.selected {
		start[i] = s.Objects[i]
	}
	b.drag = &boardDrag{x: x, y: y, mode: mode, start: start, anchor: hit}
}

func pickLayer(objects []planner.Object, x, y float64, first int) int {
	for n := range len(objects) {
		i := (first - n + len(objects)) % len(objects)
		if objects[i].Hit(x, y) {
			return i
		}
	}
	return -1
}

func (b *StrategyBoard) placement(x, y float64) (planner.Object, bool) {
	t, ok := planner.Tools[b.pending]
	if !ok || math.Abs(x) > planner.Extent || math.Abs(y) > planner.Extent {
		return planner.Object{}, false
	}
	if b.snap {
		x, y = math.Round(x), math.Round(y)
	}
	return planner.Object{Kind: t.Key, Label: t.Label, Radius: t.Radius, X: x, Y: y}, true
}

func boardHandles(o planner.Object) []string {
	if _, max, _ := o.ShapeRange(); max != 0 {
		return []string{"shape", "rotate", "size"}
	}
	return []string{"rotate", "size"}
}

func handlePosition(o planner.Object, mode string, unit float64) (float64, float64) {
	s, c := math.Sincos(o.Angle * math.Pi / 180)
	if mode == "shape" {
		var dx, dy float64
		switch o.Kind {
		case "cone":
			dx, dy = math.Sincos(o.ShapeValue() * math.Pi / 360)
			dy = -dy
		case "donut":
			dx = -o.ShapeValue()
		case "rect":
			dx, dy = o.ShapeValue(), .6
		}
		return o.X + o.Radius*(dx*c-dy*s), o.Y + o.Radius*(dx*s+dy*c)
	}
	if mode == "rotate" {
		d := o.Radius + 16/unit
		return o.X + d*s, o.Y - d*c
	}
	return o.X + o.Radius*c, o.Y + o.Radius*s
}

func (b *StrategyBoard) moveDrag(x, y float64) {
	d := b.drag
	dx, dy := x-d.x, y-d.y
	if dx == 0 && dy == 0 {
		for i, o := range d.start {
			b.editor.Step().Objects[i] = o
		}
		return
	}
	if d.mode == "move" {
		if b.snap {
			anchor := d.start[d.anchor]
			dx, dy = math.Round(anchor.X+dx)-anchor.X, math.Round(anchor.Y+dy)-anchor.Y
		}
		dx, dy = constrainTranslation(d.start, dx, dy)
	}
	for i, o := range d.start {
		switch d.mode {
		case "rotate":
			o.Angle = math.Round(math.Atan2(x-o.X, o.Y-y) * 180 / math.Pi)
		case "size":
			o.Radius = planner.Clamp(math.Hypot(x-o.X, y-o.Y), .5, planner.Extent)
		case "shape":
			dx, dy := o.Local(x, y)
			lo, hi, _ := o.ShapeRange()
			switch o.Kind {
			case "cone":
				o.Shape = planner.Clamp(math.Abs(math.Atan2(dx, -dy))*360/math.Pi, lo, hi)
			case "donut":
				o.Shape = planner.Clamp(math.Hypot(dx, dy), lo, hi)
			case "rect":
				o.Shape = planner.Clamp(math.Abs(dx), lo, hi)
			}
		default:
			o.X, o.Y = o.X+dx, o.Y+dy
		}
		b.editor.Step().Objects[i] = o
	}
}

// Clamp the translation, not individual objects: a party keeps its formation
// at the edge of the board and off-grid distances survive a snapped move.
func constrainTranslation(objects map[int]planner.Object, dx, dy float64) (float64, float64) {
	for _, o := range objects {
		dx = planner.Clamp(dx, -planner.Extent-o.X, planner.Extent-o.X)
		dy = planner.Clamp(dy, -planner.Extent-o.Y, planner.Extent-o.Y)
	}
	return dx, dy
}

func (b *StrategyBoard) endDrag(commit bool) {
	d := b.drag
	if d == nil {
		return
	}
	objects := b.editor.Step().Objects
	// Swap preview values out so history captures the pre-gesture document.
	changed := false
	for i, original := range d.start {
		final := objects[i]
		changed = changed || original.X != final.X || original.Y != final.Y || original.Angle != final.Angle || original.Radius != final.Radius || original.Shape != final.Shape
		d.start[i], objects[i] = objects[i], original
	}
	if commit && changed {
		b.editor.Checkpoint()
		for i, final := range d.start {
			objects[i] = final
		}
	}
	b.drag, b.rebuild = nil, true
}

func (b *StrategyBoard) nudge() {
	if len(b.selected) == 0 || b.pending != "" {
		return
	}
	distance := .1
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		distance = 1
	}
	var dx, dy float64
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		dx = -distance
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		dx = distance
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		dy = -distance
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		dy = distance
	default:
		return
	}
	start := make(map[int]planner.Object, len(b.selected))
	objects := b.editor.Step().Objects
	for i := range b.selected {
		start[i] = objects[i]
	}
	dx, dy = constrainTranslation(start, dx, dy)
	if dx != 0 || dy != 0 {
		b.editor.Checkpoint()
		for i := range start {
			objects[i].X += dx
			objects[i].Y += dy
		}
		b.rebuild = true
	}
}
