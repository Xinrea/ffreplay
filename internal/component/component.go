package component

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/yohamta/donburi"
)

var Camera = donburi.NewComponentType[model.CameraData]()
var Status = donburi.NewComponentType[model.StatusData]()
var Marker = donburi.NewComponentType[model.MarkerData]()
var Sprite = donburi.NewComponentType[model.SpriteData]()
var Timeline = donburi.NewComponentType[model.TimelineData]()
var Velocity = donburi.NewComponentType[model.VelocityData]()
var Global = donburi.NewComponentType[model.GlobalData]()
var Map = donburi.NewComponentType[model.MapData]()
