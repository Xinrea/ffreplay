package model

import (
	"math"
)

type VelocityData struct {
	AX, AY, X, Y float64
}

func (v *VelocityData) Clear() {
	v.X = 0
	v.Y = 0
}

func (v *VelocityData) Normalize(m float64) {
	if v.X != 0 && v.Y != 0 {
		v.X = v.X / math.Abs(v.X) * math.Sqrt(m)
		v.Y = v.Y / math.Abs(v.Y) * math.Sqrt(m)
	}
}
