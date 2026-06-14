//go:build !js

package ui

// cutToClipboard is a no-op on non-JS platforms.
func (se *ScriptEditorWindow) cutToClipboard() {
	// No clipboard access on native builds.
	// On native desktop builds, users can use system shortcuts.
}

// copyToClipboard is a no-op on non-JS platforms.
func (se *ScriptEditorWindow) copyToClipboard() {
	// No clipboard access on native builds.
}

// pasteFromClipboard is a no-op on non-JS platforms.
func (se *ScriptEditorWindow) pasteFromClipboard() {
	// No clipboard access on native builds.
}
