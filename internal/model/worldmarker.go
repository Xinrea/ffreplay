package model

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f64"
)

type WorldMarkerType int

const (
	WorldMarkerA WorldMarkerType = iota
	WorldMarkerB
	WorldMarkerC
	WorldMarkerD
	WorldMarker1
	WorldMarker2
	WorldMarker3
	WorldMarker4
)

var WorldMarkerConfigs = map[WorldMarkerType]*WorldMarkerConfig{}

type WorldMarkerData struct {
	Type     WorldMarkerType
	Position f64.Vec2
}

type WorldMarkerConfig struct {
	Type            WorldMarkerType
	Texture         *ebiten.Image
	Background      *ebiten.Image
	BackgroundColor color.NRGBA
}

func (m *WorldMarkerConfig) preprocess() {
	m.Background = ebiten.NewImage(m.Texture.Bounds().Dx(), m.Texture.Bounds().Dy())
	mainColor := m.BackgroundColor
	w := float32(m.Background.Bounds().Dx())
	h := float32(m.Background.Bounds().Dy())

	if m.Type > WorldMarkerD {
		mainColor.A = 128
		m.Background.Fill(mainColor)
		mainColor.A = 230
		vector.StrokeRect(m.Background, 0, 0, w, h, 16, mainColor, true)
	} else {
		mainColor.A = 128
		vector.DrawFilledCircle(
			m.Background,
			float32(m.Background.Bounds().Dx()/2),
			float32(m.Background.Bounds().Dy()/2),
			float32(m.Background.Bounds().Dx()/2),
			mainColor,
			true)

		mainColor.A = 230
		vector.StrokeCircle(
			m.Background,
			float32(m.Background.Bounds().Dx()/2),
			float32(m.Background.Bounds().Dy()/2),
			float32(m.Background.Bounds().Dx()/2-4),
			8,
			mainColor,
			true)
	}
}

func initWorldMarkerTextures() {
	// setup marker configs
	WorldMarkerConfigs[WorldMarker1] = &WorldMarkerConfig{
		Type:            WorldMarker1,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarker1.png"),
		BackgroundColor: color.NRGBA{184, 94, 107, 0},
	}
	WorldMarkerConfigs[WorldMarker2] = &WorldMarkerConfig{
		Type:            WorldMarker2,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarker2.png"),
		BackgroundColor: color.NRGBA{253, 255, 205, 0},
	}
	WorldMarkerConfigs[WorldMarker3] = &WorldMarkerConfig{
		Type:            WorldMarker3,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarker3.png"),
		BackgroundColor: color.NRGBA{163, 210, 248, 0},
	}
	WorldMarkerConfigs[WorldMarker4] = &WorldMarkerConfig{
		Type:            WorldMarker4,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarker4.png"),
		BackgroundColor: color.NRGBA{204, 161, 250, 0},
	}
	WorldMarkerConfigs[WorldMarkerA] = &WorldMarkerConfig{
		Type:            WorldMarkerA,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarkerA.png"),
		BackgroundColor: color.NRGBA{184, 94, 107, 0},
	}
	WorldMarkerConfigs[WorldMarkerB] = &WorldMarkerConfig{
		Type:            WorldMarkerB,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarkerB.png"),
		BackgroundColor: color.NRGBA{253, 255, 205, 0},
	}
	WorldMarkerConfigs[WorldMarkerC] = &WorldMarkerConfig{
		Type:            WorldMarkerC,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarkerC.png"),
		BackgroundColor: color.NRGBA{163, 210, 248, 0},
	}
	WorldMarkerConfigs[WorldMarkerD] = &WorldMarkerConfig{
		Type:            WorldMarkerD,
		Texture:         texture.NewTextureFromFile("asset/marker/worldmarkerD.png"),
		BackgroundColor: color.NRGBA{204, 161, 250, 0},
	}

	for _, config := range WorldMarkerConfigs {
		config.preprocess()
	}
}
