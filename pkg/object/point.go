package object

import . "github.com/Xinrea/ffreplay/pkg/vector"

// PointObject is a simple object that has only position, used for player and enemy position.
// It implements Object interface, but it does not have any effect on RotateBy and Scale.
type PointObject struct {
	DefaultObject
}

// make sure PointObject implements Object interface.
var _ Object = (*PointObject)(nil)

func NewPointObject(pos Vector) *PointObject {
	return &PointObject{
		DefaultObject{
			anchor: pos,
		},
	}
}

func (p *PointObject) IsPointInside(v Vector) bool {
	return p.anchor == v
}
