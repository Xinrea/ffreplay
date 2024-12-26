package model

import (
	"fmt"

	"github.com/Xinrea/ffreplay/pkg/texture"
)

var MarkerTextures = []*texture.Texture{}

func init() {
	// headmarker id 1 to 17
	for i := 1; i <= 17; i++ {
		MarkerTextures = append(MarkerTextures, texture.NewTextureFromFile(fmt.Sprintf("asset/marker/marker%d.png", i)))
	}
}
