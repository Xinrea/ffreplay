package script

import (
	"image/color"
	"log"
	"time"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/system/script/userdefine"
	"github.com/Xinrea/ffreplay/pkg/vector"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/image/math/f64"
)

// ffLoader is the module loader for the "ff" Lua module.
// It registers all FF14-specific API functions available from Lua scripts.
func (sr *ScriptRunner) ffLoader(L *lua.LState) int {
	exports := map[string]lua.LGFunction{
		// Utility
		"sleep": sr.ffSleep,

		// Entity creation
		"create_player": sr.ffCreatePlayer,
		"create_boss":   sr.ffCreateBoss,

		// Camera
		"camera_set_pos":  sr.ffCameraSetPos,
		"camera_get_pos":  sr.ffCameraGetPos,
		"camera_zoom":     sr.ffCameraZoom,
		"camera_get_zoom": sr.ffCameraGetZoom,

		// Map
		"load_map": sr.ffLoadMap,

		// World markers
		"add_waymark":    sr.ffAddWaymark,
		"remove_waymark": sr.ffRemoveWaymark,

		// Telegraphs / AoE visualization
		"draw_circle": sr.ffDrawCircle,
		"draw_rect":   sr.ffDrawRect,

		// Text annotations
		"draw_text": sr.ffDrawText,
	}

	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)

	return 1
}

// --- Utility ---

func (sr *ScriptRunner) ffSleep(L *lua.LState) int {
	ms := L.ToInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)

	return 0
}

// --- Entity Creation ---

func (sr *ScriptRunner) ffCreatePlayer(L *lua.LState) int {
	jobName := L.CheckString(1)
	pos := userdefine.ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))

	rt := role.StringToRole(jobName)
	if rt == -1 {
		L.ArgError(1, "Invalid job name: "+jobName)

		return 0
	}

	p := entry.NewPlayer(sr.ecs, rt, f64.Vec2{pos[0], pos[1]}, nil)

	// Notify callback for UI update (party list, etc.)
	if sr.OnPlayerCreated != nil {
		sr.OnPlayerCreated(p)
	}

	ud := L.NewUserData()
	ud.Value = &userdefine.Player{Entry: p}
	L.SetMetatable(ud, L.GetTypeMetatable(userdefine.PlayerTypeName))
	L.Push(ud)

	return 1
}

func (sr *ScriptRunner) ffCreateBoss(L *lua.LState) int {
	name := L.CheckString(1)
	gameID := int64(L.ToInt(2))
	pos := userdefine.ScriptPosition(float64(L.ToNumber(3)), float64(L.ToNumber(4)))
	ringSize := float64(L.ToNumber(5))

	if ringSize == 0 {
		ringSize = 5 // default boss radius in script units
	}

	ringScale := userdefine.ScriptDistance(ringSize) * 2 / 512

	boss := entry.NewEnemy(
		sr.ecs,
		f64.Vec2{pos[0], pos[1]},
		ringScale,
		gameID,
		gameID, // id = gameID for script-created entities
		name,
		true, // isBoss
		1,    // instanceCount
	)
	for _, inst := range component.Sprite.Get(boss).Instances {
		inst.BTick = 0
		inst.ETick = 1<<62 - 1
	}

	ud := L.NewUserData()
	ud.Value = &userdefine.Boss{Entry: boss}
	L.SetMetatable(ud, L.GetTypeMetatable(userdefine.BossTypeName))
	L.Push(ud)

	return 1
}

// --- Camera ---

func (sr *ScriptRunner) ffCameraSetPos(L *lua.LState) int {
	camera := entry.GetCamera(sr.ecs)
	camera.Position = userdefine.ScriptPosition(float64(L.ToNumber(1)), float64(L.ToNumber(2)))

	return 0
}

func (sr *ScriptRunner) ffCameraGetPos(L *lua.LState) int {
	camera := entry.GetCamera(sr.ecs)
	pos := userdefine.WorldToScriptPosition(camera.Position)

	L.Push(lua.LNumber(pos[0]))
	L.Push(lua.LNumber(pos[1]))

	return 2
}

func (sr *ScriptRunner) ffCameraZoom(L *lua.LState) int {
	level := L.ToInt(1)
	camera := entry.GetCamera(sr.ecs)
	camera.ZoomFactor = level

	return 0
}

func (sr *ScriptRunner) ffCameraGetZoom(L *lua.LState) int {
	camera := entry.GetCamera(sr.ecs)
	L.Push(lua.LNumber(camera.ZoomFactor))

	return 1
}

// --- Map ---

func (sr *ScriptRunner) ffLoadMap(L *lua.LState) int {
	mapID := L.ToInt(1)

	if m, ok := model.MapCache[mapID]; ok {
		config := m.Load()
		current := config.Maps[config.CurrentMap]
		camera := entry.GetCamera(sr.ecs)
		origin := vector.NewVector(current.Offset.X*userdefine.ScriptUnit, current.Offset.Y*userdefine.ScriptUnit)
		userdefine.SetScriptOrigin(origin)
		camera.Position = origin

		mapEntry, _ := component.Map.First(sr.ecs.World)
		if mapEntry != nil {
			component.Map.Set(mapEntry, &model.MapData{Config: config})
		}

		log.Printf("Script: loaded map %d", mapID)

		return 0
	}

	L.ArgError(1, "Map preset not found")

	return 0
}

// --- World Markers ---

func (sr *ScriptRunner) ffAddWaymark(L *lua.LState) int {
	markerType := model.WorldMarkerType(L.ToInt(1))
	pos := userdefine.ScriptPosition(float64(L.ToNumber(2)), float64(L.ToNumber(3)))

	if markerType < model.WorldMarkerA || markerType > model.WorldMarker4 {
		L.ArgError(1, "Invalid marker type")

		return 0
	}

	entry.NewWorldMarker(sr.ecs, markerType, f64.Vec2{pos[0], pos[1]})

	return 0
}

func (sr *ScriptRunner) ffRemoveWaymark(L *lua.LState) int {
	markerType := model.WorldMarkerType(L.ToInt(1))

	for m := range component.WorldMarker.Iter(sr.ecs.World) {
		marker := component.WorldMarker.Get(m)
		if marker.Type == markerType {
			m.Remove()

			return 0
		}
	}

	return 0
}

// --- Telegraphs / AoE Visualization ---

func (sr *ScriptRunner) ffDrawCircle(L *lua.LState) int {
	pos := userdefine.ScriptPosition(float64(L.ToNumber(1)), float64(L.ToNumber(2)))
	radius := userdefine.ScriptDistance(float64(L.ToNumber(3)))
	durationMs := int64(0)
	if L.GetTop() >= 4 {
		durationMs = int64(L.ToInt(4))
	}

	fill := color.NRGBA{235, 140, 52, 128} // orange, semi-transparent
	stroke := color.NRGBA{235, 140, 52, 200}

	td := model.NewTelegraphCircle(
		pos,
		radius,
		fill, stroke,
		durationMs,
	)
	entry.NewTelegraph(sr.ecs, td)

	return 0
}

func (sr *ScriptRunner) ffDrawRect(L *lua.LState) int {
	pos := userdefine.ScriptPosition(float64(L.ToNumber(1)), float64(L.ToNumber(2)))
	width := userdefine.ScriptDistance(float64(L.ToNumber(3)))
	height := userdefine.ScriptDistance(float64(L.ToNumber(4)))
	durationMs := int64(0)
	if L.GetTop() >= 5 {
		durationMs = int64(L.ToInt(5))
	}

	fill := color.NRGBA{235, 140, 52, 128}
	stroke := color.NRGBA{235, 140, 52, 200}

	td := model.NewTelegraphRect(
		pos,
		0, // AnchorMiddle
		width, height,
		fill, stroke,
		durationMs,
	)
	entry.NewTelegraph(sr.ecs, td)

	return 0
}

// --- Text Annotations ---

func (sr *ScriptRunner) ffDrawText(L *lua.LState) int {
	pos := userdefine.ScriptPosition(float64(L.ToNumber(1)), float64(L.ToNumber(2)))
	text := L.CheckString(3)
	durationMs := int64(0)
	if L.GetTop() >= 4 {
		durationMs = int64(L.ToInt(4))
	}

	c := color.NRGBA{255, 255, 255, 255}

	td := model.NewTelegraphText(
		pos,
		text,
		c,
		durationMs,
	)
	entry.NewTelegraph(sr.ecs, td)

	return 0
}
