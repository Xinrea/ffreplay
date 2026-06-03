package userdefine

import lua "github.com/yuin/gopher-lua"

// RegisterTypes registers all user-defined Lua types (player, boss, etc.)
// with the given Lua state.
func RegisterTypes(L *lua.LState) {
	registerPlayerType(L)
	registerBossType(L)
}
