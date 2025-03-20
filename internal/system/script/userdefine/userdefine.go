package userdefine

import lua "github.com/yuin/gopher-lua"

func RegisterTypes(L *lua.LState) {
	registerPlayerType(L)
}
