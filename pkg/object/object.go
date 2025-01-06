package object

import (
	"image/color"

	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

var objectTextureCache = make(map[string]*ebiten.Image)

type ObjectOption struct {
	FillColor   color.NRGBA
	StrokeColor color.NRGBA
	StrokeWidth float64
}

var DefaultNegativeSkillRangeOption = ObjectOption{
	FillColor:   color.NRGBA{235, 140, 52, 128},
	StrokeColor: color.NRGBA{235, 140, 52, 200},
	StrokeWidth: 4,
}

var DefaultPositiveSkillRangeOption = ObjectOption{
	FillColor:   color.NRGBA{102, 204, 255, 128},
	StrokeColor: color.NRGBA{102, 204, 255, 200},
	StrokeWidth: 4,
}

type Object interface {
	Position() Vector
	UpdatePosition(v Vector)
	// Rotate rotates the Object by given angle in radian, and anchor is defined by the need of game (Position point).
	// For example, circle object rotates around its center, but rectangle object rotates around its width center.
	// if you want to rotate the object around a self-defined point, you need to translate the object first.
	Rotate(r float64) Object
	UpdateRotate(r float64)
	Translate(v Vector) Object
	Scale(s float64) Object
	IsPointInside(v Vector) bool
	Render(canvas *ebiten.Image, geoM ebiten.GeoM, colorScale color.Color)
}

type DefaultObject struct {
	anchor   Vector
	rotate   float64
	scale    float64
	texture  *ebiten.Image
	initialM ebiten.GeoM
}

var _ Object = (*DefaultObject)(nil)

func (d *DefaultObject) Position() Vector {
	return d.anchor
}

func (d *DefaultObject) UpdatePosition(v Vector) {
	d.anchor = v
}

func (d *DefaultObject) Rotate(r float64) Object {
	d.rotate += r

	return d
}

func (d *DefaultObject) UpdateRotate(r float64) {
	d.rotate = r
}

func (d *DefaultObject) Translate(v Vector) Object {
	d.anchor = d.anchor.Add(v)

	return d
}

func (d *DefaultObject) Scale(v float64) Object {
	d.scale = d.scale * v

	return d
}

func (d *DefaultObject) IsPointInside(v Vector) bool {
	return false
}

func (d *DefaultObject) CenterMatrix() ebiten.GeoM {
	return ebiten.GeoM{}
}

func (d *DefaultObject) Render(canvas *ebiten.Image, geoM ebiten.GeoM, colorScale color.Color) {
	if d.texture == nil {
		return
	}

	m := d.initialM
	m.Scale(d.scale, d.scale)
	m.Rotate(d.rotate)
	m.Translate(d.anchor[0], d.anchor[1])
	m.Concat(geoM)

	cm := colorm.ColorM{}
	cm.ScaleWithColor(colorScale)
	colorm.DrawImage(canvas, d.texture, cm, &colorm.DrawImageOptions{
		GeoM:   m,
		Filter: ebiten.FilterLinear,
	})
}
