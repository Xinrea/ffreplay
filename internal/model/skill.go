package model

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	asset "github.com/Xinrea/ffreplay"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
)

var BorderTexture = texture.NewTextureFromFile("asset/skillborder.png")

type Skill struct {
	ID          int64
	Name        string
	Icon        string
	StartTick   int64
	Cast        int64
	Recast      int64
	IsGCD       bool
	EffectRange object.Object
	Target      object.Object

	Initializer func(target object.Object, effectRange object.Object, facing float64, pos vector.Vector)
	Updater     func(target object.Object, effectRange object.Object, facing float64, pos vector.Vector)
}

func (s Skill) Texture() *ebiten.Image {
	return texture.NewAbilityTexture(s.Icon)
}

// Initialize initializes the skill's effect range, should be called before the skill is used.
func (s *Skill) Initialize(facing float64, pos vector.Vector) {
	if s.Initializer == nil {
		return
	}

	s.Initializer(s.Target, s.EffectRange, facing, pos)
}

// Update updates the skill's effect range, should be called every tick.
func (s *Skill) Update(facing float64, pos vector.Vector) {
	if s.Updater == nil {
		return
	}

	s.Updater(s.Target, s.EffectRange, facing, pos)
}

type ActionInfo struct {
	ID    int64
	Name  string
	IsGCD bool
}

var actionEntries = []ActionInfo{}

func initActionPreset() {
	f, err := asset.AssetFS.Open("asset/gamedata/Action.csv")
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		log.Panic(err)
	}
	// remove headers
	records = records[3:]
	for _, record := range records {
		id, _ := strconv.Atoi(record[0])
		actionEntries = append(actionEntries, ActionInfo{
			ID:    int64(id),
			Name:  record[1],
			IsGCD: record[41] == "58",
		})
	}
}

var additionalDB = sync.Map{}

func GetAction(id int64) *ActionInfo {
	// not normal ability
	if id > 1000000 {
		return nil
	}
	// try to get from action.csv
	if id >= 0 && id < int64(len(actionEntries)) {
		return &actionEntries[id]
	}

	if g, ok := additionalDB.Load(id); ok {
		if g == nil {
			return nil
		}

		if g, ok := g.(*ActionInfo); ok {
			return g
		}

		return nil
	}
	// https://www.garlandtools.org/db/doc/action/en/2/25865.json
	resp, err := http.Get(fmt.Sprintf("https://www.garlandtools.org/db/doc/action/en/2/%d.json", id))
	if err != nil {
		log.Println(err)

		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		additionalDB.Store(id, nil)

		return nil
	}

	var Data struct {
		Action struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			GCD  int    `json:"gcd"`
		} `json:"action"`
	}

	jsonStr, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(jsonStr, &Data)
	if err != nil {
		log.Println(string(jsonStr))
		log.Println(err)

		return nil
	}

	actionInfo := &ActionInfo{
		ID:    Data.Action.ID,
		Name:  Data.Action.Name,
		IsGCD: Data.Action.GCD == 1,
	}
	additionalDB.Store(id, actionInfo)

	return actionInfo
}
