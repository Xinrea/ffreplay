package scenes

import (
	"image/color"
	"log"
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
		// create a background base on mapID
		if m, ok := model.MapCache[fight.Maps[0].ID]; ok {
			entry.NewMap(ms.ecs, m.Load())
		} else {
			queryMapItem := func(id int) model.MapItem {
				// get default map from fflogs
				gameMap := ms.client.QueryMapInfo(id)
				// fflogs map offset is based on top-left corner
				texture := texture.NewMapTexture(gameMap.FileName)
				// TODO refactor to remove scale/offset compensation
				// SizeFactor 显然最开始都是按照 400 来计算的，因此要进行补偿
				scale := 6.25 * 400 / gameMap.SizeFactor
				// meters into pixels
				mapItem := model.MapItem{
					ID:      id,
					Texture: texture,
					Scale:   scale,
					Offset: struct {
						X float64
						Y float64
					}{
						-float64(gameMap.OffsetX),
						-float64(gameMap.OffsetY),
					},
				}
				return mapItem
			}
			mapItems := map[int]model.MapItem{}
			for _, m := range fight.Maps {
				mapItems[m.ID] = queryMapItem(m.ID)
			}
			log.Println("Initial map", mapItems[fight.Maps[0].ID])
			entry.NewMap(ms.ecs, &model.MapConfig{
				CurrentMap: fight.Maps[0].ID,
				Maps:       mapItems,
			})
		}
		// query worldMarkers
		markers := ms.client.QueryWorldMarkers(ms.code, fight.ID)
		// create markers
		for _, m := range markers {
			if m.MapID != fight.Maps[0].ID {
				continue
			}
			entry.NewMarker(ms.ecs, model.MarkerA+model.MarkerType(m.Icon-1), f64.Vec2{float64(m.X) / 100 * 25, float64(m.Y) / 100 * 25})
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
		filterMapChange := func() []fflogs.FFLogsEvent {
			ret := []fflogs.FFLogsEvent{}
			for _, e := range events {
				if e.Type == fflogs.MapChange {
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

		{
			ms.system.AddMapChangeEvents(filterMapChange())
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
