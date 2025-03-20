package userdefine

import (
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model/role"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/image/math/f64"
)

type Player struct {
	ID    int
	PosX  float64
	PosY  float64
	Job   string
	Class string
}

const luaPlayerTypeName = "player"

// Registers my person type to given L.
func registerPlayerType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPlayerTypeName)
	L.SetGlobal("player", mt)
	// static attributes
	L.SetField(mt, "new", L.NewFunction(newPlayer))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), personMethods))
}

func newPlayer(L *lua.LState) int {
	pos := f64.Vec2{
		float64(L.ToNumber(2)) * 100, float64(L.ToNumber(3)) * 100,
	}
	job := L.CheckString(1)

	rt := role.StringToRole(job)
	if rt == -1 {
		L.ArgError(3, "Invalid job:"+job)

		return 0
	}

	p := entry.NewPlayer(rt, pos, nil)
	status := entry.StatusOf(p)

	player := &Player{
		ID:  int(status.ID),
		Job: job,
	}

	ud := L.NewUserData()
	ud.Value = player
	L.SetMetatable(ud, L.GetTypeMetatable(luaPlayerTypeName))
	L.Push(ud)

	return 1
}

// Checks whether the first lua argument is a *LUserData with *Person and returns this *Person.
func checkPlayer(L *lua.LState) *Player {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Player); ok {
		return v
	}

	L.ArgError(1, "player expected")

	return nil
}

var personMethods = map[string]lua.LGFunction{
	"id": personGetID,
}

func personGetID(L *lua.LState) int {
	p := checkPlayer(L)

	L.Push(lua.LNumber(p.ID))

	return 1
}
