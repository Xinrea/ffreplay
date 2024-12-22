package model

import (
	"image/color"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f64"
)

type MarkerType int

const (
	MarkerA MarkerType = iota
	MarkerB
	MarkerC
	MarkerD
	Marker1
	Marker2
	Marker3
	Marker4
)

var MarkerConfigs = map[MarkerType]*MarkerConfig{}

type MarkerData struct {
	Type     MarkerType
	Position f64.Vec2
}

type MarkerConfig struct {
	Type            MarkerType
	Texture         *texture.Texture
	Background      *ebiten.Image
	BackgroundColor color.NRGBA
}

func (m *MarkerConfig) preprocess() {
	m.Background = ebiten.NewImage(m.Texture.Img().Bounds().Dx(), m.Texture.Img().Bounds().Dy())
	mainColor := m.BackgroundColor
	w := float32(m.Background.Bounds().Dx())
	h := float32(m.Background.Bounds().Dy())
	if m.Type > MarkerD {
		mainColor.A = 128
		m.Background.Fill(mainColor)
		mainColor.A = 230
		vector.StrokeRect(m.Background, 0, 0, w, h, 16, mainColor, true)
	} else {
		mainColor.A = 128
		vector.DrawFilledCircle(m.Background, float32(m.Background.Bounds().Dx()/2), float32(m.Background.Bounds().Dy()/2), float32(m.Background.Bounds().Dx()/2), mainColor, true)
		mainColor.A = 230
		vector.StrokeCircle(m.Background, float32(m.Background.Bounds().Dx()/2), float32(m.Background.Bounds().Dy()/2), float32(m.Background.Bounds().Dx()/2-4), 8, mainColor, true)
	}
}

func init() {
	// setup marker configs
	MarkerConfigs[Marker1] = &MarkerConfig{
		Type:            Marker1,
		Texture:         texture.NewTextureFromFile("asset/marker/marker1.png"),
		BackgroundColor: color.NRGBA{184, 94, 107, 0},
	}
	MarkerConfigs[Marker2] = &MarkerConfig{
		Type:            Marker2,
		Texture:         texture.NewTextureFromFile("asset/marker/marker2.png"),
		BackgroundColor: color.NRGBA{253, 255, 205, 0},
	}
	MarkerConfigs[Marker3] = &MarkerConfig{
		Type:            Marker3,
		Texture:         texture.NewTextureFromFile("asset/marker/marker3.png"),
		BackgroundColor: color.NRGBA{163, 210, 248, 0},
	}
	MarkerConfigs[Marker4] = &MarkerConfig{
		Type:            Marker4,
		Texture:         texture.NewTextureFromFile("asset/marker/marker4.png"),
		BackgroundColor: color.NRGBA{204, 161, 250, 0},
	}
	MarkerConfigs[MarkerA] = &MarkerConfig{
		Type:            MarkerA,
		Texture:         texture.NewTextureFromFile("asset/marker/markerA.png"),
		BackgroundColor: color.NRGBA{184, 94, 107, 0},
	}
	MarkerConfigs[MarkerB] = &MarkerConfig{
		Type:            MarkerB,
		Texture:         texture.NewTextureFromFile("asset/marker/markerB.png"),
		BackgroundColor: color.NRGBA{253, 255, 205, 0},
	}
	MarkerConfigs[MarkerC] = &MarkerConfig{
		Type:            MarkerC,
		Texture:         texture.NewTextureFromFile("asset/marker/markerC.png"),
		BackgroundColor: color.NRGBA{163, 210, 248, 0},
	}
	MarkerConfigs[MarkerD] = &MarkerConfig{
		Type:            MarkerD,
		Texture:         texture.NewTextureFromFile("asset/marker/markerD.png"),
		BackgroundColor: color.NRGBA{204, 161, 250, 0},
	}

	for _, config := range MarkerConfigs {
		config.preprocess()
	}
}
