package entry

import (
	"fmt"
	"log"
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/layer"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

var ctx *ecs.ECS

func SetContext(ecs *ecs.ECS) {
	ctx = ecs
}

var (
	Player = newArchetype(
		tag.GameObject,
		tag.Player,
		tag.PartyMember,
		tag.Buffable,
		component.Velocity,
		component.Sprite,
		component.Status,
	)
	Pet = newArchetype(tag.GameObject,
		tag.Pet,
		tag.PartyMember,
		tag.Buffable,
		component.Velocity,
		component.Sprite,
		component.Status,
	)
	LimitBreak = newArchetype(tag.GameObject,
		tag.LimitBreak,
		component.Sprite,
		component.Status,
	)
	Enemy = newArchetype(
		tag.GameObject,
		tag.Enemy,
		tag.Buffable,
		component.Velocity,
		component.Sprite,
		component.Status,
	)
	Background  = newArchetype(tag.Background, component.Map)
	Camera      = newArchetype(tag.Camera, component.Camera)
	Timeline    = newArchetype(tag.Timeline, component.Timeline)
	WorldMarker = newArchetype(tag.WorldMarker, component.WorldMarker)
	Global      = newArchetype(tag.Global, component.Global)
)

type archetype struct {
	components []donburi.IComponentType
}

func newArchetype(cs ...donburi.IComponentType) *archetype {
	return &archetype{
		components: cs,
	}
}

func (a *archetype) Spawn(cs ...donburi.IComponentType) *donburi.Entry {
	e := ctx.World.Entry(ctx.Create(
		layer.Default,
		append(a.components, cs...)...,
	))

	return e
}

// boss gameID is unique in ffxiv, id is used in events.
func NewEnemy(
	pos f64.Vec2,
	ringSize float64,
	gameID int64,
	id int64,
	name string,
	isBoss bool,
	instanceCount int,
) *donburi.Entry {
	enemy := Enemy.Spawn()
	textureRing := texture.NewTextureFromFile("asset/target_enemy.png")

	erole := role.Boss
	if !isBoss {
		erole = role.NPC
	}

	status := &model.StatusData{
		GameID:   gameID,
		ID:       id,
		Name:     name,
		Role:     erole,
		HP:       1,
		MaxHP:    1,
		Mana:     10000,
		MaxMana:  10000,
		BuffList: model.NewBuffList(),
		RingConfig: model.RingConfiguration{
			Texture: textureRing,
			Scale:   ringSize,
		},
	}

	instances := []*model.Instance{}
	for i := 0; i < instanceCount; i++ {
		instances = append(instances, &model.Instance{
			Status:     status,
			Face:       0,
			Object:     object.NewPointObject(vector.NewVector(pos[0], pos[1])),
			LastActive: -1,
		})
	}

	status.Instances = instances
	component.Status.Set(enemy, status)

	return enemy
}

func NewPet(gameID int64, id int64, name string, instanceCount int) *donburi.Entry {
	pet := Pet.Spawn()

	status := &model.StatusData{
		GameID:   gameID,
		ID:       id,
		Name:     name,
		Role:     role.Pet,
		HP:       1,
		MaxHP:    1,
		Mana:     10000,
		MaxMana:  10000,
		BuffList: model.NewBuffList(),
	}

	instances := []*model.Instance{}
	for i := 0; i < instanceCount; i++ {
		instances = append(instances, &model.Instance{
			Status:     status,
			Face:       0,
			Object:     object.NewPointObject(vector.NewVector(0, 0)),
			LastActive: -1,
		})
	}

	status.Instances = instances
	component.Status.Set(pet, status)

	return pet
}

func NewLimitBreakNPC(gameID int64, id int64) *donburi.Entry {
	limitBreak := LimitBreak.Spawn()

	status := &model.StatusData{
		GameID:   gameID,
		ID:       id,
		Name:     "LimitBreak",
		Role:     role.LimitBreak,
		HP:       1,
		MaxHP:    1,
		BuffList: model.NewBuffList(),
	}

	instances := []*model.Instance{}
	instances = append(instances, &model.Instance{
		Status:     status,
		Face:       0,
		Object:     object.NewPointObject(vector.NewVector(0, 0)),
		LastActive: -1,
	})

	status.Instances = instances
	component.Status.Set(limitBreak, status)

	log.Println("LimitBreak:", id)

	return limitBreak
}

func NewPlayer(role role.RoleType, pos f64.Vec2, detail *fflogs.PlayerDetail) *donburi.Entry {
	player := Player.Spawn()

	var id int64 = 0

	name := "测试玩家"

	if detail != nil {
		id = detail.ID
		name = fmt.Sprintf("%s @%s", detail.Name, detail.Server)
		log.Println("Player:", name)
	}

	obj := object.NewPointObject(vector.NewVector(pos[0], pos[1]))

	status := &model.StatusData{
		GameID:   -1,
		ID:       id,
		Name:     name,
		Role:     role,
		HP:       210000,
		MaxHP:    210000,
		Mana:     10000,
		MaxMana:  10000,
		BuffList: model.NewBuffList(),
		RingConfig: model.RingConfiguration{
			Texture: texture.NewTextureFromFile("asset/target_normal.png"),
			Scale:   0.1842,
		},
	}

	instances := []*model.Instance{
		{
			Status: status,
			Face:   0,
			Object: obj,
		},
	}
	status.Instances = instances

	// this scales target ring into size 50pixel, which means 1m in game
	component.Status.Set(player, status)

	return player
}

func NewMap(m *model.MapConfig) *donburi.Entry {
	bg := Background.Spawn()
	component.Map.Set(bg, &model.MapData{
		Config: m,
	})

	return bg
}

func NewGlobal() *donburi.Entry {
	global := Global.Spawn()
	component.Global.Set(global, &model.GlobalData{
		Tick:                0,
		Speed:               10,
		WorldMarkerSelected: -1,
		ShowTargetRing:      true,
		RangeDisplay:        true,
	})

	return global
}

func NewCamera() *donburi.Entry {
	camera := Camera.Spawn()
	component.Camera.Set(camera, &model.CameraData{
		ZoomFactor: 0,
		Rotation:   0,
	})

	return camera
}

func NewTimeline(data *model.TimelineData) *donburi.Entry {
	timeline := Timeline.Spawn()
	component.Timeline.Set(timeline, data)

	return timeline
}

func NewWorldMarker(markerType model.WorldMarkerType, pos f64.Vec2) *donburi.Entry {
	// each type of marker can only exists one instance
	for m := range component.WorldMarker.Iter(ctx.World) {
		marker := component.WorldMarker.Get(m)
		if marker.Type == markerType {
			marker.Position = pos

			return m
		}
	}

	marker := WorldMarker.Spawn()
	component.WorldMarker.Set(marker, &model.WorldMarkerData{
		Type:     markerType,
		Position: pos,
	})

	return marker
}

func GetGlobal() *model.GlobalData {
	return component.Global.Get(tag.Global.MustFirst(ctx.World))
}

func GetCamera() *model.CameraData {
	return component.Camera.Get(tag.Camera.MustFirst(ctx.World))
}

func IsDebug() bool {
	return component.Global.Get(tag.Global.MustFirst(ctx.World)).Debug
}

func GetTick() int64 {
	return component.Global.Get(tag.Global.MustFirst(ctx.World)).Tick / 10
}

func GetSpeed() int64 {
	return component.Global.Get(tag.Global.MustFirst(ctx.World)).Speed
}

func GetMap() *model.MapData {
	return component.Map.Get(tag.Background.MustFirst(ctx.World))
}

func GetPhase() int {
	global := component.Global.Get(tag.Global.MustFirst(ctx.World))
	if global.Phases == nil {
		return 0
	}
	// find phase by current tick
	tick := global.Tick / 10
	for i, p := range global.Phases {
		if p > tick {
			return i - 1
		}
	}

	return len(global.Phases) - 1
}

var statusCache sync.Map

func GetStatusByID(id int64) *model.StatusData {
	if v, ok := statusCache.Load(id); ok {
		if status, ok := v.(*model.StatusData); ok {
			return status
		}
	}

	for e := range component.Status.Iter(ctx.World) {
		status := component.Status.Get(e)
		if status.ID == id {
			statusCache.Store(id, status)

			return status
		}
	}

	return nil
}

func GetPlayerList() []*donburi.Entry {
	players := []*donburi.Entry{}
	for e := range tag.Player.Iter(ctx.World) {
		players = append(players, e)
	}

	return players
}

func GetBuffables() []*donburi.Entry {
	buffables := []*donburi.Entry{}
	for e := range tag.Buffable.Iter(ctx.World) {
		buffables = append(buffables, e)
	}

	return buffables
}

func GetGameObjects() []*donburi.Entry {
	objects := []*donburi.Entry{}
	for e := range tag.GameObject.Iter(ctx.World) {
		objects = append(objects, e)
	}

	return objects
}

func GetWorldMarkers() []*donburi.Entry {
	markers := []*donburi.Entry{}
	for e := range tag.WorldMarker.Iter(ctx.World) {
		markers = append(markers, e)
	}

	return markers
}

func StatusOf(e *donburi.Entry) *model.StatusData {
	return component.Status.Get(e)
}
