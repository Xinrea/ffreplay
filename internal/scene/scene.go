package scene

import (
	"github.com/Xinrea/ffreplay/internal/scene/scenes"
	"github.com/hajimehoshi/ebiten/v2"
)

type Scene interface {
	Reset()
	Update()
	Layout(w int, h int)
	Draw(screen *ebiten.Image)
}

type SceneManager struct {
	current Scene
	scenes  map[string]Scene
}

var sceneManager *SceneManager

func NewSceneManager() *SceneManager {
	return &SceneManager{
		scenes: make(map[string]Scene),
	}
}

func (sm *SceneManager) ResetScene() {
	sm.current.Reset()
}

func (sm *SceneManager) SetScene(name string) {
	sm.current = sm.scenes[name]
}

func (sm *SceneManager) AddScene(name string, opt *scenes.FFLogsOpt) {
	var scene Scene = nil

	switch name {
	case "playground":
		scene = scenes.NewPlayGroundScene()
	case "replay":
		scene = scenes.NewFFScene(opt)
	case "falloffaith":
		scene = scenes.NewFallOfFaithScene()
	}

	sm.scenes[name] = scene
	if sm.current == nil {
		sm.current = scene
	}
}

func (sm *SceneManager) Update() {
	sm.current.Update()
}

func (sm *SceneManager) Layout(w, h int) {
	sm.current.Layout(w, h)
}

func (sm *SceneManager) Draw(screen *ebiten.Image) {
	sm.current.Draw(screen)
}

func GetSceneManager() *SceneManager {
	return sceneManager
}
