package userdefine

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	lua "github.com/yuin/gopher-lua"
)

// PlayerTypeName is the Lua metatable name for player userdata.
const PlayerTypeName = "ff_player"

// Player wraps a donburi Entry for a player in Lua userdata.
type Player struct {
	Entry *donburi.Entry
}

// registerPlayerType registers the player type and its methods with the Lua state.
func registerPlayerType(L *lua.LState) {
	mt := L.NewTypeMetatable(PlayerTypeName)
	L.SetGlobal("player", mt)

	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"id":               playerGetID,
		"pos":              playerGetSetPos,
		"set_pos":          playerSetPos,
		"face":             playerSetFace,
		"walk":             playerWalk,
		"apply_buff":       playerApplyBuff,
		"remove_buff":      playerRemoveBuff,
		"set_headmarker":   playerSetHeadMarker,
		"clear_headmarker": playerClearHeadMarker,
	}))
}

func checkPlayer(L *lua.LState, pos int) *Player {
	ud := L.CheckUserData(pos)
	if v, ok := ud.Value.(*Player); ok {
		return v
	}
	L.ArgError(pos, "player expected")
	return nil
}

func playerGetID(L *lua.LState) int {
	p := checkPlayer(L, 1)
	status := component.Status.Get(p.Entry)
	L.Push(lua.LNumber(status.ID))
	return 1
}

func playerGetSetPos(L *lua.LState) int {
	p := checkPlayer(L, 1)
	sprite := component.Sprite.Get(p.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	pos := sprite.Instances[0].Object.Position()
	if L.GetTop() == 1 {
		scriptPos := WorldToScriptPosition(pos)
		L.Push(lua.LNumber(scriptPos[0]))
		L.Push(lua.LNumber(scriptPos[1]))
		return 2
	}
	newPos := ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))
	sprite.Instances[0].Object.UpdatePosition(newPos)
	return 0
}

func playerSetPos(L *lua.LState) int {
	p := checkPlayer(L, 1)
	sprite := component.Sprite.Get(p.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	pos := ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))
	sprite.Instances[0].Object.UpdatePosition(pos)
	return 0
}

func playerSetFace(L *lua.LState) int {
	p := checkPlayer(L, 1)
	sprite := component.Sprite.Get(p.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	sprite.Instances[0].Face = float64(L.ToNumber(2))
	return 0
}

// playerWalk starts a walk animation to (x, y) over duration_ms.
// Usage: player:walk(x, y, duration_ms)
func playerWalk(L *lua.LState) int {
	p := checkPlayer(L, 1)
	sprite := component.Sprite.Get(p.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}

	pos := ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))
	durationMs := int64(L.ToInt(4))
	if durationMs <= 0 {
		durationMs = 1000
	}

	inst := sprite.Instances[0]
	inst.WalkStart = inst.Object.Position()
	inst.WalkTarget = object.NewPointObject(pos)
	inst.WalkTick = -1 // -1 means "needs initialization on next WalkUpdate"
	inst.WalkDuration = durationMs * 60 / 1000

	return 0
}

// playerApplyBuff applies a buff to the player.
// Usage: player:apply_buff(buff_id, duration_ms, stacks)
func playerApplyBuff(L *lua.LState) int {
	p := checkPlayer(L, 1)
	status := component.Status.Get(p.Entry)

	buffID := int64(L.ToInt(2))
	durationMs := int64(L.ToInt(3))
	stacks := 1
	if L.GetTop() >= 4 {
		stacks = L.ToInt(4)
	}

	info := model.GetBuffInfo(buffID)
	name := "Unknown"
	icon := ""
	if info != nil {
		name = info.Name
		icon = info.Icon
	}

	buff := &model.Buff{
		Type:     model.NormalBuff,
		ID:       buffID,
		Name:     name,
		Icon:     icon,
		Stacks:   stacks,
		Duration: durationMs,
	}
	status.EnsureBuffList().Add(buff)

	return 0
}

// playerRemoveBuff removes a buff from the player.
// Usage: player:remove_buff(buff_id)
func playerRemoveBuff(L *lua.LState) int {
	p := checkPlayer(L, 1)
	status := component.Status.Get(p.Entry)
	buffID := int64(L.ToInt(2))

	status.EnsureBuffList().Remove(&model.Buff{ID: buffID})

	return 0
}

// playerSetHeadMarker sets a head marker on the player.
// Usage: player:set_headmarker(marker_type) -- 1-8
func playerSetHeadMarker(L *lua.LState) int {
	p := checkPlayer(L, 1)
	status := component.Status.Get(p.Entry)
	status.HeadMarker = L.ToInt(2)

	return 0
}

// playerClearHeadMarker removes the head marker from the player.
// Usage: player:clear_headmarker()
func playerClearHeadMarker(L *lua.LState) int {
	p := checkPlayer(L, 1)
	status := component.Status.Get(p.Entry)
	status.HeadMarker = 0

	return 0
}
