package model

import (
	"math"

	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

type CameraData struct {
	ViewPort   f64.Vec2
	Position   vector.Vector
	ZoomFactor int
	Rotation   float64
}

func (c *CameraData) viewportCenter() f64.Vec2 {
	return f64.Vec2{
		c.ViewPort[0] * 0.5,
		c.ViewPort[1] * 0.5,
	}
}

// Window size without scaled
func (c *CameraData) WindowSize() (float64, float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	return c.ViewPort[0] / s, c.ViewPort[1] / s
}

func (c *CameraData) Update(v f64.Vec2) {
	s := ebiten.Monitor().DeviceScaleFactor()
	c.ViewPort = f64.Vec2{v[0] * s, v[1] * s}
}

func (c *CameraData) WorldMatrix() ebiten.GeoM {
	s := ebiten.Monitor().DeviceScaleFactor()
	m := ebiten.GeoM{}
	m.Translate(-c.viewportCenter()[0], -c.viewportCenter()[1])
	m.Scale(
		math.Pow(1.01, float64(c.ZoomFactor))/s,
		math.Pow(1.01, float64(c.ZoomFactor))/s,
	)
	m.Rotate(c.Rotation)
	m.Translate(c.Position[0], c.Position[1])
	return m
}

func (c *CameraData) WorldToScreen(x, y float64) (float64, float64) {
	worldInverted := c.WorldMatrixInverted()
	return worldInverted.Apply(x, y)
}

func (c *CameraData) ScreenToWorld(x, y float64) (float64, float64) {
	world := c.WorldMatrix()
	return world.Apply(x, y)
}

func (c *CameraData) WorldMatrixInverted() ebiten.GeoM {
	worldGeo := c.WorldMatrix()
	worldGeo.Invert()
	return worldGeo
}
