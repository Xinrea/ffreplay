// Package planner models static strategy diagrams, independently of the replay clock and ECS.
package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/Xinrea/ffreplay/internal/gameboard"
	"github.com/Xinrea/ffreplay/internal/model/role"
)

const (
	// MaxFileSize bounds the encoded plan, including JSON-escaped labels.
	MaxFileSize = 6 << 20
	MaxSteps    = 64
	MaxObjects  = 128
	Extent      = 20 // arena half-width, in metres
)

// Stable tool keys used in plans, wire mappings, and the editor.
const (
	KindCircle = "circle"
	KindDonut  = "donut"
	KindCone   = "cone"
	KindRect   = "rect"
	KindArrow  = "arrow"
	KindText   = "text"
	KindBoss   = "boss"
)

type Tool struct {
	Key, Label, Icon string
	Radius           float64
}

func Palette() [][]Tool {
	var party, markers []Tool

	options := role.PlayableJobOptions()
	jobs := make([]Tool, 0, len(options))

	for r := role.MT; r <= role.D4; r++ {
		name := r.String()
		party = append(party, Tool{name, name, "asset/role/" + strings.ToLower(name) + ".png", 1.2})
	}

	for _, job := range options {
		name := job.Role.String()
		jobs = append(jobs, Tool{name, name, "asset/role/" + name + ".png", 1.2})
	}

	for key := range strings.SplitSeq("ABCD1234", "") {
		markers = append(markers, Tool{key, key, "asset/marker/worldmarker" + key + ".png", 1.6})
	}

	mechanics := []Tool{
		{KindCircle, "圆形", "", 5},
		{KindDonut, "环形", "", 6},
		{KindCone, "扇形", "", 7},
		{KindRect, "直线", "", 7},
		{KindArrow, "箭头", "", 4},
		{KindText, "文字", "", 2},
	}

	return [][]Tool{
		party,
		markers,
		slices.Concat(mechanics, extraIcons("机制")),
		jobs,
		extraIcons("指示器"),
		append([]Tool{{KindBoss, "Boss", "asset/target_enemy.png", 3}}, extraIcons("敌方")...),
		extraIcons("职能"),
	}
}

var Tools = func() map[string]Tool {
	m := map[string]Tool{}

	for _, group := range Palette() {
		for _, t := range group {
			m[t.Key] = t
		}
	}

	return m
}()

type Object struct {
	ID     int               `json:"id,omitempty"`   // step-local tether anchor; zero for unconnected legacy tokens
	Game   *gameboard.Object `json:"game,omitempty"` // original game attributes; immutable while editing
	Kind   string            `json:"kind"`
	Label  string            `json:"label"`
	X      float64           `json:"x"`
	Y      float64           `json:"y"`
	Radius float64           `json:"radius"`
	Angle  float64           `json:"angle"`           // clockwise degrees; zero points north
	Color  int               `json:"color"`           // one of four diagram inks
	Shape  float64           `json:"shape,omitempty"` // aperture (degrees) or inner/width ratio; zero uses tool default
	Buffs  []int             `json:"buffs,omitempty"` // ordered status icon IDs above a party/job token
}

// Step stores objects in paint order. The last object is picked first.
type Step struct {
	Name           string   `json:"name"`
	Floor          string   `json:"floor,omitempty"` // optional built-in floor catalogue entry
	Objects        []Object `json:"objects"`
	Links          []Link   `json:"links,omitempty"`
	GameBackground uint16   `json:"gameBackground,omitempty"`
}

type Plan struct {
	Version int    `json:"version"`
	Square  bool   `json:"square"`
	Grid    bool   `json:"grid"`
	Steps   []Step `json:"steps"`
}

func New() Plan {
	return Plan{Version: 1, Grid: true, Steps: []Step{{Name: "起始站位"}}}
}

func Decode(b []byte) (Plan, error) {
	var p Plan
	if len(b) > MaxFileSize {
		return p, fmt.Errorf("文件超过 %d MiB", MaxFileSize>>20)
	}

	err := json.Unmarshal(b, &p)
	if err != nil {
		return Plan{}, errors.New("不是有效的 JSON 方案")
	}

	if p.Version != 1 || len(p.Steps) == 0 || len(p.Steps) > MaxSteps {
		return Plan{}, errors.New("不支持的方案版本或步骤数量")
	}

	for i, s := range p.Steps {
		err := s.validate()
		if err != nil {
			return Plan{}, fmt.Errorf("步骤 %d：%w", i+1, err)
		}
	}

	return p, nil
}

// MissingParty supplies only missing generic raid slots at clock positions.
// Existing positions, names and paint order are untouched.
func (s Step) MissingParty() []Object {
	seen := map[string]bool{}
	for _, o := range s.Objects {
		seen[o.Kind] = true
	}

	var missing []Object

	for i, key := range []string{"MT", "D4", "H2", "D2", "ST", "D1", "H1", "D3"} {
		if seen[key] {
			continue
		}

		sin, cos := math.Sincos(float64(i) * math.Pi / 4)
		t := Tools[key]
		missing = append(missing, Object{
			Kind: key, Label: t.Label, Radius: t.Radius,
			X: math.Round(100*sin) / 10, Y: -math.Round(100*cos) / 10,
		})
	}

	return missing
}

func (s Step) validate() error {
	err := s.validateLinks()
	if err != nil {
		return err
	}

	if len(s.Objects) > MaxObjects || s.Name == "" || len([]rune(s.Name)) > 60 ||
		len(s.Floor) > 128 {
		return errors.New("步骤超出限制")
	}

	for j, o := range s.Objects {
		// Upgrade only the former built-in name in saved drafts/files/links.
		if o.Kind == KindBoss && o.Label == "首领" {
			s.Objects[j].Label = Tools[o.Kind].Label
		}

		err := o.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s Step) clone() Step {
	s.Links = slices.Clone(s.Links)

	s.Objects = slices.Clone(s.Objects)

	for i := range s.Objects {
		s.Objects[i].Buffs = slices.Clone(s.Objects[i].Buffs)
	}

	return s
}

type snapshot struct {
	Plan

	Current  int
	revision uint64
}

// Editor keeps navigation out of the undo stack. Checkpoint once per committed edit
// (for a pointer drag, immediately before the first move, not on every tick).
type Editor struct {
	Plan

	Current          int
	undo, redo       []snapshot
	revision, serial uint64
}

func (e *Editor) Step() *Step { return &e.Steps[e.Current] }

// Revision identifies a document state, not a navigation or selection change.
// Undo restores its ID; editing a new branch always gets a fresh one.
func (e *Editor) Revision() uint64 { return e.revision }
func (e *Editor) CanUndo() bool    { return len(e.undo) > 0 }
func (e *Editor) CanRedo() bool    { return len(e.redo) > 0 }

func (e *Editor) Checkpoint() {
	if len(e.undo) == 32 {
		e.undo = slices.Delete(e.undo, 0, 1)
	}

	e.undo = append(e.undo, e.snapshot())
	e.redo = nil
	e.serial++
	e.revision = e.serial
}

func (e *Editor) History(redo bool) bool {
	from, to := &e.undo, &e.redo
	if redo {
		from, to = to, from
	}

	if len(*from) == 0 {
		return false
	}

	*to = append(*to, e.snapshot())
	last := len(*from) - 1
	s := (*from)[last]
	*from = slices.Delete(*from, last, last+1)
	e.Plan, e.Current, e.revision = s.Plan, s.Current, s.revision

	return true
}

func (e *Editor) AddStep(copyCurrent bool) bool {
	if len(e.Steps) >= MaxSteps {
		return false
	}

	e.Checkpoint()

	var s Step
	if copyCurrent {
		s = e.Step().clone()
	}

	e.Current++
	s.Name = fmt.Sprintf("步骤 %d", e.Current+1)
	e.Steps = slices.Insert(e.Steps, e.Current, s)

	return true
}

func (e *Editor) DeleteStep() bool {
	if len(e.Steps) == 1 {
		return false
	}

	e.Checkpoint()
	e.Steps = slices.Delete(e.Steps, e.Current, e.Current+1)
	e.Current = min(e.Current, len(e.Steps)-1)

	return true
}

func (e *Editor) snapshot() snapshot {
	p := e.Plan

	p.Steps = slices.Clone(p.Steps)

	for i := range p.Steps {
		p.Steps[i] = p.Steps[i].clone()
	}

	return snapshot{p, e.Current, e.revision}
}

func Clamp(x, low, high float64) float64 { return min(max(x, low), high) }

func (o Object) ShapeRange() (low, high, initial float64) {
	switch o.Kind {
	case KindCone:
		return 1, 360, 90
	case KindDonut:
		return .1, .9, .5
	case KindRect:
		return .05, 1, .3
	default:
		return 0, 0, 0
	}
}

func (o Object) ShapeValue() float64 {
	_, _, initial := o.ShapeRange()
	if o.Shape == 0 {
		return initial
	}

	return o.Shape
}

// Aperture also handles annular sectors imported as game donuts.
func (o Object) Aperture() float64 {
	if o.Kind == KindCone {
		return o.ShapeValue()
	}

	if o.Kind == KindDonut && o.Game != nil {
		if a := o.Game.Params[0]; a > 0 && a < 360 {
			return float64(a)
		}
	}

	return 360
}

// Local returns the point in unrotated, unit-radius object coordinates.
func (o Object) Local(x, y float64) (float64, float64) {
	s, c := math.Sincos(o.Angle * math.Pi / 180)
	dx, dy := (x-o.X)/o.Radius, (y-o.Y)/o.Radius

	return dx*c + dy*s, -dx*s + dy*c
}

func (o Object) Hit(x, y float64) bool {
	if o.Game != nil && o.Game.Flags&1 == 0 {
		return false
	}

	dx, dy := o.Local(x, y)

	switch o.Kind {
	case KindCircle, KindCone, KindDonut:
		return o.hitRadial(dx, dy)
	case KindRect:
		return inBox(dx, dy, o.ShapeValue(), 1)
	case KindArrow:
		return inBox(dx, dy, .25, 1)
	case KindText:
		return inBox(dx, dy, 1, .5)
	default:
		return inBox(dx, dy, 1, 1)
	}
}

func inBox(x, y, w, h float64) bool { return math.Abs(x) <= w && math.Abs(y) <= h }
func (o Object) hitRadial(x, y float64) bool {
	r := math.Hypot(x, y)
	if r > 1 || o.Kind == KindDonut && r < o.ShapeValue() {
		return false
	}

	return r == 0 || math.Abs(math.Atan2(x, -y))*180/math.Pi <= o.Aperture()/2
}

func (o Object) validGame() bool {
	return o.Game == nil || (gameKind(o.Game.Type) == o.Kind && len(o.Game.Text) <= 512)
}

func (o Object) validate() error {
	if !o.validGame() {
		return errors.New("游戏对象无效")
	}

	if _, ok := Tools[o.Kind]; !ok || o.Label == "" || len([]rune(o.Label)) > 60 {
		return errors.New("对象类型或名称无效")
	}

	if !o.validGeometry() || o.Color < 0 || o.Color >= 4 {
		return errors.New("对象属性无效")
	}

	if !o.validBuffs() {
		return errors.New("头顶 Buff 无效")
	}

	return nil
}

func (o Object) validGeometry() bool {
	lo, hi, _ := o.ShapeRange()

	return inBox(o.X, o.Y, Extent, Extent) && o.Radius >= .5 && o.Radius <= Extent &&
		math.Abs(o.Angle) <= 360 &&
		(o.Shape == 0 || o.Shape >= lo && o.Shape <= hi)
}
