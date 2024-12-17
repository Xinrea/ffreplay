package renderer

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) BackgroundRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	ground := component.Sprite.Get(tag.Background.MustFirst(ecs.World))
	ground.Render(camera, screen)
}
