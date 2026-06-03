package userdefine

import (
	"sync"

	"github.com/Xinrea/ffreplay/pkg/vector"
)

const ScriptUnit = 25

var scriptCoords = struct {
	sync.RWMutex
	origin vector.Vector
}{
	origin: vector.NewVector(0, 0),
}

func SetScriptOrigin(origin vector.Vector) {
	scriptCoords.Lock()
	defer scriptCoords.Unlock()

	scriptCoords.origin = origin
}

func ScriptPosition(x, y float64) vector.Vector {
	scriptCoords.RLock()
	defer scriptCoords.RUnlock()

	return scriptCoords.origin.Add(vector.NewVector(x*ScriptUnit, y*ScriptUnit))
}

func WorldToScriptPosition(pos vector.Vector) vector.Vector {
	scriptCoords.RLock()
	defer scriptCoords.RUnlock()

	return pos.Sub(scriptCoords.origin).Scale(1.0 / ScriptUnit)
}

func ScriptDistance(v float64) float64 {
	return v * ScriptUnit
}
