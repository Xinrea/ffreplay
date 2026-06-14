package ui

import (
	"image/color"
	"strings"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/system/script"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	etext "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ScriptEditorWindow is a floating panel for editing and running Lua scripts.
// It opens as an overlay when activated from the sidebar.
type ScriptEditorWindow struct {
	visible bool

	// Text editing state
	text    string
	cursorX int
	cursorY int
	focused bool
	lines   []string

	// Selection: selStart/selEnd are normalized so that start <= end.
	// When hasSelection is false, the selection is empty.
	hasSelection bool
	selStartLine int
	selStartCol  int
	selEndLine   int
	selEndCol    int

	// shiftHeld tracks whether Shift is being held for selection extension.
	shiftHeld bool
	// mouseDrag tracks whether we are in a mouse-drag selection.
	mouseDrag bool

	// Undo stack
	undoStack []string
	undoMax   int

	// Widget tree
	root          widget.Containerer // the shared ebitenui root
	wrapper       *widget.Container
	wrapperInRoot bool

	// Sub-widgets
	textContainer *widget.Container

	// Style
	scale      float64
	builtScale float64 // scale at which the UI was last built
	editorW    int
	editorH    int
}

const (
	scriptEditorDefaultW = 480
	scriptEditorDefaultH = 360
	scriptEditorTitleH   = 28
	scriptEditorBtnH     = 32
	scriptEditorFontSize = 13.0
)

var (
	editorSelColor  = color.NRGBA{60, 80, 140, 180}
	editorTextColor = color.NRGBA{200, 210, 230, 255}
)

// NewScriptEditorWindow creates a new script editor window.
func NewScriptEditorWindow(scale float64) *ScriptEditorWindow {
	if scale <= 0 {
		scale = 1
	}

	se := &ScriptEditorWindow{
		scale:    scale,
		editorW:  int(float64(scriptEditorDefaultW) * scale),
		editorH:  int(float64(scriptEditorDefaultH) * scale),
		undoMax:  50,
		text: `-- 在此编写 Lua 脚本
-- 使用 ff 模块访问游戏 API
-- 例如: ff.add_player("测试", "Scholar", 100, 100)
`,
	}

	se.lines = strings.Split(se.text, "\n")
	se.cursorY = len(se.lines) - 1
	se.cursorX = len(se.lines[se.cursorY])

	return se
}

// IsVisible returns whether the editor window is currently open.
func (se *ScriptEditorWindow) IsVisible() bool {
	return se.visible
}

// Show makes the editor window visible.
func (se *ScriptEditorWindow) Show() {
	if se.visible {
		return
	}
	se.visible = true
	se.focused = true

	if se.wrapper == nil {
		se.buildUI()
	}

	if !se.wrapperInRoot && se.root != nil {
		_ = se.root.AddChild(se.wrapper)
		se.wrapperInRoot = true
	}
}

// Hide closes the editor window.
func (se *ScriptEditorWindow) Hide() {
	if !se.visible {
		return
	}
	se.visible = false
	se.focused = false
	se.clearSelection()

	if se.wrapperInRoot && se.root != nil && se.wrapper != nil {
		se.root.RemoveChild(se.wrapper)
		se.wrapperInRoot = false
	}
}

// Toggle toggles the editor visibility.
func (se *ScriptEditorWindow) Toggle() {
	if se.visible {
		se.Hide()
	} else {
		se.Show()
	}
}

// Focused reports whether the editor has keyboard focus.
func (se *ScriptEditorWindow) Focused() bool {
	return se.visible && se.focused
}

// SetRoot sets the shared ebitenui root container.
func (se *ScriptEditorWindow) SetRoot(root widget.Containerer) {
	se.root = root
}

// buildUI constructs the widget tree for the editor window.
func (se *ScriptEditorWindow) buildUI() {
	s := se.scale
	fontSize := scriptEditorFontSize * s
	btnFace := newEUIFace(fontSize)
	titleFace := newEUIFace(fontSize)
	titleH := int(float64(scriptEditorTitleH) * s)
	btnH := int(float64(scriptEditorBtnH) * s)
	editorAreaH := se.editorH - titleH - btnH - int(8*s)

	// --- Title bar ---
	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			euiimage.NewNineSliceColor(color.NRGBA{40, 45, 70, 255}),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(se.editorW, titleH),
		),
	)
	titleBar.AddChild(widget.NewText(
		widget.TextOpts.Text("Lua 脚本编辑器", &titleFace, color.NRGBA{220, 225, 255, 255}),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding:            &widget.Insets{Left: int(10 * s)},
		})),
	))

	// --- Text editing area ---
	se.textContainer = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			euiimage.NewNineSliceColor(color.NRGBA{20, 22, 35, 255}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    int(6 * s),
				Bottom: int(6 * s),
				Left:   int(8 * s),
				Right:  int(8 * s),
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(se.editorW, editorAreaH),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				se.focused = true
			}),
			widget.WidgetOpts.MouseButtonPressedHandler(func(args *widget.WidgetMouseButtonPressedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft {
					se.focused = true
					se.mouseDrag = true
					se.clearSelection()
					se.setCursorFromClick(args.OffsetY, s)
				}
			}),
			widget.WidgetOpts.CursorMoveHandler(func(args *widget.WidgetCursorMoveEventArgs) {
				if se.mouseDrag && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
					se.extendSelectionToClick(args.OffsetY, s)
				}
			}),
		),
	)

	se.refreshTextWidgets()

	// --- Button bar ---
	buttonBar := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(int(6 * s)),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top:    int(4 * s),
				Bottom: int(4 * s),
				Left:   int(8 * s),
				Right:  int(8 * s),
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(se.editorW, btnH),
		),
	)

	btnImage := &widget.ButtonImage{
		Idle:         euiimage.NewNineSliceColor(color.NRGBA{50, 55, 80, 255}),
		Hover:        euiimage.NewNineSliceColor(color.NRGBA{65, 70, 100, 255}),
		Pressed:      euiimage.NewNineSliceColor(color.NRGBA{40, 44, 65, 255}),
		PressedHover: euiimage.NewNineSliceColor(color.NRGBA{40, 44, 65, 255}),
	}

	// Run button
	buttonBar.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("▶ 运行", &btnFace, &widget.ButtonTextColor{
			Idle:  color.NRGBA{130, 220, 140, 255},
			Hover: color.NRGBA{160, 250, 170, 255},
		}),
		widget.ButtonOpts.Image(btnImage),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(int(8 * s))),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			se.runScript()
		}),
	))

	// Clear button
	buttonBar.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("清空", &btnFace, &widget.ButtonTextColor{
			Idle:  color.NRGBA{200, 200, 200, 255},
			Hover: color.NRGBA{240, 240, 240, 255},
		}),
		widget.ButtonOpts.Image(btnImage),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(int(8 * s))),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			se.pushUndo()
			se.text = ""
			se.lines = []string{""}
			se.cursorX = 0
			se.cursorY = 0
			se.clearSelection()
			se.refreshTextWidgets()
		}),
	))

	// Close button
	buttonBar.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("关闭", &btnFace, &widget.ButtonTextColor{
			Idle:  color.NRGBA{220, 140, 140, 255},
			Hover: color.NRGBA{250, 170, 170, 255},
		}),
		widget.ButtonOpts.Image(btnImage),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(int(8 * s))),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			se.Hide()
		}),
	))

	// --- Main content ---
	content := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
	)
	content.AddChild(titleBar)
	content.AddChild(se.textContainer)
	content.AddChild(buttonBar)

	// --- Wrapper (positioned as floating overlay) ---
	se.wrapper = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			euiimage.NewNineSliceColor(color.NRGBA{30, 32, 48, 250}),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(se.editorW, se.editorH),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	se.wrapper.AddChild(content)
	se.builtScale = se.scale
}

// setCursorFromClick sets the cursor position based on a click Y offset.
func (se *ScriptEditorWindow) setCursorFromClick(offsetY int, s float64) {
	lineH := float64(scriptEditorFontSize+4) * s
	clickY := offsetY - int(6*s)
	if clickY < 0 {
		clickY = 0
	}
	se.cursorY = int(float64(clickY) / lineH)
	if se.cursorY >= len(se.lines) {
		se.cursorY = len(se.lines) - 1
	}
	if se.cursorY < 0 {
		se.cursorY = 0
	}
	se.cursorX = len(se.lines[se.cursorY])
}

// extendSelectionToClick extends the selection to the clicked position.
func (se *ScriptEditorWindow) extendSelectionToClick(offsetY int, s float64) {
	if !se.hasSelection {
		// Start selection from cursor position
		se.selStartLine = se.cursorY
		se.selStartCol = se.cursorX
		se.hasSelection = true
	}

	lineH := float64(scriptEditorFontSize+4) * s
	clickY := offsetY - int(6*s)
	if clickY < 0 {
		clickY = 0
	}
	endLine := int(float64(clickY) / lineH)
	if endLine >= len(se.lines) {
		endLine = len(se.lines) - 1
	}
	if endLine < 0 {
		endLine = 0
	}
	endCol := len(se.lines[endLine])

	se.selEndLine = endLine
	se.selEndCol = endCol
	se.normalizeSelection()
	se.cursorY = endLine
	se.cursorX = endCol
	se.refreshTextWidgets()
}

// clearSelection clears the text selection.
func (se *ScriptEditorWindow) clearSelection() {
	se.hasSelection = false
	se.selStartLine = 0
	se.selStartCol = 0
	se.selEndLine = 0
	se.selEndCol = 0
}

// normalizeSelection ensures start <= end.
func (se *ScriptEditorWindow) normalizeSelection() {
	if se.selStartLine > se.selEndLine ||
		(se.selStartLine == se.selEndLine && se.selStartCol > se.selEndCol) {
		se.selStartLine, se.selEndLine = se.selEndLine, se.selStartLine
		se.selStartCol, se.selEndCol = se.selEndCol, se.selStartCol
	}
}

// setSelectionRange sets the selection from the current cursor position.
// extendToX/extendToY is the other end of the selection.
func (se *ScriptEditorWindow) setSelectionRange(fromLine, fromCol, toLine, toCol int) {
	se.selStartLine = fromLine
	se.selStartCol = fromCol
	se.selEndLine = toLine
	se.selEndCol = toCol
	se.normalizeSelection()
	se.hasSelection = true
}

// getSelectedText returns the currently selected text.
func (se *ScriptEditorWindow) getSelectedText() string {
	if !se.hasSelection {
		return ""
	}
	se.normalizeSelection()

	if se.selStartLine == se.selEndLine {
		if se.selStartLine < len(se.lines) {
			line := se.lines[se.selStartLine]
			if se.selStartCol <= len(line) && se.selEndCol <= len(line) {
				return line[se.selStartCol:se.selEndCol]
			}
		}
		return ""
	}

	var parts []string
	for i := se.selStartLine; i <= se.selEndLine && i < len(se.lines); i++ {
		line := se.lines[i]
		if i == se.selStartLine {
			if se.selStartCol < len(line) {
				parts = append(parts, line[se.selStartCol:])
			}
		} else if i == se.selEndLine {
			if se.selEndCol <= len(line) {
				parts = append(parts, line[:se.selEndCol])
			} else {
				parts = append(parts, line)
			}
		} else {
			parts = append(parts, line)
		}
	}
	return strings.Join(parts, "\n")
}

// deleteSelection removes the selected text.
func (se *ScriptEditorWindow) deleteSelection() {
	if !se.hasSelection {
		return
	}
	se.pushUndo()
	se.normalizeSelection()

	if se.selStartLine == se.selEndLine {
		line := se.lines[se.selStartLine]
		se.lines[se.selStartLine] = line[:se.selStartCol] + line[se.selEndCol:]
		se.cursorY = se.selStartLine
		se.cursorX = se.selStartCol
	} else {
		startLine := se.lines[se.selStartLine]
		endLine := se.lines[se.selEndLine]
		// Join start (up to selStartCol) + end (from selEndCol)
		se.lines[se.selStartLine] = startLine[:se.selStartCol] + endLine[se.selEndCol:]
		// Remove intermediate lines
		se.lines = append(se.lines[:se.selStartLine+1], se.lines[se.selEndLine+1:]...)
		se.cursorY = se.selStartLine
		se.cursorX = se.selStartCol
	}

	se.clearSelection()
	se.refreshTextWidgets()
}

// pushUndo saves the current text state to the undo stack.
func (se *ScriptEditorWindow) pushUndo() {
	se.undoStack = append(se.undoStack, se.text)
	if len(se.undoStack) > se.undoMax {
		se.undoStack = se.undoStack[len(se.undoStack)-se.undoMax:]
	}
}

// undo restores the last saved text state.
func (se *ScriptEditorWindow) undo() {
	if len(se.undoStack) == 0 {
		return
	}
	prev := se.undoStack[len(se.undoStack)-1]
	se.undoStack = se.undoStack[:len(se.undoStack)-1]
	se.text = prev
	se.lines = strings.Split(prev, "\n")
	se.clearSelection()
	se.cursorY = len(se.lines) - 1
	if se.cursorY < 0 {
		se.cursorY = 0
		se.lines = []string{""}
	}
	se.cursorX = len(se.lines[se.cursorY])
	se.refreshTextWidgets()
}

// selectAll selects all text in the editor.
func (se *ScriptEditorWindow) selectAll() {
	se.hasSelection = true
	se.selStartLine = 0
	se.selStartCol = 0
	se.selEndLine = len(se.lines) - 1
	if se.selEndLine >= 0 {
		se.selEndCol = len(se.lines[se.selEndLine])
	}
	se.cursorY = se.selEndLine
	se.cursorX = se.selEndCol
	se.refreshTextWidgets()
}

// clampCursor ensures cursor is within bounds.
func (se *ScriptEditorWindow) clampCursor() {
	if len(se.lines) == 0 {
		se.lines = []string{""}
	}
	if se.cursorY >= len(se.lines) {
		se.cursorY = len(se.lines) - 1
	}
	if se.cursorY < 0 {
		se.cursorY = 0
	}
	if se.cursorX > len(se.lines[se.cursorY]) {
		se.cursorX = len(se.lines[se.cursorY])
	}
	if se.cursorX < 0 {
		se.cursorX = 0
	}
}

// isLineInSelection returns true if the given line index falls within the selection.
func (se *ScriptEditorWindow) isLineInSelection(lineIdx int) bool {
	if !se.hasSelection {
		return false
	}
	se.normalizeSelection()
	return lineIdx >= se.selStartLine && lineIdx <= se.selEndLine
}

// getLineSelectionCols returns the start and end column of the selection on a given line.
// Returns (start, end, fullySelected). fullySelected means the entire line is within the selection.
func (se *ScriptEditorWindow) getLineSelectionCols(lineIdx int) (int, int, bool) {
	se.normalizeSelection()
	if lineIdx < se.selStartLine || lineIdx > se.selEndLine {
		return 0, 0, false
	}
	if lineIdx > se.selStartLine && lineIdx < se.selEndLine {
		return 0, len(se.lines[lineIdx]), true
	}
	if lineIdx == se.selStartLine && lineIdx == se.selEndLine {
		return se.selStartCol, se.selEndCol, false
	}
	if lineIdx == se.selStartLine {
		return se.selStartCol, len(se.lines[lineIdx]), false
	}
	// lineIdx == selEndLine
	return 0, se.selEndCol, false
}

// refreshTextWidgets rebuilds the text display widgets from the current text,
// including selection highlights.
func (se *ScriptEditorWindow) refreshTextWidgets() {
	if se.textContainer == nil {
		return
	}
	se.textContainer.RemoveChildren()

	s := se.scale
	face := newEUIFace(scriptEditorFontSize * s)
	lineH := int(float64(scriptEditorFontSize+4) * s)
	editorInnerW := se.editorW - int(16*s)

	se.clampCursor()

	for i, line := range se.lines {
		se.renderLine(i, line, face, editorInnerW, lineH, s)
	}
}

// renderLine renders a single line of text, with selection highlighting if applicable.
func (se *ScriptEditorWindow) renderLine(lineIdx int, line string, face etext.Face, width, lineH int, s float64) {
	if se.hasSelection && se.isLineInSelection(lineIdx) {
		selStart, selEnd, fullySelected := se.getLineSelectionCols(lineIdx)
		_ = fullySelected

		// Clamp to line bounds
		if selStart < 0 {
			selStart = 0
		}
		if selEnd > len(line) {
			selEnd = len(line)
		}

		// Build a horizontal row with three segments: before, selected, after
		row := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			)),
		)

		// Before selection
		if selStart > 0 {
			beforeText := line[:selStart]
			// Add cursor indicator if needed
			if se.focused && lineIdx == se.cursorY && se.cursorX == selStart {
				beforeText += "|"
			}
			row.AddChild(widget.NewText(
				widget.TextOpts.Text(beforeText, &face, editorTextColor),
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.MinSize(len(beforeText)*int(8*s), lineH),
				),
			))
		}

		// Selected portion (highlighted)
		selText := line[selStart:selEnd]
		if se.focused && lineIdx == se.cursorY && se.cursorX >= selStart && se.cursorX <= selEnd {
			// Insert cursor within selection
			cursorOffset := se.cursorX - selStart
			if cursorOffset <= len(selText) {
				selText = selText[:cursorOffset] + "|" + selText[cursorOffset:]
			}
		}
		if selText != "" || (se.focused && lineIdx == se.cursorY && se.cursorX >= selStart && se.cursorX <= selEnd) {
			selContainer := widget.NewContainer(
				widget.ContainerOpts.BackgroundImage(
					euiimage.NewNineSliceColor(editorSelColor),
				),
			)
			selContainer.AddChild(widget.NewText(
				widget.TextOpts.Text(selText, &face, color.NRGBA{255, 255, 255, 255}),
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			))
			row.AddChild(selContainer)
		}

		// After selection
		if selEnd < len(line) {
			afterText := line[selEnd:]
			if se.focused && lineIdx == se.cursorY && se.cursorX > selEnd {
				cursorOffset := se.cursorX - selEnd
				if cursorOffset <= len(afterText) {
					afterText = afterText[:cursorOffset] + "|" + afterText[cursorOffset:]
				}
			}
			row.AddChild(widget.NewText(
				widget.TextOpts.Text(afterText, &face, editorTextColor),
				widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			))
		}

		se.textContainer.AddChild(row)
	} else {
		// No selection on this line - render normally with cursor
		displayText := line
		if se.focused && lineIdx == se.cursorY {
			if se.cursorX <= len(line) {
				displayText = line[:se.cursorX] + "|" + line[se.cursorX:]
			} else {
				displayText = line + "|"
			}
		}
		txt := widget.NewText(
			widget.TextOpts.Text(displayText, &face, editorTextColor),
			widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(width, lineH),
			),
		)
		se.textContainer.AddChild(txt)
	}
}

// HandleInput processes keyboard input for the script editor.
func (se *ScriptEditorWindow) HandleInput() {
	if !se.visible || !se.focused {
		return
	}

	changed := false

	ctrlHeld := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shiftNow := ebiten.IsKeyPressed(ebiten.KeyShift)

	// --- Ctrl shortcuts (check before character input) ---

	// Ctrl+A: select all
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		se.selectAll()
		return
	}

	// Ctrl+Z: undo
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		se.undo()
		return
	}

	// Ctrl+X: cut
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyX) {
		se.cutToClipboard()
		return
	}

	// Ctrl+C: copy
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		se.copyToClipboard()
		return
	}

	// Ctrl+V: paste
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		se.pasteFromClipboard()
		changed = true
	}

	// --- Text character input (skip when Ctrl is held to avoid stray chars) ---
	if !ctrlHeld {
		chars := ebiten.AppendInputChars(nil)
		for _, ch := range chars {
			// Skip control characters
			if ch < 32 && ch != '\t' {
				continue
			}
			// Replace selection with typed character
			if se.hasSelection {
				se.deleteSelection()
			}
			se.pushUndo()
			if se.cursorY < len(se.lines) {
				line := se.lines[se.cursorY]
				if se.cursorX > len(line) {
					se.cursorX = len(line)
				}
				se.lines[se.cursorY] = line[:se.cursorX] + string(ch) + line[se.cursorX:]
				se.cursorX++
				changed = true
			}
		}
	}

	// --- Handle special keys ---
	handleKey := func(key ebiten.Key) {
		if !inpututil.IsKeyJustPressed(key) {
			return
		}

		switch key {
		case ebiten.KeyEnter:
			if se.hasSelection {
				se.deleteSelection()
			}
			se.pushUndo()
			if se.cursorY < len(se.lines) {
				line := se.lines[se.cursorY]
				rest := ""
				if se.cursorX <= len(line) {
					rest = line[se.cursorX:]
					se.lines[se.cursorY] = line[:se.cursorX]
				} else {
					se.lines[se.cursorY] = line
				}
				se.lines = append(se.lines[:se.cursorY+1], append([]string{rest}, se.lines[se.cursorY+1:]...)...)
				se.cursorY++
				se.cursorX = 0
				changed = true
			}

		case ebiten.KeyBackspace:
			if se.hasSelection {
				se.deleteSelection()
				changed = true
			} else if se.cursorX > 0 && se.cursorY < len(se.lines) {
				se.pushUndo()
				line := se.lines[se.cursorY]
				if se.cursorX <= len(line) {
					se.lines[se.cursorY] = line[:se.cursorX-1] + line[se.cursorX:]
				}
				se.cursorX--
				changed = true
			} else if se.cursorX == 0 && se.cursorY > 0 {
				se.pushUndo()
				prevLine := se.lines[se.cursorY-1]
				currLine := ""
				if se.cursorY < len(se.lines) {
					currLine = se.lines[se.cursorY]
				}
				se.cursorX = len(prevLine)
				se.lines[se.cursorY-1] = prevLine + currLine
				se.lines = append(se.lines[:se.cursorY], se.lines[se.cursorY+1:]...)
				se.cursorY--
				changed = true
			}

		case ebiten.KeyDelete:
			if se.hasSelection {
				se.deleteSelection()
				changed = true
			} else if se.cursorY < len(se.lines) {
				se.pushUndo()
				line := se.lines[se.cursorY]
				if se.cursorX < len(line) {
					se.lines[se.cursorY] = line[:se.cursorX] + line[se.cursorX+1:]
					changed = true
				} else if se.cursorY+1 < len(se.lines) {
					se.lines[se.cursorY] = line + se.lines[se.cursorY+1]
					se.lines = append(se.lines[:se.cursorY+1], se.lines[se.cursorY+2:]...)
					changed = true
				}
			}

		case ebiten.KeyArrowLeft:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			if se.cursorX > 0 {
				se.cursorX--
			} else if se.cursorY > 0 {
				se.cursorY--
				if se.cursorY < len(se.lines) {
					se.cursorX = len(se.lines[se.cursorY])
				}
			}
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyArrowRight:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			if se.cursorY < len(se.lines) && se.cursorX < len(se.lines[se.cursorY]) {
				se.cursorX++
			} else if se.cursorY+1 < len(se.lines) {
				se.cursorY++
				se.cursorX = 0
			}
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyArrowUp:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			if se.cursorY > 0 {
				se.cursorY--
				if se.cursorY < len(se.lines) {
					if se.cursorX > len(se.lines[se.cursorY]) {
						se.cursorX = len(se.lines[se.cursorY])
					}
				}
			}
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyArrowDown:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			if se.cursorY+1 < len(se.lines) {
				se.cursorY++
				if se.cursorY < len(se.lines) {
					if se.cursorX > len(se.lines[se.cursorY]) {
						se.cursorX = len(se.lines[se.cursorY])
					}
				}
			}
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyHome:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			se.cursorX = 0
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyEnd:
			if shiftNow {
				if !se.hasSelection {
					se.selStartLine = se.cursorY
					se.selStartCol = se.cursorX
				}
			} else {
				se.clearSelection()
			}
			if se.cursorY < len(se.lines) {
				se.cursorX = len(se.lines[se.cursorY])
			}
			if shiftNow {
				se.selEndLine = se.cursorY
				se.selEndCol = se.cursorX
				se.normalizeSelection()
				se.hasSelection = true
			}
			changed = true

		case ebiten.KeyTab:
			se.clearSelection()
			se.pushUndo()
			if se.cursorY < len(se.lines) {
				line := se.lines[se.cursorY]
				if se.cursorX > len(line) {
					se.cursorX = len(line)
				}
				se.lines[se.cursorY] = line[:se.cursorX] + "    " + line[se.cursorX:]
				se.cursorX += 4
				changed = true
			}

		case ebiten.KeyEscape:
			se.focused = false
			se.clearSelection()
			changed = true
		}
	}

	allKeys := []ebiten.Key{
		ebiten.KeyEnter, ebiten.KeyBackspace, ebiten.KeyDelete,
		ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyArrowUp, ebiten.KeyArrowDown,
		ebiten.KeyHome, ebiten.KeyEnd, ebiten.KeyTab, ebiten.KeyEscape,
	}
	for _, key := range allKeys {
		handleKey(key)
	}

	// End mouse drag when mouse button is released
	if se.mouseDrag && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		se.mouseDrag = false
	}

	se.shiftHeld = shiftNow

	if changed {
		se.text = strings.Join(se.lines, "\n")
		se.refreshTextWidgets()
	}
}

// runScript executes the current script text.
func (se *ScriptEditorWindow) runScript() {
	runner := script.ActiveRunner
	if runner == nil {
		return
	}

	code := strings.TrimSpace(se.text)
	if code == "" {
		return
	}

	runner.Run(code)
}

// GetText returns the current script text.
func (se *ScriptEditorWindow) GetText() string {
	return se.text
}

// SetText sets the script text content.
func (se *ScriptEditorWindow) SetText(t string) {
	se.text = t
	se.lines = strings.Split(t, "\n")
	se.clearSelection()
	se.cursorY = len(se.lines) - 1
	if se.cursorY < 0 {
		se.cursorY = 0
		se.lines = []string{""}
	}
	se.cursorX = len(se.lines[se.cursorY])
	if se.textContainer != nil {
		se.refreshTextWidgets()
	}
}

// Update processes input and manages the editor state.
func (se *ScriptEditorWindow) Update(scale float64) {
	if !se.visible {
		return
	}

	// Rebuild UI when device scale factor changes
	if scale > 0 && scale != se.builtScale {
		se.scale = scale
		se.editorW = int(float64(scriptEditorDefaultW) * scale)
		se.editorH = int(float64(scriptEditorDefaultH) * scale)
		if se.wrapper != nil && se.root != nil && se.wrapperInRoot {
			se.root.RemoveChild(se.wrapper)
			se.wrapperInRoot = false
		}
		se.wrapper = nil
		se.buildUI()
		if se.root != nil {
			_ = se.root.AddChild(se.wrapper)
			se.wrapperInRoot = true
		}
		se.refreshTextWidgets()
	}

	se.HandleInput()

	// Track focus for global UI state
	if se.focused {
		entry.GetGlobal(ecsInstance).UIFocus = true
	}
}
