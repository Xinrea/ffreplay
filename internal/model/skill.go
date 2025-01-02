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
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

var BorderTexture = texture.NewTextureFromFile("asset/skillborder.png")
var BorderGeoM = ebiten.GeoM{}

func init() {
	BorderGeoM = texture.CenterGeoM(BorderTexture)
}

type Skill struct {
	ID        int64
	Name      string
	Icon      string
	StartTick int64
	Cast      int64
	Recast    int64
	IsGCD     bool

	SkillEvents *TimelineData
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

func init() {
	f, err := asset.AssetFS.Open("asset/gamedata/Action.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
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
		return g.(*ActionInfo)
	}
	// https://www.garlandtools.org/db/doc/action/en/2/25865.json
	resp, err := http.Get(fmt.Sprintf("https://www.garlandtools.org/db/doc/action/en/2/%d.json", id))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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
