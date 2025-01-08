package model

import (
	"fmt"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

var MarkerTextures = []*ebiten.Image{}

func initMarkerTextures() {
	// headmarker id 1 to 17
	for i := 1; i <= 17; i++ {
		MarkerTextures = append(MarkerTextures, texture.NewTextureFromFile(fmt.Sprintf("asset/marker/marker%d.png", i)))
	}
}
