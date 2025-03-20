package script

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/system/script/userdefine"
	lua "github.com/yuin/gopher-lua"
)

type ScriptRunner struct {
	vm *lua.LState
}

func NewScriptRunner() *ScriptRunner {
	vm := lua.NewState()
	vm.PreloadModule("ff", Loader)
	userdefine.RegisterTypes(vm)

	return &ScriptRunner{vm: vm}
}

func (s *ScriptRunner) Run(script string) {
	go s.doRun(script)
}

func (s *ScriptRunner) doRun(script string) {
	defer s.vm.Close()

	if err := s.vm.DoString(script); err != nil {
		log.Println("Running script error:", err)
	}
}
