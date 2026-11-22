package planner_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/Xinrea/ffreplay/internal/planner"
)

const paladin = "Paladin"

func TestStatusIconCatalogue(t *testing.T) {
	icons := planner.FindBuffIcons("")

	files, err := filepath.Glob("../../asset/abilities/2?????-??????.png")

	if err != nil || len(files) == 0 || len(files) != len(icons) {
		t.Fatalf(
			"catalogue vs artwork: %d/%d (%v)",
			len(icons),
			len(files),
			err,
		)
	}

	for _, icon := range icons {
		_, err := os.Stat(filepath.Join("../../asset/abilities", icon.Asset))
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, query := range []string{"铁壁", "213911"} {
		matches := planner.FindBuffIcons(query)
		if !slices.ContainsFunc(matches, func(b planner.BuffIcon) bool { return b.ID == 213911 }) {
			t.Fatalf("missing search result for %q", query)
		}
	}
}

func TestBuffListHistoryAndRoundTrip(t *testing.T) {
	e := planner.Editor{Plan: planner.New()}

	e.Step().Objects = []planner.Object{
		{Kind: paladin, Label: "骑士", Radius: 1},
		{Kind: "H1", Label: "H1", Radius: 1},
		{Kind: planner.KindBoss, Label: "首领", Radius: 3},
	}

	if !e.ToggleBuff(0, 213911) || !e.ToggleBuff(0, 210101) || !e.ToggleBuff(1, 215004) {
		t.Fatal("valid buff rejected")
	}

	rev := e.Revision()
	if e.ToggleBuff(2, 213911) || e.ToggleBuff(0, -1) || rev != e.Revision() {
		t.Fatal("invalid edit changed history")
	}

	checkBuffHistory(t, &e)
}

func checkBuffHistory(t *testing.T, e *planner.Editor) {
	t.Helper()
	e.AddStep(true)
	e.ToggleBuff(0, 213911)

	if !slices.Equal(e.Steps[0].Objects[0].Buffs, []int{213911, 210101}) ||
		!slices.Equal(e.Step().Objects[0].Buffs, []int{210101}) {
		t.Fatal("copy did not isolate ordered list")
	}

	if !e.History(false) || !slices.Equal(e.Step().Objects[0].Buffs, e.Steps[0].Objects[0].Buffs) ||
		!e.History(true) {
		t.Fatal("buff undo/redo")
	}

	raw, err := json.Marshal(e.Plan)
	if err != nil {
		t.Fatal(err)
	}

	checkBuffRoundTrip(t, raw)
}

func checkBuffRoundTrip(t *testing.T, raw []byte) {
	t.Helper()

	p, err := planner.Decode(raw)
	if err != nil || !slices.Equal(p.Steps[1].Objects[0].Buffs, []int{210101}) {
		t.Fatalf("round trip: %v", err)
	}
}

func TestInvalidBuffLists(t *testing.T) {
	icons := planner.FindBuffIcons("")

	for name, ids := range map[string][]int{
		"unknown": {1}, "duplicate": {213911, 213911},
		"too many": make([]int, planner.MaxBuffIcons+1),
	} {
		o := planner.Object{Kind: "MT", Label: "MT", Radius: 1, Buffs: ids}

		if name == "too many" {
			for i := range ids {
				ids[i] = icons[i].ID
			}
		}

		p := planner.New()
		p.Steps[0].Objects = []planner.Object{o}

		data, _ := json.Marshal(p)

		_, err := planner.Decode(data)
		if err == nil {
			t.Fatal("accepted " + name)
		}
	}

	e := planner.Editor{Plan: planner.New()}

	e.Step().Objects = []planner.Object{{Kind: paladin}}

	for i := range planner.MaxBuffIcons {
		if !e.ToggleBuff(0, icons[i].ID) {
			t.Fatal(strconv.Itoa(i))
		}
	}

	if e.ToggleBuff(0, icons[planner.MaxBuffIcons].ID) {
		t.Fatal("accepted overflowing list")
	}
}
