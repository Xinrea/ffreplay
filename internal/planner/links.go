package planner

import (
	"errors"
	"iter"
	"math"
	"slices"

	"github.com/Xinrea/ffreplay/internal/gameboard"
)

// An unattached end (ID zero) is a fixed point. The game stores only fixed
// points; never infer attachments merely because two objects overlap.
type Endpoint struct {
	ID int     `json:"id,omitempty"`
	X  float64 `json:"x,omitempty"`
	Y  float64 `json:"y,omitempty"`
}
type Link struct {
	Start Endpoint          `json:"start"`
	End   Endpoint          `json:"end"`
	Style string            `json:"style"`
	Color int               `json:"color"`
	Width float64           `json:"width,omitempty"` // metres; zero = editor default
	Above int               `json:"above,omitempty"` // number of object layers below this line
	Game  *gameboard.Object `json:"game,omitempty"`
}

func (l Link) Stroke() float64 {
	if l.Width > 0 {
		return l.Width
	}
	return .25
}
func (l Link) Recolored() bool { return l.Game != nil && nearestInk(l.Game.RGBA) != l.Color }

var LinkStyles = []string{"line", "close", "far", "--", "+-", "++"}

// Layers yields either an object index or a line index, never both.
// Links are sorted by Above, including when imported from the game's mixed stack.
func (s Step) Layers() iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		j := 0
		for i := 0; i <= len(s.Objects); i++ {
			for j < len(s.Links) && s.Links[j].Above == i {
				if !yield(-1, j) {
					return
				}
				j++
			}
			if i < len(s.Objects) && !yield(i, -1) {
				return
			}
		}
	}
}

func (s Step) Position(p Endpoint) (x, y float64, ok bool) {
	if p.ID == 0 {
		return p.X, p.Y, true
	}
	for _, o := range s.Objects {
		if o.ID == p.ID {
			return o.X, o.Y, true
		}
	}
	return 0, 0, false
}

func (s Step) validateLinks() error {
	if len(s.Links) > MaxObjects || s.GameBackground > 7 {
		return errors.New("连线数量或游戏背景无效")
	}
	ids := map[int]bool{}
	for _, o := range s.Objects {
		if o.ID == 0 {
			continue
		}
		if o.ID < 0 || o.ID > MaxObjects || ids[o.ID] {
			return errors.New("连线锚点 ID 无效")
		}
		ids[o.ID] = true
	}
	for i, l := range s.Links {
		if l.Above < 0 || l.Above > len(s.Objects) || i > 0 && l.Above < s.Links[i-1].Above {
			return errors.New("连线图层无效")
		}
		if l.Game != nil && l.Game.Type != 12 {
			return errors.New("无效游戏连线")
		}
		if !slices.Contains(LinkStyles, l.Style) || l.Color < 0 || l.Color > 3 || l.Width < 0 || l.Width > Extent {
			return errors.New("连线样式无效")
		}
		for _, p := range []Endpoint{l.Start, l.End} {
			if p.ID != 0 && !ids[p.ID] || math.Abs(p.X) > Extent || math.Abs(p.Y) > Extent {
				return errors.New("连线端点不存在或越界")
			}
		}
	}
	return nil
}

func (s *Step) anchor(index int, point Endpoint) Endpoint {
	if index < 0 {
		return point
	}
	o := &s.Objects[index]
	if o.ID == 0 {
		used := map[int]bool{}
		for _, other := range s.Objects {
			used[other.ID] = true
		}
		for id := 1; id <= MaxObjects; id++ {
			if !used[id] {
				o.ID = id
				break
			}
		}
	}
	return Endpoint{ID: o.ID}
}

func (e *Editor) Connect(a, b int, start, end Endpoint, style string) int {
	s := e.Step()
	if a < -1 || b < -1 || a >= len(s.Objects) || b >= len(s.Objects) || a >= 0 && a == b ||
		len(s.Links) >= MaxObjects || !slices.Contains(LinkStyles, style) {
		return -1
	}
	if a < 0 && b < 0 && start == end {
		return -1
	}
	color := 0
	if style == "far" {
		color = 2
	} else if style == "close" {
		color = 3
	}
	e.Checkpoint()
	i := 0
	for i < len(s.Links) && s.Links[i].Above == 0 {
		i++
	}
	s.Links = slices.Insert(s.Links, i, Link{Start: s.anchor(a, start), End: s.anchor(b, end), Style: style, Color: color})
	return i
}

func (s *Step) RemoveObject(i int) {
	if id := s.Objects[i].ID; id != 0 {
		s.Links = slices.DeleteFunc(s.Links, func(l Link) bool { return l.Start.ID == id || l.End.ID == id })
	}
	for j := range s.Links {
		if i < s.Links[j].Above {
			s.Links[j].Above--
		}
	}
	s.Objects = slices.Delete(s.Objects, i, i+1)
}
