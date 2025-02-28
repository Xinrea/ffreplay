package object

import (
	"fmt"
	"math"

	. "github.com/Xinrea/ffreplay/pkg/line"
	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type RectObject struct {
	DefaultObject
	width  float64
	height float64
}

// make sure RectObject implements Object interface.
var _ Object = (*RectObject)(nil)

const (
	AnchorBottomMiddle = iota
	AnchorLeftMiddle
	AnchorRightMiddle
	AnchorMiddle
)

// NewRectObject creates a new RectObject with given position and size.
func NewRectObject(opt ObjectOption, pos Vector, anchor int, rw, rh float64) *RectObject {
	hashStr := fmt.Sprintf("rect-%v-%v-%v-%v-%v", opt.FillColor, opt.StrokeColor, opt.StrokeWidth, rw, rh)
	w := rw + opt.StrokeWidth
	h := rh + opt.StrokeWidth
	initialM := ebiten.GeoM{}
	initialM.Rotate(-math.Pi)

	switch anchor {
	case AnchorMiddle:
		initialM.Translate(w/2, h/2)
	case AnchorBottomMiddle:
		initialM.Translate(w/2, 0)
	case AnchorLeftMiddle:
		initialM.Translate(w, h/2)
	case AnchorRightMiddle:
		initialM.Translate(0, h/2)
	default:
		initialM.Translate(w/2, 0)
	}

	if cachedTexture, ok := objectTextureCache[hashStr]; ok {
		return &RectObject{
			DefaultObject: DefaultObject{
				anchor:   pos,
				rotate:   0,
				scale:    1,
				texture:  cachedTexture,
				initialM: initialM,
			},
			width:  rw,
			height: rh,
		}
	}

	texture := ebiten.NewImage(int(w), int(h))
	texture.Fill(opt.FillColor)

	vector.StrokeRect(
		texture,
		float32(opt.StrokeWidth)/2,
		float32(opt.StrokeWidth)/2,
		float32(rw),
		float32(rh),
		float32(opt.StrokeWidth),
		opt.StrokeColor,
		true)

	objectTextureCache[hashStr] = texture

	return &RectObject{
		DefaultObject: DefaultObject{
			otype:    TypeRect,
			anchor:   pos,
			rotate:   0,
			scale:    1,
			texture:  texture,
			initialM: initialM,
		},
		width:  w,
		height: h,
	}
}

// lines are bounds of the rectangle.
//
// which are 4 lines:
// 1(x, y)     2(x+w, y)
// 4(x, y+h)   3(x+w, y+h).
func (r *RectObject) lines() [4]Line {
	return [4]Line{
		NewLine(Vector{0, 0}, Vector{r.width, 0}).Apply(r.initialM),
		NewLine(Vector{r.width, 0}, Vector{r.width, r.height}).Apply(r.initialM),
		NewLine(Vector{r.width, r.height}, Vector{0, r.height}).Apply(r.initialM),
		NewLine(Vector{0, r.height}, Vector{0, 0}).Apply(r.initialM),
	}
}

func (r *RectObject) IsPointInside(v Vector) bool {
	// make sure the line has at least one end outside the rect, so we make the end point to be very far away
	pRelative := v.Sub(r.anchor)
	testLine := NewLine(pRelative, Vector{pRelative[0] + 999999, pRelative[1]})

	cnt := 0

	for _, l := range r.lines() {
		l = l.Translate(r.anchor)
		l = l.Rotate(r.rotate)
		l = l.Scale(Vector{r.scale, r.scale})

		if l.IsIntersecting(testLine) {
			cnt++
		}
	}

	return cnt%2 == 1
}
