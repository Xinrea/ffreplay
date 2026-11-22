package planner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

func TestStatusIconCatalogue(t *testing.T) {
	files, err := filepath.Glob("../../asset/abilities/2?????-??????.png")
	if err != nil || len(files) == 0 || len(files) != len(FindBuffIcons("")) {
		t.Fatalf("catalogue vs artwork: %d/%d (%v)", len(buffIcons), len(files), err)
	}
	for _, icon := range buffIcons {
		if _, err := os.Stat(filepath.Join("../../asset/abilities", icon.Asset)); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"铁壁", "213911"} {
		matches := FindBuffIcons(query)
		if !slices.ContainsFunc(matches, func(b BuffIcon) bool { return b.ID == 213911 }) {
			t.Fatalf("missing search result for %q", query)
		}
	}
}

func TestBuffListHistoryAndRoundTrip(t *testing.T) {
	e := Editor{Plan: New()}
	e.Step().Objects = []Object{{Kind: "Paladin", Label: "骑士", Radius: 1}, {Kind: "H1", Label: "H1", Radius: 1}, {Kind: "boss", Label: "首领", Radius: 3}}
	if !e.ToggleBuff(0, 213911) || !e.ToggleBuff(0, 210101) || !e.ToggleBuff(1, 215004) {
		t.Fatal("valid buff rejected")
	}
	rev := e.Revision()
	if e.ToggleBuff(2, 213911) || e.ToggleBuff(0, -1) || rev != e.Revision() {
		t.Fatal("invalid edit changed history")
	}
	e.AddStep(true)
	e.ToggleBuff(0, 213911)
	if !slices.Equal(e.Steps[0].Objects[0].Buffs, []int{213911, 210101}) || !slices.Equal(e.Step().Objects[0].Buffs, []int{210101}) {
		t.Fatal("copy did not isolate ordered list")
	}
	if !e.History(false) || !slices.Equal(e.Step().Objects[0].Buffs, e.Steps[0].Objects[0].Buffs) || !e.History(true) {
		t.Fatal("buff undo/redo")
	}
	raw, err := json.Marshal(e.Plan)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Decode(raw)
	if err != nil || !slices.Equal(p.Steps[1].Objects[0].Buffs, []int{210101}) {
		t.Fatalf("round trip: %v", err)
	}
}

func TestInvalidBuffLists(t *testing.T) {
	for name, ids := range map[string][]int{"unknown": {1}, "duplicate": {213911, 213911}, "too many": make([]int, MaxBuffIcons+1)} {
		o := Object{Kind: "MT", Label: "MT", Radius: 1, Buffs: ids}
		if name == "too many" {
			for i := range ids {
				ids[i] = buffIcons[i].ID
			}
		}
		p := New()
		p.Steps[0].Objects = []Object{o}
		data, _ := json.Marshal(p)
		if _, err := Decode(data); err == nil {
			t.Fatal("accepted " + name)
		}
	}
	e := Editor{Plan: New()}
	e.Step().Objects = []Object{{Kind: "Paladin"}}
	for i := range MaxBuffIcons {
		if !e.ToggleBuff(0, buffIcons[i].ID) {
			t.Fatal(strconv.Itoa(i))
		}
	}
	if e.ToggleBuff(0, buffIcons[MaxBuffIcons].ID) {
		t.Fatal("accepted overflowing list")
	}
}
