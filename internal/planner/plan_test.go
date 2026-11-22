package planner_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Xinrea/ffreplay/internal/planner"
)

func TestStepIndependenceAndHistory(t *testing.T) {
	e := planner.Editor{Plan: planner.New()}

	e.Step().Objects = []planner.Object{{Kind: "MT", Label: "主坦", Radius: 1}}

	if e.DeleteStep() {
		t.Fatal("deleted only step")
	}

	e.AddStep(true)
	e.Checkpoint()

	e.Step().Objects[0].X = 5

	if e.Steps[0].Objects[0].X != 0 {
		t.Fatal("copy aliases the first step")
	}

	checkStepHistory(t, &e)
}

func checkStepHistory(t *testing.T, e *planner.Editor) {
	t.Helper()

	if !e.History(false) || e.Step().Objects[0].X != 0 {
		t.Fatal("undo move")
	}

	if !e.History(false) || len(e.Steps) != 1 || e.Current != 0 {
		t.Fatal("undo duplicate")
	}

	if !e.History(true) || !e.History(true) || e.Step().Objects[0].X != 5 {
		t.Fatal("redo lost edit")
	}

	checkStepBranch(t, e)
}

func checkStepBranch(t *testing.T, e *planner.Editor) {
	t.Helper()
	e.History(false)
	e.AddStep(false)

	if len(e.Step().Objects) != 0 || e.History(true) {
		t.Fatal("new branch must be empty and discard redo")
	}

	e.Current = 0
	e.DeleteStep()

	if !e.History(false) || e.Current != 0 || e.Step().Objects[0].Label != "主坦" {
		t.Fatal("restore deleted first step")
	}
}

func TestPlanRoundTrip(t *testing.T) {
	p := planner.New()
	p.Square = true

	for _, group := range planner.Palette() {
		for _, tool := range group {
			p.Steps[0].Objects = append(p.Steps[0].Objects, planner.Object{
				Kind:   tool.Key,
				Label:  tool.Label,
				X:      -4.5,
				Y:      3,
				Angle:  90,
				Radius: tool.Radius,
				Color:  2,
			})
		}
	}

	p.Steps = append(p.Steps, planner.Step{Name: "二运", Floor: "m4s"})

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	got, err := planner.Decode(b)
	if err != nil || !reflect.DeepEqual(p, got) {
		t.Fatalf("round trip: %v, %#v", err, got)
	}
}

func TestRevisionFollowsHistoryNotNavigation(t *testing.T) {
	e := planner.Editor{Plan: planner.New()}
	e.AddStep(true)
	saved := e.Revision()

	e.Current = 0

	if e.Revision() != saved {
		t.Fatal("navigation marked dirty")
	}

	e.AddStep(false)
	branch := e.Revision()
	e.History(false)

	if e.Revision() != saved || !e.CanRedo() {
		t.Fatal("undo must return to saved state")
	}

	e.AddStep(true)

	if e.Revision() == branch || e.CanRedo() {
		t.Fatal("new branch reused revision")
	}
}

func TestFillMissingParty(t *testing.T) {
	s := planner.Step{Objects: []planner.Object{{Kind: "MT", Label: "existing", X: 5}}}

	missing := s.MissingParty()
	if len(missing) != 7 {
		t.Fatal(missing)
	}

	s.Objects = append(s.Objects, missing...)
	if len(s.MissingParty()) != 0 || s.Objects[0].X != 5 {
		t.Fatal("already placed slots changed")
	}
}

func TestInvalidPlan(t *testing.T) {
	for _, raw := range []string{
		``, `{"version":2}`, `{"version":1,"steps":[]}`, `{"version":1,"steps":[null]}`,
		strings.Repeat(" ", planner.MaxFileSize+1),
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"unregistered","label":"?","radius":1}]}]}`,
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"donut","label":"?","radius":1,"shape":3}]}]}`,
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"MT","label":"?","radius":1,"color":4}]}]}`,
	} {
		_, err := planner.Decode([]byte(raw))
		if err == nil {
			t.Fatal("accepted invalid document")
		}
	}
}

func TestBossDefaultLabel(t *testing.T) {
	if planner.Tools[planner.KindBoss].Label != "Boss" {
		t.Fatal("palette should use Boss")
	}

	p := planner.New()
	for _, label := range []string{"首领", "绝望之龙"} {
		p.Steps[0].Objects = append(
			p.Steps[0].Objects,
			planner.Object{Kind: planner.KindBoss, Label: label, Radius: 3},
		)
	}

	raw, _ := json.Marshal(p)

	got, err := planner.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}

	objects := got.Steps[0].Objects
	if objects[0].Label != "Boss" || objects[1].Label != "绝望之龙" {
		t.Fatal(objects)
	}
}

func TestShapePicking(t *testing.T) {
	for _, tc := range []struct {
		kind               string
		angle, shape, x, y float64
		hit                bool
	}{
		{planner.KindDonut, 0, 0, 0, 0, false},
		{planner.KindDonut, 0, .8, .7, 0, false},
		{planner.KindDonut, 0, .5, .8, 0, true},
		{planner.KindCone, 90, 0, .8, 0, true},
		{planner.KindCone, 90, 0, -.8, 0, false},
		{planner.KindCone, 0, 240, .8, 0, true},
		{planner.KindRect, 90, .1, .5, 0, true},
		{planner.KindRect, 90, .1, 0, .5, false},
	} {
		o := planner.Object{Kind: tc.kind, Radius: 1, Angle: tc.angle, Shape: tc.shape}
		if o.Hit(tc.x, tc.y) != tc.hit {
			t.Errorf("%+v", tc)
		}
	}
}
