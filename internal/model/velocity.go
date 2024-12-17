package model

import (
	"math"

	"github.com/Xinrea/ffreplay/util"
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

func (v *VelocityData) Next(maxVel float64, friction float64) {
	totalAccelY := v.AY
	if v.Y > 0 {
		totalAccelY -= friction
	}
	if v.Y < 0 {
		totalAccelY += friction
	}
	if v.Y == 0 {
		if v.AY < 0 {
			totalAccelY += friction
		}
		if v.AY > 0 {
			totalAccelY -= friction
		}
	}
	totalAccelX := v.AX
	if v.X > 0 {
		totalAccelX -= friction
	}
	if v.X < 0 {
		totalAccelX += friction
	}
	if v.X == 0 {
		if v.AX < 0 {
			totalAccelX += friction
		}
		if v.AX > 0 {
			totalAccelX -= friction
		}
	}
	v.X = util.RangeLimit(v.X+totalAccelX, maxVel)
	v.Y = util.RangeLimit(v.Y+totalAccelY, maxVel)
}
