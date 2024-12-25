package scenes

import (
	"encoding/json"
	"image/color"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/renderer"
	"github.com/Xinrea/ffreplay/internal/system"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

type FFScene struct {
	ecs    *ecs.ECS
	client *fflogs.FFLogsClient
	code   string
	fight  int
	system *system.System
}

type MapPreset struct {
	Maps []MapPresetItem
}

type MapPresetItem struct {
	ID     int
	Path   string
	Offset struct {
		X float64
		Y float64
	}
	Phases []struct {
		Path   string
		Offset struct {
			X float64
			Y float64
		}
	}
}

func (m MapPresetItem) Load() *model.MapConfig {
	config := &model.MapConfig{}
	config.ID = m.ID
	config.DefaultMap.Texture = texture.NewTextureFromFile(m.Path)
	config.DefaultMap.Offset.X = m.Offset.X
	config.DefaultMap.Offset.Y = m.Offset.Y
	for _, p := range m.Phases {
		item := model.MapItem{}
		item.Texture = texture.NewTextureFromFile(p.Path)
		item.Offset.X = p.Offset.X
		item.Offset.Y = p.Offset.Y
		config.Phases = append(config.Phases, item)
	}
	return config
}

var MapCache = map[int]MapPresetItem{}

func init() {
	if util.IsWasm() {
		resp, err := http.Get("asset/floor/floor.json")
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		var config MapPreset
		err = json.NewDecoder(resp.Body).Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range config.Maps {
			MapCache[m.ID] = m
		}
		return
	}
	f, err := os.Open("asset/floor/floor.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var config MapPreset
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range config.Maps {
		MapCache[m.ID] = m
	}
}

func NewFFScene(client *fflogs.FFLogsClient, code string, fight int) *FFScene {
	ecs := ecs.NewECS(donburi.NewWorld())
	ms := &FFScene{
		ecs:    ecs,
		client: client,
		code:   code,
		fight:  fight,
		system: system.NewSystem(true),
	}
	ms.init()
	return ms
}

func (ms *FFScene) init() {
	// create a global config entry
	g := entry.NewGlobal(ms.ecs)
	// create basic camera
	entry.NewCamera(ms.ecs)
	global := component.Global.Get(g)

	ms.ecs.AddSystem(ms.system.Update)

	renderer := renderer.NewRenderer()
	renderer.Init(ms.ecs)

	go func() {
		fights := ms.client.QueryReportFights(ms.code)
		fightIndex := -1
		if ms.fight == -1 {
			ms.fight = fights[len(fights)-1].ID
			fightIndex = len(fights) - 1
		} else {
			for i := range fights {
				if fights[i].ID == ms.fight {
					fightIndex = i
					break
				}
			}
		}

		if fightIndex == -1 {
			log.Fatal("Invalid fight id")
		}

		fight := fights[fightIndex]
		log.Println("Fight name:", fight.Name)
		global.FightDuration.Store(int64(fight.EndTime - fight.StartTime))
		phases := []int64{}
		for _, p := range fight.PhaseTransitions {
			phases = append(phases, util.MSToTick(p.StartTime-int64(fight.StartTime)))
		}
		sort.Slice(phases, func(i, j int) bool {
			return phases[i] < phases[j]
		})
		global.Phases = phases
		log.Println(fight.Maps)
		// create a background base on mapID
		if m, ok := MapCache[fight.Maps[0].ID]; ok {
			entry.NewMap(ms.ecs, m.Load())
		} else {
			// get default map from fflogs
			gameMap := ms.client.QueryMapInfo(fight.Maps[0].ID)
			// fflogs map offset is based on top-left corner
			texture := texture.NewMapTexture(gameMap.FileName)
			// TODO refactor to remove scale/offset compensation
			// 地图的偏移确实没有问题，但是一开始在实现 m1s-m4s 以及绝伊甸时忽略了这个，直接在 player 的位置
			// 坐标计算时进行了固定的偏移，相当于得到的地图偏移是(-100, -100)，单位实际上都是游戏内的 m
			// 如果某些地图偏移不是 (-100, -100)，那么需要下面的计算，进行偏移补偿
			// SizeFactor 显然最开始也是先入为主忽略了，都是按照 400 来计算的，因此也要进行补偿
			scale := 6.25 * 400 / gameMap.SizeFactor
			oX := (gameMap.OffsetX + 100) * 25
			oY := (gameMap.OffsetY + 100) * 25
			mapItem := model.MapItem{
				Texture: texture,
				Scale:   scale,
				Offset: struct {
					X float64
					Y float64
				}{
					-float64(oX),
					-float64(oY),
				},
			}
			log.Println(mapItem)
			entry.NewMap(ms.ecs, &model.MapConfig{
				ID:         fight.Maps[0].ID,
				DefaultMap: mapItem,
			})
		}
		// query worldMarkers
		markers := ms.client.QueryWorldMarkers(ms.code, fight.ID)
		// create markers
		for _, m := range markers {
			if m.MapID != fight.Maps[0].ID {
				continue
			}
			entry.NewMarker(ms.ecs, model.MarkerA+model.MarkerType(m.Icon-1), f64.Vec2{float64(m.X-10000) / 100 * 25, float64(m.Y-10000) / 100 * 25})
		}
		// initialize player events
		players := ms.client.QueryFightPlayers(ms.code, fight.ID)
		actors := ms.client.QueryActors(ms.code)
		log.Println("Actors:", actors)
		if len(actors) == 0 {
			log.Fatal("No actor found")
			return
		}
		actorInfo := func(id int64) fflogs.Actor {
			for _, b := range actors {
				if b.ID == id {
					return b
				}
			}
			return fflogs.Actor{}
		}

		pBeforeLoad := time.Now()
		events := data.FetchLogEvents(ms.client, ms.code, fight)
		filterTarget := func(targetID int64) []fflogs.FFLogsEvent {
			ret := []fflogs.FFLogsEvent{}
			for _, e := range events {
				if e.SourceID != nil && *e.SourceID == targetID {
					ret = append(ret, e)
				}
			}
			return ret
		}
		filterLimitbreak := func() []fflogs.FFLogsEvent {
			ret := []fflogs.FFLogsEvent{}
			for _, e := range events {
				if e.Type == fflogs.Limitbreakupdate {
					ret = append(ret, e)
				}
			}
			return ret
		}
		global.LoadTotal = len(events)
		var wg sync.WaitGroup
		for _, p := range players.Tanks {
			wg.Add(1)
			go func() {
				defer wg.Done()
				events := filterTarget(p.ID)
				data.PreloadAbilityInfo(events, &global.LoadCount)
				ms.system.AddEventLine(p.ID, events)
			}()
		}
		for _, p := range players.Healers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				events := filterTarget(p.ID)
				data.PreloadAbilityInfo(events, &global.LoadCount)
				ms.system.AddEventLine(p.ID, events)
			}()
		}
		for _, p := range players.DPS {
			wg.Add(1)
			go func() {
				defer wg.Done()
				events := filterTarget(p.ID)
				data.PreloadAbilityInfo(events, &global.LoadCount)
				ms.system.AddEventLine(p.ID, events)
			}()
		}

		// initialize pet events
		for _, e := range fight.FriendlyPets {
			wg.Add(1)
			go func() {
				defer wg.Done()
				events := filterTarget(e.ID)
				data.PreloadAbilityInfo(events, &global.LoadCount)
				ms.system.AddEventLine(e.ID, events)
			}()
		}

		// initialize enemy events
		for _, e := range fight.EnemyNPCs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				events := filterTarget(e.ID)
				data.PreloadAbilityInfo(events, &global.LoadCount)
				ms.system.AddEventLine(e.ID, events)
			}()
		}

		// create environment NPC events
		{
			events := filterTarget(-1)
			data.PreloadAbilityInfo(events, &global.LoadCount)
			ms.system.AddEventLine(-1, events)
		}

		// create limitbreak events
		{
			ms.system.AddLimitbreakEvents(filterLimitbreak())
		}

		// create players
		posPreset := []f64.Vec2{
			{0, -200},
			{0, 200},
			{200, 0},
			{-200, 0},
			{-200, 200},
			{200, 200},
			{-200, -200},
			{200, -200},
		}

		playerCnt := 0
		for _, t := range players.Tanks {
			ms.system.AddEntry(t.ID, entry.NewPlayer(ms.ecs, t.Type, posPreset[playerCnt], &t))
			playerCnt++
		}
		for _, h := range players.Healers {
			ms.system.AddEntry(h.ID, entry.NewPlayer(ms.ecs, h.Type, posPreset[playerCnt], &h))
			playerCnt++
		}
		for _, d := range players.DPS {
			ms.system.AddEntry(d.ID, entry.NewPlayer(ms.ecs, d.Type, posPreset[playerCnt], &d))
			playerCnt++
		}
		getInstanceCount := func(id int64) int {
			for _, e := range fight.EnemyNPCs {
				if e.ID == id {
					return e.InstanceCount
				}
			}
			return 1
		}
		// create pets
		for _, p := range fight.FriendlyPets {
			ms.system.AddEntry(p.ID, entry.NewPet(ms.ecs, p.GameID, p.ID, "", p.InstanceCount))
		}
		// create enemies
		for _, e := range fight.EnemyNPCs {
			info := actorInfo(e.ID)
			ms.system.AddEntry(e.ID, entry.NewEnemy(ms.ecs, f64.Vec2{0, 0}, 5, info.GameID, e.ID, info.Name, info.SubType == "Boss", getInstanceCount(e.ID)))
		}

		// create environment NPC
		ms.system.AddEntry(-1, entry.NewEnemy(ms.ecs, f64.Vec2{}, 1, 0, -1, "environment", false, 1))

		// create a timeline
		// entry.NewTimeline(ms.ecs, &model.TimelineData{
		// 	Name:      "example",
		// 	BeginTime: util.Time(),
		// 	Events: []*model.Event{
		// 		{
		// 			Offset: 0,
		// 			Action: func(ecs *ecs.ECS) {
		// 				for e := range tag.PartyMember.Iter(ecs.World) {
		// 					entry.CastSkill(ecs, 2000, 1000, skills.NewSkillTestFan(ecs, enemy, e, 30))
		// 				}
		// 			},
		// 		},
		// 		{
		// 			Offset: 2000,
		// 			Action: func(ecs *ecs.ECS) {
		// 				for e := range tag.PartyMember.Iter(ecs.World) {
		// 					entry.CastSkill(ecs, 5000, 5000, skills.NewSkillTestRectLocked(ecs, enemy, e, 200))
		// 				}
		// 			},
		// 		},
		// 		{
		// 			Offset: 7500,
		// 			Action: func(ecs *ecs.ECS) {
		// 				for e := range tag.PartyMember.Iter(ecs.World) {
		// 					entry.CastSkill(ecs, 2000, 1000, skills.NewSkillTestRect(ecs, enemy, e, 200))
		// 				}
		// 			},
		// 		},
		// 	},
		// })

		wg.Wait()
		log.Println("Loading cost", time.Since(pBeforeLoad))
		global.Loaded.Store(true)
		log.Println("Game scene initialized")
	}()
}

func (ms *FFScene) Reset() {
	ms.system.Reset()
}

func (ms *FFScene) Update() {
	ms.ecs.Update()
}

func (ms *FFScene) Layout(w, h int) {
	ms.system.Layout(w, h)
}

func (ms *FFScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	ms.ecs.Draw(screen)
}
