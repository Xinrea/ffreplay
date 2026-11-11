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
	"github.com/Xinrea/ffreplay/util"
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

	Initialize func(r object.Object, facing float64, pos vector.Vector)
}

func (s Skill) Texture() *ebiten.Image {
	return texture.NewAbilityTexture(s.Icon)
}

type ActionInfo struct {
	ID    int64
	Name  string
	IsGCD bool
}

var actionEntries = []ActionInfo{}

var actionPresetUnavailable bool

func initActionPreset() {
	f, err := asset.AssetFS.Open("asset/gamedata/Action.csv")
	if err != nil {
		handleActionPresetError(err)

		return
	}
	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		handleActionPresetError(err)

		return
	}

	entries, err := parseActionRecords(records)
	if err != nil {
		handleActionPresetError(err)

		return
	}

	actionEntries = entries
}

// The sheet starts with three header rows. A short 200 (empty body or a
// truncated upload) parses successfully; slicing it used to panic.
func parseActionRecords(records [][]string) ([]ActionInfo, error) {
	if len(records) < 3 {
		return nil, fmt.Errorf("Action.csv has %d records", len(records))
	}

	entries := make([]ActionInfo, 0, len(records)-3)
	for _, record := range records[3:] {
		if len(record) < 5 {
			return nil, fmt.Errorf("Action.csv row has %d columns", len(record))
		}

		id, _ := strconv.Atoi(record[0])
		entries = append(entries, ActionInfo{
			ID:    int64(id),
			Name:  record[1],
			IsGCD: record[4] == "2" || record[4] == "3",
		})
	}

	return entries, nil
}

func handleActionPresetError(err error) {
	if util.IsWasm() {
		log.Printf("Skip Action.csv preset: %v", err)

		actionPresetUnavailable = true

		return
	}

	log.Panic(err)
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

	if actionPresetUnavailable {
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
