package ui

import (
	"github.com/yohamta/donburi"
)

// AddPlayerToPartyList adds a player entry to the PartyList view in the UI.
// This is called when scripts create players dynamically.
func AddPlayerToPartyList(entry *donburi.Entry) {
	partyLists := root.FilterByTagName("PartyList")
	if len(partyLists) == 0 {
		return
	}

	partyList := partyLists[0]
	partyList.AddChild(NewPlayerItem(entry))
}
