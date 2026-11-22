package planner_test

import (
	"testing"

	"github.com/Xinrea/ffreplay/internal/gameboard"
	"github.com/Xinrea/ffreplay/internal/planner"
)

func exportOne(t *testing.T, o planner.Object) gameboard.Object {
	t.Helper()

	p := planner.New()
	p.Steps[0].Objects = []planner.Object{o}

	code, err := planner.ExportGame(p, 0)
	if err != nil {
		t.Fatal(err)
	}

	b, err := gameboard.Decode(code)
	if err != nil {
		t.Fatal(err)
	}

	return b.Objects[0]
}

func TestFanAngleInDegrees(t *testing.T) {
	for _, angle := range []float64{1, 45, 93, 360} {
		g := exportOne(t, planner.Object{Kind: planner.KindCone, Shape: angle, Radius: 6})
		if g.Params[0] != uint16(angle) {
			t.Fatalf("want %v degrees, got %d", angle, g.Params[0])
		}
	}
}

func TestGameCircleColor(t *testing.T) {
	original := gameboard.Object{Type: 9, Flags: 1, X: 2560, Y: 1920, Scale: 80, RGBA: [4]byte{11, 185, 241, 21}}

	code, err := gameboard.Encode(gameboard.Board{Background: 2, Objects: []gameboard.Object{original}})
	if err != nil {
		t.Fatal(err)
	}

	p, _, err := planner.ImportGame(code)
	if err != nil {
		t.Fatal(err)
	}

	o := p.Steps[0].Objects[0]
	if o.Color != 1 || !o.UseGameInk() || exportOne(t, o) != original {
		t.Fatal("lost imported circle ink")
	}

	o.Color = 3

	edited := exportOne(t, o)
	if edited.RGBA != [4]byte{121, 216, 164, 21} {
		t.Fatal("recolor lost color or transparency", edited)
	}

	created := exportOne(t, planner.Object{Kind: planner.KindCircle, Color: 1, Radius: 4})
	if created.RGBA != [4]byte{90, 195, 243, 0} {
		t.Fatal("new circle ignored the selected color", created)
	}
}
