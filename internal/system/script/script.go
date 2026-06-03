package script

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/system/script/userdefine"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	lua "github.com/yuin/gopher-lua"
)

// ActiveRunner is the currently active ScriptRunner. It is set by the scene
// when entering script-capable modes and can be accessed by the UI layer.
var ActiveRunner *ScriptRunner

// ScriptRunner manages the Lua scripting engine for the playground mode.
// It wraps a gopher-lua LState and provides access to the ECS world for
// the Lua API bindings.
type ScriptRunner struct {
	vm              *lua.LState
	ecs             *ecs.ECS
	OnPlayerCreated func(entry *donburi.Entry)
}

// NewScriptRunner creates a new ScriptRunner with the given ECS.
// It initializes the Lua VM, registers the "ff" module and all user-defined types.
// Also sets itself as the ActiveRunner.
func NewScriptRunner(ecs *ecs.ECS) *ScriptRunner {
	vm := lua.NewState()
	sr := &ScriptRunner{vm: vm, ecs: ecs}
	userdefine.SetScriptOrigin(vector.NewVector(0, 0))

	// Register the "ff" module with FF14-specific API bindings
	vm.PreloadModule("ff", sr.ffLoader)

	// Register user-defined Lua types (player, boss, etc.)
	userdefine.RegisterTypes(vm)

	// Set as active runner
	ActiveRunner = sr

	return sr
}

// ECS returns the ECS world associated with this ScriptRunner.
func (sr *ScriptRunner) ECS() *ecs.ECS {
	return sr.ecs
}

// Run executes a Lua script string in a goroutine.
// Scripts run asynchronously to avoid blocking the game loop.
func (sr *ScriptRunner) Run(script string) {
	go sr.doRun(script)
}

// RunFile loads and executes a Lua script file in a goroutine.
func (sr *ScriptRunner) RunFile(path string) {
	go sr.doRunFile(path)
}

// doRun executes the script in the current goroutine.
func (sr *ScriptRunner) doRun(script string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Script panic:", r)
		}
	}()

	if err := sr.vm.DoString(script); err != nil {
		log.Println("Script error:", err)
	}
}

// doRunFile loads and executes a Lua script file in the current goroutine.
func (sr *ScriptRunner) doRunFile(path string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Script panic:", r)
		}
	}()

	if err := sr.vm.DoFile(path); err != nil {
		log.Println("Script error:", err)
	}
}

// Close releases the Lua VM resources.
func (sr *ScriptRunner) Close() {
	if sr.vm != nil {
		sr.vm.Close()
	}
}
