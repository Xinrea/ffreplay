package model

import (
	"os"
	"strconv"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

// HeadMarkerTextures maps head marker type (1-based) to its texture.
var HeadMarkerTextures = map[int]*ebiten.Image{}

func initHeadMarkerTextures() {
	for i := 1; i <= 8; i++ {
		path := "asset/headmarker/marker" + strconv.Itoa(i) + ".png"
		if _, err := os.Stat(path); err == nil {
			HeadMarkerTextures[i] = texture.NewTextureFromFile(path)
		}
	}
}

// GetHeadMarkerTexture returns the texture for a head marker type, or nil.
func GetHeadMarkerTexture(t int) *ebiten.Image {
	if t <= 0 {
		return nil
	}
	return HeadMarkerTextures[t]
}
