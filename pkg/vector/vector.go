package vector

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/math/f64"
)

type Vector f64.Vec2

func NewVector(x, y float64) Vector {
	return Vector{x, y}
}

func (v Vector) Add(v2 Vector) Vector {
	return Vector{v[0] + v2[0], v[1] + v2[1]}
}

func (v Vector) Sub(v2 Vector) Vector {
	return Vector{v[0] - v2[0], v[1] - v2[1]}
}

func (v Vector) Mul(v2 Vector) Vector {
	return Vector{v[0] * v2[0], v[1] * v2[1]}
}

func (v Vector) Div(v2 Vector) Vector {
	return Vector{v[0] / v2[0], v[1] / v2[1]}
}

func (v Vector) Scale(s float64) Vector {
	return Vector{v[0] * s, v[1] * s}
}

func (v Vector) Rotate(r float64) Vector {
	cos := math.Cos(r)
	sin := math.Sin(r)
	x := v[0]*cos - v[1]*sin
	y := v[0]*sin + v[1]*cos

	return Vector{x, y}
}

func (v Vector) Length() float64 {
	return math.Hypot(v[0], v[1])
}

func (v *Vector) Normalize() {
	l := v.Length()
	v[0] /= l
	v[1] /= l
}

func (v Vector) Dot(v2 Vector) float64 {
	return v[0]*v2[0] + v[1]*v2[1]
}

func (v Vector) Cross(v2 Vector) float64 {
	return v[0]*v2[1] - v[1]*v2[0]
}

func (v Vector) Lerp(v2 Vector, t float64) Vector {
	x := v[0] + (v2[0]-v[0])*t
	y := v[1] + (v2[1]-v[1])*t

	return Vector{x, y}
}

func (v *Vector) Set(v2 Vector) {
	v[0] = v2[0]
	v[1] = v2[1]
}

// Angle returns the angle between two vectors in radians.
func (v Vector) Angle(v2 Vector) float64 {
	return math.Atan2(v[1], v[0]) - math.Atan2(v2[1], v2[0])
}

// Radian returns the angle of the vector in radians, relative to the north direction.
func (v Vector) Radian() float64 {
	return v.Angle(Vector{0, -1})
}

func (v Vector) Distance(v2 Vector) float64 {
	return math.Hypot(v2[0]-v[0], v2[1]-v[1])
}

func (v Vector) Apply(matrix ebiten.GeoM) Vector {
	x, y := matrix.Apply(v[0], v[1])

	return Vector{x, y}
}
