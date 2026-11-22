package planner

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestStepIndependenceAndHistory(t *testing.T) {
	e := Editor{Plan: New()}
	e.Step().Objects = []Object{{Kind: "MT", Label: "主坦", Radius: 1}}
	if e.DeleteStep() {
		t.Fatal("deleted only step")
	}
	e.AddStep(true)
	e.Checkpoint()
	e.Step().Objects[0].X = 5
	if e.Steps[0].Objects[0].X != 0 {
		t.Fatal("copy aliases the first step")
	}
	if !e.History(false) || e.Step().Objects[0].X != 0 {
		t.Fatal("undo move")
	}
	if !e.History(false) || len(e.Steps) != 1 || e.Current != 0 {
		t.Fatal("undo duplicate")
	}
	if !e.History(true) || !e.History(true) || e.Step().Objects[0].X != 5 {
		t.Fatal("redo lost edit")
	}
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
	p := New()
	p.Square = true
	for _, group := range Palette() {
		for _, tool := range group {
			p.Steps[0].Objects = append(p.Steps[0].Objects, Object{
				Kind: tool.Key, Label: tool.Label, X: -4.5, Y: 3, Angle: 90, Radius: tool.Radius, Color: 2,
			})
		}
	}
	p.Steps = append(p.Steps, Step{Name: "二运", Floor: "m4s"})
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(b)
	if err != nil || !reflect.DeepEqual(p, got) {
		t.Fatalf("round trip: %v, %#v", err, got)
	}
}

func TestRevisionFollowsHistoryNotNavigation(t *testing.T) {
	e := Editor{Plan: New()}
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
	s := Step{Objects: []Object{{Kind: "MT", Label: "existing", X: 5}}}
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
	for _, raw := range []string{``, `{"version":2}`, `{"version":1,"steps":[]}`, `{"version":1,"steps":[null]}`,
		strings.Repeat(" ", MaxFileSize+1),
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"unregistered","label":"?","radius":1}]}]}`,
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"donut","label":"?","radius":1,"shape":3}]}]}`,
		`{"version":1,"steps":[{"name":"test","objects":[{"kind":"MT","label":"?","radius":1,"color":4}]}]}`,
	} {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Fatal("accepted invalid document")
		}
	}
}

func TestBossDefaultLabel(t *testing.T) {
	if Tools["boss"].Label != "Boss" {
		t.Fatal("palette should use Boss")
	}
	p := New()
	for _, label := range []string{"首领", "绝望之龙"} {
		p.Steps[0].Objects = append(p.Steps[0].Objects, Object{Kind: "boss", Label: label, Radius: 3})
	}
	raw, _ := json.Marshal(p)
	got, err := Decode(raw)
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
		{"donut", 0, 0, 0, 0, false}, {"donut", 0, .8, .7, 0, false}, {"donut", 0, .5, .8, 0, true},
		{"cone", 90, 0, .8, 0, true}, {"cone", 90, 0, -.8, 0, false}, {"cone", 0, 240, .8, 0, true},
		{"rect", 90, .1, .5, 0, true}, {"rect", 90, .1, 0, .5, false},
	} {
		o := Object{Kind: tc.kind, Radius: 1, Angle: tc.angle, Shape: tc.shape}
		if o.Hit(tc.x, tc.y) != tc.hit {
			t.Errorf("%+v", tc)
		}
	}
}
