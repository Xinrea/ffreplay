package scenes

import (
	"image/color"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/data"
	"github.com/Xinrea/ffreplay/internal/data/fflogs"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/errors"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/renderer"
	"github.com/Xinrea/ffreplay/internal/system"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

type FFScene struct {
	ecs      *ecs.ECS
	client   *fflogs.FFLogsClient
	code     string
	fight    int
	system   *system.System
	renderer *renderer.Renderer
	ui       ui.UI
	global   *model.GlobalData
	camera   *model.CameraData
	screenW  int
	screenH  int
}

type FFLogsOpt struct {
	ClientID     string
	ClientSecret string
	Report       string
	Fight        int
}

func NewFFScene(opt *FFLogsOpt) *FFScene {
	ecs := ecs.NewECS(donburi.NewWorld())
	// create a global config entry
	entry.NewGlobal(ecs)
	// create basic camera
	entry.NewCamera(ecs)
	entry.NewMap(ecs, nil)
	// init system
	system := system.NewSystem()
	system.Init(ecs)
	// init renderer
	renderer := renderer.NewRenderer()
	renderer.Init(ecs)
	ms := &FFScene{
		ecs:      ecs,
		global:   entry.GetGlobal(ecs),
		camera:   entry.GetCamera(ecs),
		system:   system,
		renderer: renderer,
	}

	if opt != nil {
		ms.ui = ui.NewReplayUI(ecs)
		ms.client = fflogs.NewFFLogsClient(opt.ClientID, opt.ClientSecret)
		ms.code = opt.Report
		ms.fight = opt.Fight
		ms.global.ReplayMode = true

		go ms.loadFFLogsReport()
	} else {
		ms.ui = ui.NewPlaygroundUI(ecs)
		ms.global.Loaded.Store(true)
	}

	log.Println("Scene created")

	return ms
}

func (ms *FFScene) loadFFLogsReport() {
	fights := ms.client.QueryReportFights(ms.code)
	fightIndex := ms.findFightIndex(fights)

	if fightIndex == -1 {
		log.Println("Invalid fight id")
		os.Exit(errors.ErrorInvalidFightID)
	}

	fight := fights[fightIndex]

	log.Println("Fight name:", fight.Name)

	ms.global.FightDuration.Store(int64(fight.EndTime - fight.StartTime))

	ms.global.Phases = extractPhaseTicks(fight)

	// setup map
	ms.setupMap(fight)

	// initialize player events
	players := ms.client.QueryFightPlayers(ms.code, fight.ID)

	actors := ms.client.QueryActors(ms.code)
	if len(actors) == 0 {
		log.Fatal("No actor found")

		return
	}

	pBeforeLoad := time.Now()
	status, events := data.FetchLogEvents(ms.client, ms.code, fight)

	isDungeonReport := events[0].Type == fflogs.DungeonStart
	if isDungeonReport {
		ms.global.RenderNPC = true
	}

	ms.global.LoadTotal = len(events)

	var wg sync.WaitGroup

	ms.loadPlayerEvents(&wg, players.Tanks, status, events)
	ms.loadPlayerEvents(&wg, players.Healers, status, events)
	ms.loadPlayerEvents(&wg, players.DPS, status, events)
	ms.loadPetEvents(&wg, fight.FriendlyPets, status, events)
	ms.loadEnemyEvents(&wg, fight.EnemyNPCs, status, events)
	ms.loadEnvironmentEvents(status, events)
	ms.loadSpecialEvents(events, fight, isDungeonReport)

	// create players
	ms.createPlayers(players)

	// create pets
	ms.createPets(fight.FriendlyPets)

	// create enemies
	ms.createEnemies(fight, actors)

	// create environment NPC
	ms.system.AddEntry(-1, entry.NewEnemy(ms.ecs, f64.Vec2{}, 1, 0, -1, "environment", false, 1))

	wg.Wait()
	log.Println("Loading cost", time.Since(pBeforeLoad))
	ms.global.Loaded.Store(true)
	log.Println("FFLogs report loaded")
}

func (ms *FFScene) findFightIndex(fights []fflogs.ReportFight) int {
	if ms.fight == -1 {
		ms.fight = fights[len(fights)-1].ID

		return len(fights) - 1
	}

	for i := range fights {
		if fights[i].ID == ms.fight {
			return i
		}
	}

	return -1
}

func (ms *FFScene) loadPlayerEvents(
	wg *sync.WaitGroup,
	players []fflogs.PlayerDetail,
	status map[int64]data.InstanceStatus,
	events []fflogs.FFLogsEvent,
) {
	for _, p := range players {
		wg.Add(1)

		go func(p fflogs.PlayerDetail) {
			defer wg.Done()

			playerEvents := ms.filterTargetEvents(p.ID, events)
			data.PreloadAbilityInfo(playerEvents, &ms.global.LoadCount)
			ms.system.AddEventLine(p.ID, status[p.ID], playerEvents)
		}(p)
	}
}

func (ms *FFScene) loadPetEvents(
	wg *sync.WaitGroup,
	pets []fflogs.ReportFightNPC,
	status map[int64]data.InstanceStatus,
	events []fflogs.FFLogsEvent,
) {
	for _, e := range pets {
		wg.Add(1)

		go func(e fflogs.ReportFightNPC) {
			defer wg.Done()

			petEvents := ms.filterTargetEvents(e.ID, events)
			data.PreloadAbilityInfo(petEvents, &ms.global.LoadCount)
			ms.system.AddEventLine(e.ID, status[e.ID], petEvents)
		}(e)
	}
}

func (ms *FFScene) loadEnemyEvents(
	wg *sync.WaitGroup,
	enemies []fflogs.ReportFightNPC,
	status map[int64]data.InstanceStatus,
	events []fflogs.FFLogsEvent,
) {
	for _, e := range enemies {
		wg.Add(1)

		go func(e fflogs.ReportFightNPC) {
			defer wg.Done()

			enemyEvents := ms.filterTargetEvents(e.ID, events)
			data.PreloadAbilityInfo(enemyEvents, &ms.global.LoadCount)
			ms.system.AddEventLine(e.ID, status[e.ID], enemyEvents)
		}(e)
	}
}

func (ms *FFScene) loadEnvironmentEvents(status map[int64]data.InstanceStatus, events []fflogs.FFLogsEvent) {
	environmentEvents := ms.filterTargetEvents(-1, events)
	data.PreloadAbilityInfo(environmentEvents, &ms.global.LoadCount)
	ms.system.AddEventLine(-1, status[-1], environmentEvents)
}

func (ms *FFScene) loadSpecialEvents(events []fflogs.FFLogsEvent, fight fflogs.ReportFight, isDungeonReport bool) {
	ms.system.AddLimitbreakEvents(ms.filterLimitbreakEvents(events))
	ms.system.AddMapChangeEvents(ms.filterMapChangeEvents(events, fight, isDungeonReport))
	ms.system.AddWorldMarkerEvents(ms.filterMarkerChangeEvents(events))
}

func (ms *FFScene) createPlayers(players *fflogs.PlayerDetails) {
	for _, t := range players.Tanks {
		ms.system.AddEntry(t.ID, entry.NewPlayer(ms.ecs, role.StringToRole(t.Type), f64.Vec2{}, &t))
	}

	for _, h := range players.Healers {
		ms.system.AddEntry(h.ID, entry.NewPlayer(ms.ecs, role.StringToRole(h.Type), f64.Vec2{}, &h))
	}

	for _, d := range players.DPS {
		ms.system.AddEntry(d.ID, entry.NewPlayer(ms.ecs, role.StringToRole(d.Type), f64.Vec2{}, &d))
	}
}

func (ms *FFScene) createPets(pets []fflogs.ReportFightNPC) {
	for _, p := range pets {
		ms.system.AddEntry(p.ID, entry.NewPet(ms.ecs, p.GameID, p.ID, "", p.InstanceCount))
	}
}

func (ms *FFScene) createEnemies(fight fflogs.ReportFight, actors []fflogs.Actor) {
	for _, e := range fight.EnemyNPCs {
		info := ms.actorInfo(e.ID, actors)
		ms.system.AddEntry(
			e.ID,
			entry.NewEnemy(
				ms.ecs,
				f64.Vec2{0, 0},
				5,
				info.GameID,
				e.ID,
				info.Name,
				info.SubType == "Boss",
				ms.getInstanceCount(e.ID, fight),
			),
		)
	}
}

func (ms *FFScene) actorInfo(id int64, actors []fflogs.Actor) fflogs.Actor {
	for _, b := range actors {
		if b.ID == id {
			return b
		}
	}

	return fflogs.Actor{}
}

func (ms *FFScene) getInstanceCount(id int64, fight fflogs.ReportFight) int {
	for _, e := range fight.EnemyNPCs {
		if e.ID == id {
			return e.InstanceCount
		}
	}

	return 1
}

func (ms *FFScene) filterTargetEvents(targetID int64, events []fflogs.FFLogsEvent) []fflogs.FFLogsEvent {
	ret := []fflogs.FFLogsEvent{}

	for _, e := range events {
		if e.SourceID != nil && *e.SourceID == targetID {
			ret = append(ret, e)
		}
	}

	return ret
}

func (ms *FFScene) filterLimitbreakEvents(events []fflogs.FFLogsEvent) []fflogs.FFLogsEvent {
	ret := []fflogs.FFLogsEvent{}

	for _, e := range events {
		if e.Type == fflogs.Limitbreakupdate {
			ret = append(ret, e)
		}
	}

	return ret
}

func (ms *FFScene) filterMapChangeEvents(
	events []fflogs.FFLogsEvent,
	fight fflogs.ReportFight,
	isDungeonReport bool,
) []fflogs.FFLogsEvent {
	ret := []fflogs.FFLogsEvent{}

	if isDungeonReport {
		ret = append(ret, fflogs.FFLogsEvent{
			Type:  fflogs.MapChange,
			MapID: &fight.Maps[0].ID,
		})
	}

	for _, e := range events {
		if e.Type == fflogs.MapChange {
			ret = append(ret, e)
		}
	}

	return ret
}

func (ms *FFScene) filterMarkerChangeEvents(events []fflogs.FFLogsEvent) []fflogs.FFLogsEvent {
	ret := []fflogs.FFLogsEvent{}

	for _, e := range events {
		if e.Type == fflogs.WorldMarkerPlaced || e.Type == fflogs.WorldMarkerRemoved {
			ret = append(ret, e)
		}
	}

	return ret
}

func (ms *FFScene) setupMap(fight fflogs.ReportFight) {
	// create a background base on mapID
	if m, ok := model.MapCache[fight.Maps[0].ID]; ok {
		config := m.Load()
		current := config.Maps[config.CurrentMap]
		ms.camera.Position = vector.NewVector(current.Offset.X*25, current.Offset.Y*25)
		component.Map.Get(component.Map.MustFirst(ms.ecs.World)).Config = config
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

		ms.camera.Position = vector.NewVector(mapItems[fight.Maps[0].ID].Offset.X*25, mapItems[fight.Maps[0].ID].Offset.Y*25)
		component.Map.Get(component.Map.MustFirst(ms.ecs.World)).Config = &model.MapConfig{
			CurrentMap:   fight.Maps[0].ID,
			CurrentPhase: -1,
			Maps:         mapItems,
		}
	}
	// query worldMarkers
	markers := ms.client.QueryWorldMarkers(ms.code, fight.ID)
	// create markers
	for _, m := range markers {
		if m.MapID != fight.Maps[0].ID {
			continue
		}

		entry.NewWorldMarker(
			ms.ecs,
			model.WorldMarkerA+model.WorldMarkerType(m.Icon-1),
			f64.Vec2{
				float64(m.X) / 100 * 25,
				float64(m.Y) / 100 * 25,
			},
		)
	}
}

func extractPhaseTicks(fight fflogs.ReportFight) []int64 {
	phases := []int64{}
	for _, p := range fight.PhaseTransitions {
		phases = append(phases, util.MSToTick(p.StartTime-int64(fight.StartTime)))
	}

	sort.Slice(phases, func(i, j int) bool {
		return phases[i] < phases[j]
	})

	return phases
}

func (ms *FFScene) Reset() {
	ms.system.Reset()
}

func (ms *FFScene) Update() {
	ms.ecs.Update()
	ms.ui.Update(ms.screenW, ms.screenH)
}

func (ms *FFScene) Layout(w, h int) {
	ms.system.Layout(w, h)
	ms.screenW = w
	ms.screenH = h
}

func (ms *FFScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	ms.ecs.Draw(screen)
	ms.ui.Draw(screen)
}
