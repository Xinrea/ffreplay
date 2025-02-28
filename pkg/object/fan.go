package object

import (
	"fmt"
	"math"

	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type FanObject struct {
	DefaultObject
	angle  float64
	length float64
}

// make sure FanObject implements Object interface.
var _ Object = (*FanObject)(nil)

func NewFanObject(opt ObjectOption, pos Vector, angle, length float64) *FanObject {
	hashStr := fmt.Sprintf("fan-%v-%v-%v-%v-%v", opt.FillColor, opt.StrokeColor, opt.StrokeWidth, angle, length)
	// always towards to the north
	initialM := ebiten.GeoM{}
	initialM.Translate(-length, -length)
	initialM.Rotate(-math.Pi / 2)

	if cachedTexture, ok := objectTextureCache[hashStr]; ok {
		return &FanObject{
			DefaultObject: DefaultObject{
				anchor:   pos,
				rotate:   0,
				scale:    1,
				texture:  cachedTexture,
				initialM: initialM,
			},
			angle:  angle,
			length: length,
		}
	}
	// ctx draw and initialM is only used for texture rendering
	// object logic is not related to them
	rad := gg.Radians(angle / 2)
	ctx := gg.NewContext(int(2*length), int(2*length))
	ctx.MoveTo(length, length)
	ctx.DrawArc(length, length, length, -rad, rad)
	ctx.LineTo(length, length)
	ctx.ClosePath()
	ctx.SetColor(opt.FillColor)
	ctx.FillPreserve()
	ctx.SetColor(opt.StrokeColor)
	ctx.SetLineWidth(opt.StrokeWidth)
	ctx.Stroke()

	texture := ebiten.NewImageFromImage(ctx.Image())
	objectTextureCache[hashStr] = texture

	return &FanObject{
		DefaultObject: DefaultObject{
			otype:    TypeFan,
			anchor:   pos,
			rotate:   0,
			scale:    1,
			texture:  texture,
			initialM: initialM,
		},
		angle:  angle,
		length: length,
	}
}

func (f *FanObject) IsPointInside(v Vector) bool {
	pRelative := v.Sub(f.anchor)
	centerLine := NewVector(0, -1).Rotate(f.rotate)
	relativeAngle := pRelative.Angle(centerLine)
	half := math.Pi * f.angle / 180 / 2

	return pRelative.Length() <= f.length && relativeAngle <= half && relativeAngle >= -half
}
