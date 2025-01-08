package line

import (
	. "github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
)

type Line struct {
	start, end Vector
}

func NewLine(start, end Vector) Line {
	return Line{
		start: start,
		end:   end,
	}
}

func (l Line) Apply(m ebiten.GeoM) Line {
	start := l.start.Apply(m)
	end := l.end.Apply(m)

	return NewLine(start, end)
}

func (l Line) Start() Vector {
	return l.start
}

func (l Line) End() Vector {
	return l.end
}

func (l Line) Length() float64 {
	return l.end.Sub(l.start).Length()
}

func (l Line) Rotate(r float64) Line {
	return Line{l.start.Rotate(r), l.end.Rotate(r)}
}

func (l *Line) Translate(v Vector) Line {
	return Line{l.start.Add(v), l.end.Add(v)}
}

func (l *Line) Scale(v Vector) Line {
	return Line{l.start.Mul(v), l.end.Mul(v)}
}

func (l Line) IsIntersecting(l2 Line) bool {
	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection
	x1, y1 := l.start[0], l.start[1]
	x2, y2 := l.end[0], l.end[1]
	x3, y3 := l2.start[0], l2.start[1]
	x4, y4 := l2.end[0], l2.end[1]

	denom := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if denom == 0 {
		return false
	}

	t := ((x1-x3)*(y3-y4) - (y1-y3)*(x3-x4)) / denom
	u := -((x1-x2)*(y1-y3) - (y1-y2)*(x1-x3)) / denom

	return t >= 0 && t <= 1 && u >= 0 && u <= 1
}
