package script

import (
	"time"

	"github.com/Xinrea/ffreplay/internal/entry"
	lua "github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	// register functions to the table
	mod := L.SetFuncs(L.NewTable(), exports)

	// returns the module
	L.Push(mod)

	return 1
}

var exports = map[string]lua.LGFunction{
	"party": party,
	"sleep": sleep,
}

// party returns the number of party members.
func party(L *lua.LState) int {
	L.Push(lua.LNumber(len(entry.GetPlayerList())))

	return 1
}

func sleep(L *lua.LState) int {
	lv := L.ToInt(1)
	time.Sleep(time.Duration(lv) * time.Millisecond)

	return 0
}
