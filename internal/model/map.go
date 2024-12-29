package model

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
)

type MapData struct {
	Config *MapConfig
}

type MapConfig struct {
	CurrentMap int
	Maps       map[int]MapItem
	Phases     []MapItem
}

type MapItem struct {
	ID      int
	Texture *ebiten.Image
	Scale   float64
	Offset  struct {
		X float64
		Y float64
	}
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

func (m MapPresetItem) Load() *MapConfig {
	config := &MapConfig{
		CurrentMap: m.ID,
		Maps:       make(map[int]MapItem),
	}
	defaultItem := MapItem{}
	defaultItem.ID = m.ID
	defaultItem.Texture = texture.NewTextureFromFile(m.Path)
	defaultItem.Offset.X = m.Offset.X
	defaultItem.Offset.Y = m.Offset.Y
	config.Maps[m.ID] = defaultItem
	for _, p := range m.Phases {
		item := MapItem{}
		item.ID = m.ID
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

type MapBoundary struct {
	ID   int   `json:"mapID"`
	MinX int64 `json:"mapMinX"`
	MaxX int64 `json:"mapMaxX"`
	MinY int64 `json:"mapMinY"`
	MaxY int64 `json:"mapMaxY"`
}
