package component

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/donburi"
)

var (
	Camera      = donburi.NewComponentType[model.CameraData]()
	Status      = donburi.NewComponentType[model.StatusData]()
	WorldMarker = donburi.NewComponentType[model.WorldMarkerData]()
	Sprite      = donburi.NewComponentType[model.SpriteData]()
	Timeline    = donburi.NewComponentType[model.TimelineData]()
	Velocity    = donburi.NewComponentType[model.VelocityData]()
	Global      = donburi.NewComponentType[model.GlobalData]()
	Map         = donburi.NewComponentType[model.MapData]()
)
