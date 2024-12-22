package model

import "github.com/Xinrea/ffreplay/pkg/texture"

type MapData struct {
	Config *MapConfig
}

type MapConfig struct {
	ID         int
	DefaultMap MapItem
	Phases     []MapItem
}

type MapItem struct {
	Texture *texture.Texture
	Scale   float64
	Offset  struct {
		X float64
		Y float64
	}
}
