//go:build js

package ui

import (
	"syscall/js"
)

// cutToClipboard copies selected text to clipboard then deletes it.
func (se *ScriptEditorWindow) cutToClipboard() {
	sel := se.getSelectedText()
	if sel == "" {
		return
	}
	writeClipboardJS(sel)
	se.deleteSelection()
}

// copyToClipboard copies selected text to clipboard.
func (se *ScriptEditorWindow) copyToClipboard() {
	sel := se.getSelectedText()
	if sel == "" {
		return
	}
	writeClipboardJS(sel)
	se.clearSelection()
	se.refreshTextWidgets()
}

// pasteFromClipboard reads clipboard text and inserts at cursor.
func (se *ScriptEditorWindow) pasteFromClipboard() {
	text := readClipboardJS()
	if text == "" {
		return
	}
	if se.hasSelection {
		se.deleteSelection()
	}
	se.pushUndo()
	se.insertTextAtCursor(text)
	se.refreshTextWidgets()
}

// insertTextAtCursor inserts multi-line text at the current cursor position.
func (se *ScriptEditorWindow) insertTextAtCursor(text string) {
	newLines := stringsSplit(text, "\n")
	if len(newLines) == 0 {
		return
	}

	// Insert first line at cursor position
	currentLine := se.lines[se.cursorY]
	before := currentLine[:se.cursorX]
	after := currentLine[se.cursorX:]
	se.lines[se.cursorY] = before + newLines[0]

	if len(newLines) == 1 {
		// Single line paste
		se.lines[se.cursorY] = se.lines[se.cursorY] + after
		se.cursorX = len(before) + len(newLines[0])
	} else {
		// Multi-line paste: insert middle lines, last line gets the rest
		se.lines = append(se.lines[:se.cursorY+1],
			append(newLines[1:len(newLines)-1], append([]string{newLines[len(newLines)-1] + after}, se.lines[se.cursorY+1:]...)...)...)
		se.cursorY += len(newLines) - 1
		se.cursorX = len(newLines[len(newLines)-1])
	}
}

// stringsSplit is a local copy of strings.Split to avoid import issues.
func stringsSplit(s, sep string) []string {
	result := []string{}
	for {
		idx := stringsIndex(s, sep)
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func stringsIndex(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// writeClipboardJS writes text to the browser clipboard using the Clipboard API.
func writeClipboardJS(text string) {
	js.Global().Get("navigator").Get("clipboard").Call("writeText", text)
}

// readClipboardJS reads text from the browser clipboard.
// Uses a hidden input element + paste event approach for reliability.
func readClipboardJS() string {
	doc := js.Global().Get("document")

	// Try to use the clipboard paste buffer if already set up
	pasteEl := doc.Call("getElementById", "__ffreplay_paste_buffer__")
	if pasteEl.IsNull() {
		// Create hidden paste buffer element
		textarea := doc.Call("createElement", "textarea")
		textarea.Set("id", "__ffreplay_paste_buffer__")
		textarea.Get("style").Set("position", "fixed")
		textarea.Get("style").Set("top", "-9999px")
		textarea.Get("style").Set("left", "-9999px")
		textarea.Get("style").Set("opacity", "0")
		doc.Get("body").Call("appendChild", textarea)

		// Store last paste value
		lastPaste := ""
		textarea.Set("value", "")
		pasteCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// Read paste data from clipboard
			clipData := args[0].Get("clipboardData")
			if !clipData.IsUndefined() && !clipData.IsNull() {
				pasted := clipData.Call("getData", "text/plain").String()
				if pasted != "" {
					textarea.Set("value", pasted)
				}
			}
			return nil
		})
		textarea.Call("addEventListener", "paste", pasteCallback)
		_ = lastPaste
		pasteEl = textarea
	}

	// Focus and trigger paste
	pasteEl.Call("focus")
	pasteEl.Set("value", "")

	// Read whatever is in the buffer after paste
	result := pasteEl.Get("value").String()
	pasteEl.Set("value", "")
	return result
}
