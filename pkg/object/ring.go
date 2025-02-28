package object

import (
	"fmt"
	"image/color"

	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type RingObject struct {
	DefaultObject
	innerRadius float64
	outerRadius float64
}

// make sure RingObject implements Object interface.
var _ Object = (*RingObject)(nil)

func NewRingObject(opt ObjectOption, pos Vector, innerRadius float64, outerRadius float64) *RingObject {
	hashStr := fmt.Sprintf(
		"ring-%v-%v-%v-%v-%v", opt.FillColor, opt.StrokeColor, opt.StrokeWidth, innerRadius, outerRadius)
	w := outerRadius*2 + opt.StrokeWidth
	h := outerRadius*2 + opt.StrokeWidth
	initialM := ebiten.GeoM{}
	initialM.Translate(-w/2, -h/2)

	if cachedTexture, ok := objectTextureCache[hashStr]; ok {
		return &RingObject{
			DefaultObject: DefaultObject{
				anchor:   pos,
				rotate:   0,
				scale:    1,
				texture:  cachedTexture,
				initialM: initialM,
			},
			innerRadius: innerRadius,
			outerRadius: outerRadius,
		}
	}

	texture := ebiten.NewImage(int(w), int(h))
	vector.DrawFilledCircle(texture, float32(w/2), float32(h/2), float32(outerRadius), opt.FillColor, true)
	vector.StrokeCircle(
		texture,
		float32(w/2),
		float32(h/2),
		float32(outerRadius),
		float32(opt.StrokeWidth),
		opt.StrokeColor, true)

	mask := ebiten.NewImage(int(w), int(h))
	vector.DrawFilledCircle(mask, float32(w/2), float32(h/2), float32(innerRadius), color.White, true)

	texture.DrawImage(mask, &ebiten.DrawImageOptions{
		Blend: ebiten.BlendDestinationOut,
	})

	vector.StrokeCircle(
		texture,
		float32(w/2),
		float32(h/2),
		float32(innerRadius),
		float32(opt.StrokeWidth),
		opt.StrokeColor, true)

	objectTextureCache[hashStr] = texture

	return &RingObject{
		DefaultObject: DefaultObject{
			otype:    TypeRing,
			anchor:   pos,
			rotate:   0,
			scale:    1,
			texture:  texture,
			initialM: initialM,
		},
		innerRadius: innerRadius,
		outerRadius: outerRadius,
	}
}

func (r *RingObject) IsPointInside(v Vector) bool {
	return r.Position().Sub(v).Length() <= r.outerRadius && r.Position().Sub(v).Length() >= r.innerRadius
}
