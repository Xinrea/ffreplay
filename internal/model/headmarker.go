package model

import (
	"log"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

type HeadMarkerType int

const (
	HeadMarkerType1 HeadMarkerType = iota
	HeadMarkerType2
	HeadMarkerType3
)

type HeadMarker struct {
	Type HeadMarkerType
}

func NewHeadMarker(t HeadMarkerType) *HeadMarker {
	return &HeadMarker{Type: t}
}

func (h *HeadMarker) Texture() *ebiten.Image {
	switch h.Type {
	case HeadMarkerType1:
		return texture.NewTextureFromFile("asset/headmarker/marker1.png")
	case HeadMarkerType2:
		return texture.NewTextureFromFile("asset/headmarker/marker2.png")
	case HeadMarkerType3:
		return texture.NewTextureFromFile("asset/headmarker/marker3.png")
	default:
		log.Fatal("invalid head marker type")
	}

	return nil
}
