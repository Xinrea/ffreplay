package model

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/Xinrea/ffreplay/pkg/texture"
)

type Skill struct {
	ID        int64
	Name      string
	Icon      string
	StartTick int64
	Cast      int64
	Recast    int64

	SkillEvents *TimelineData
}

func (s Skill) Texture() *texture.Texture {
	return texture.NewAbilityTexture(s.Icon)
}

var gcdDB = sync.Map{}
var LongCast = map[int64]bool{}

func IsGCD(id int64) bool {
	if g, ok := gcdDB.Load(id); ok {
		return g.(bool)
	}
	// https://www.garlandtools.org/db/doc/action/en/2/25865.json
	resp, err := http.Get(fmt.Sprintf("https://www.garlandtools.org/db/doc/action/en/2/%d.json", id))
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		gcdDB.Store(id, false)
		return false
	}

	var Data struct {
		Action struct {
			GCD int `json:"gcd"`
		} `json:"action"`
	}
	jsonStr, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(jsonStr, &Data)
	if err != nil {
		log.Println(string(jsonStr))
		log.Println(err)
		return false
	}
	gcdDB.Store(id, Data.Action.GCD == 1)
	return Data.Action.GCD == 1
}
