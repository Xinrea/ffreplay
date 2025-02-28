package object

import (
	"fmt"

	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CircleObject struct {
	DefaultObject
	radius float64
}

// make sure CircleObject implements Object interface.
var _ Object = (*CircleObject)(nil)

func NewCircleObject(opt ObjectOption, pos Vector, radius float64) *CircleObject {
	hashStr := fmt.Sprintf("circle-%v-%v-%v-%v", opt.FillColor, opt.StrokeColor, opt.StrokeWidth, radius)
	w := radius*2 + opt.StrokeWidth
	h := radius*2 + opt.StrokeWidth
	initialM := ebiten.GeoM{}
	initialM.Translate(-w/2, -h/2)

	if cachedTexture, ok := objectTextureCache[hashStr]; ok {
		return &CircleObject{
			DefaultObject: DefaultObject{
				anchor:   pos,
				rotate:   0,
				scale:    1,
				texture:  cachedTexture,
				initialM: initialM,
			},
			radius: radius,
		}
	}

	texture := ebiten.NewImage(int(w), int(h))
	vector.DrawFilledCircle(texture, float32(w/2), float32(h/2), float32(radius), opt.FillColor, true)
	vector.StrokeCircle(
		texture,
		float32(w/2),
		float32(h/2),
		float32(radius),
		float32(opt.StrokeWidth),
		opt.StrokeColor, true)

	objectTextureCache[hashStr] = texture

	return &CircleObject{
		DefaultObject: DefaultObject{
			otype:    TypeCircle,
			anchor:   pos,
			rotate:   0,
			scale:    1,
			texture:  texture,
			initialM: initialM,
		},
		radius: radius,
	}
}

func (c *CircleObject) IsPointInside(v Vector) bool {
	return c.Position().Sub(v).Length() <= c.radius
}
