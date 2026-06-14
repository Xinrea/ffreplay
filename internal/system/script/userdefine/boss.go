package userdefine

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	lua "github.com/yuin/gopher-lua"
)

const BossTypeName = "ff_boss"

type Boss struct {
	Entry *donburi.Entry
}

func registerBossType(L *lua.LState) {
	mt := L.NewTypeMetatable(BossTypeName)
	L.SetGlobal("boss", mt)

	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"id":               bossGetID,
		"pos":              bossGetSetPos,
		"set_pos":          bossSetPos,
		"face":             bossSetFace,
		"walk":             bossWalk,
		"apply_buff":       bossApplyBuff,
		"remove_buff":      bossRemoveBuff,
		"set_headmarker":   bossSetHeadMarker,
		"clear_headmarker": bossClearHeadMarker,
	}))
}

func checkBoss(L *lua.LState, pos int) *Boss {
	ud := L.CheckUserData(pos)
	if v, ok := ud.Value.(*Boss); ok {
		return v
	}
	L.ArgError(pos, "boss expected")
	return nil
}

func bossGetID(L *lua.LState) int {
	b := checkBoss(L, 1)
	L.Push(lua.LNumber(component.Status.Get(b.Entry).ID))
	return 1
}

func bossGetSetPos(L *lua.LState) int {
	b := checkBoss(L, 1)
	sprite := component.Sprite.Get(b.Entry)
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
	for _, inst := range sprite.Instances {
		inst.Object.UpdatePosition(newPos)
	}
	return 0
}

func bossSetPos(L *lua.LState) int {
	b := checkBoss(L, 1)
	sprite := component.Sprite.Get(b.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	pos := ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))
	for _, inst := range sprite.Instances {
		inst.Object.UpdatePosition(pos)
	}
	return 0
}

func bossSetFace(L *lua.LState) int {
	b := checkBoss(L, 1)
	sprite := component.Sprite.Get(b.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	angle := float64(L.ToNumber(2))
	for _, inst := range sprite.Instances {
		inst.Face = angle
	}
	return 0
}

// bossWalk starts a walk animation for the boss.
func bossWalk(L *lua.LState) int {
	b := checkBoss(L, 1)
	sprite := component.Sprite.Get(b.Entry)
	if sprite == nil || len(sprite.Instances) == 0 {
		return 0
	}
	pos := ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))
	durationMs := int64(L.ToInt(4))
	if durationMs <= 0 {
		durationMs = 1000
	}
	for _, inst := range sprite.Instances {
		inst.WalkStart = inst.Object.Position()
		inst.WalkTarget = object.NewPointObject(pos)
		inst.WalkTick = -1
		inst.WalkDuration = durationMs * 60 / 1000
	}
	return 0
}

func bossApplyBuff(L *lua.LState) int {
	b := checkBoss(L, 1)
	status := component.Status.Get(b.Entry)
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

func bossRemoveBuff(L *lua.LState) int {
	b := checkBoss(L, 1)
	status := component.Status.Get(b.Entry)
	status.EnsureBuffList().Remove(&model.Buff{ID: int64(L.ToInt(2))})
	return 0
}

func bossSetHeadMarker(L *lua.LState) int {
	b := checkBoss(L, 1)
	component.Status.Get(b.Entry).HeadMarker = L.ToInt(2)
	return 0
}

func bossClearHeadMarker(L *lua.LState) int {
	b := checkBoss(L, 1)
	component.Status.Get(b.Entry).HeadMarker = 0
	return 0
}
