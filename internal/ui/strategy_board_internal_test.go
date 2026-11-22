package ui

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/Xinrea/ffreplay/internal/planner"
)

func TestBoardLayerDragAndImport(t *testing.T) {
	const importedName = "分享的解法"

	b := StrategyBoard{
		editor:   planner.Editor{Plan: planner.New()},
		selected: map[int]bool{},
		floors:   map[string]string{"": ""},
	}
	for _, kind := range []string{"MT", "ST", "H1", "H2"} {
		b.editor.Step().Objects = append(
			b.editor.Step().Objects,
			planner.Object{Kind: kind, Label: kind, Radius: 1},
		)
	}

	b.moveLayer(0, 2)

	for i, k := range []string{"ST", "H1", "MT", "H2"} {
		if b.editor.Step().Objects[i].Kind != k {
			t.Fatal("layer move must preserve relative order")
		}
	}

	if !b.selected[2] {
		t.Fatal("selection did not move with layer")
	}

	b.selectObject(3, true)
	b.removeSelected()

	if len(b.editor.Step().Objects) != 2 {
		t.Fatal("multi-delete")
	}

	b.history(false)

	if len(b.editor.Step().Objects) != 4 || len(b.selected) != 0 {
		t.Fatal("undo multi-delete")
	}

	checkBoardImport(t, &b, importedName)
}

func checkBoardImport(t *testing.T, b *StrategyBoard, name string) {
	t.Helper()

	importPlan := planner.New()
	importPlan.Steps[0].Name = name
	data, _ := json.Marshal(importPlan)
	b.importDrop(fstest.MapFS{"test.ffplan": &fstest.MapFile{Data: data}})

	if b.editor.Step().Name != name {
		t.Fatal(b.status)
	}

	b.importDrop(fstest.MapFS{"bad.ffplan": &fstest.MapFile{Data: []byte(`not json`)}})

	if b.editor.Step().Name != name {
		t.Fatal("failed import replaced current plan")
	}

	b.history(false)

	if len(b.editor.Step().Objects) != 4 {
		t.Fatal("undo successful import")
	}

	importPlan.Steps[0].Floor = "not-in-catalogue"

	data, _ = json.Marshal(importPlan)

	_, err := b.decode(data)
	if err == nil {
		t.Fatal("accepted unknown floor")
	}
}

func TestDragIsOneEdit(t *testing.T) {
	b := StrategyBoard{editor: planner.Editor{Plan: planner.New()}}
	o := planner.Object{Kind: planner.KindCircle, Radius: 3, X: 2, Y: 4}
	b.editor.Step().Objects = []planner.Object{o}
	b.drag = &boardDrag{x: 3, y: 4, mode: dragMove, start: map[int]planner.Object{0: o}}
	b.moveDrag(5, 6)
	b.moveDrag(8, 8)
	b.endDrag(true)

	got := b.editor.Step().Objects[0]
	if got.X != 7 || got.Y != 8 {
		t.Fatal("lost grab offset", got)
	}

	if !b.editor.History(false) || !reflect.DeepEqual(b.editor.Step().Objects[0], o) ||
		b.editor.History(false) {
		t.Fatal("drag history")
	}
}

func TestFormationAndCancelledDrag(t *testing.T) {
	b := StrategyBoard{editor: planner.Editor{Plan: planner.New()}, snap: true}
	s := b.editor.Step()
	s.Objects = []planner.Object{{Kind: "MT", X: 1.4, Y: -1.2}, {Kind: "ST", X: 18.6, Y: 3.3}}
	start := func() {
		b.drag = &boardDrag{
			mode:  dragMove,
			start: map[int]planner.Object{0: s.Objects[0], 1: s.Objects[1]},
		}
	}
	initial := s.Objects[0]

	start()
	b.moveDrag(12, 2)

	if s.Objects[1].X != 20 || math.Abs(s.Objects[0].X-2.8) > 1e-9 ||
		math.Abs(s.Objects[0].Y-1) > 1e-9 ||
		math.Abs(s.Objects[1].Y-5.5) > 1e-9 {
		t.Fatal(s.Objects)
	}

	b.endDrag(false)

	if b.editor.CanUndo() || !reflect.DeepEqual(s.Objects[0], initial) {
		t.Fatal("cancel changed history or position")
	}

	start()
	b.moveDrag(3, 2)
	b.moveDrag(0, 0)
	b.endDrag(true)

	if b.editor.CanUndo() {
		t.Fatal("return to origin must be a no-op")
	}
}

func TestToolPlacementAndStepWindow(t *testing.T) {
	b := StrategyBoard{pending: planner.KindCone, snap: true}

	o, ok := b.placement(.7, -2.3)
	if !ok || o.X != 1 || o.Y != -2 {
		t.Fatal(o)
	}

	if _, ok = b.placement(21, 0); ok {
		t.Fatal("outside canvas bounds")
	}
}

func TestStepWindow(t *testing.T) {
	for _, total := range []int{1, 4, planner.MaxSteps} {
		for i := range total {
			start, end := stepWindow(i, total, 4)
			if start < 0 || end > total || end-start > 4 || i < start || i >= end {
				t.Fatal(start, end, i, total)
			}
		}
	}
}

func TestOverlapWrapsBelowSelected(t *testing.T) {
	objects := []planner.Object{
		{Kind: "MT", Radius: 2},
		{Kind: planner.KindDonut, Radius: 1},
		{Kind: planner.KindCircle, Radius: 4},
	}
	for _, tc := range []struct{ first, want int }{{2, 2}, {1, 0}, {-1, 2}} {
		if got := pickLayer(objects, 0, 0, tc.first); got != tc.want {
			t.Errorf("first %d got %d", tc.first, got)
		}
	}
}
