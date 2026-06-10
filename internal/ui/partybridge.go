package ui

import (
	"github.com/yohamta/donburi"
)

// AddPlayerToPartyList is kept as the script callback hook. The ebitenui
// playground party list reads player entries directly each frame, so no view
// mutation is needed here.
func AddPlayerToPartyList(entry *donburi.Entry) {
}
