package ui

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/ebitenui/ebitenui"
)

// BlocksCameraInput reports whether world camera controls should ignore wheel input.
func BlocksCameraInput(global *model.GlobalData) bool {
	if global == nil {
		return false
	}
	return global.UIFocus || global.UIHovered
}

// SyncEUIInputState mirrors focused ebitenui widgets into global.UIFocus.
// UIHovered is set only by panels that explicitly hit-test their own bounds
// (e.g. the property window), not from the shared root container.
func SyncEUIInputState(global *model.GlobalData, ui *ebitenui.UI) {
	if global == nil {
		return
	}
	if ui != nil && ui.HasFocus() {
		global.UIFocus = true
	}
}
