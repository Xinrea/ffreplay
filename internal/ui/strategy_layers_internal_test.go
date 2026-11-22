package ui

import (
	"image"
	"testing"
)

func TestLayerDropMatchesVisibleInsertionLine(t *testing.T) {
	// A partial last page of eight layers: visible document indices 2, 1, 0.
	p := layerPage{top: image.Pt(50, 195), first: 2, rows: 3}
	for _, tc := range []struct{ from, offset, gap, to int }{
		{0, 0, 0, 2},                // bottom to top
		{2, 3 * layerPitch, 3, 0},   // top to bottom
		{0, layerPitch + 2, 1, 1},   // above the middle row
		{0, 2*layerPitch - 1, 2, 0}, // below the middle row: already there
		{1, layerPitch, 1, 1},       // either edge of the source is a no-op
		{1, 2 * layerPitch, 2, 1},
	} {
		gap, to, ok := p.drop(p.top.Add(image.Pt(12, tc.offset)), tc.from)
		if !ok || gap != tc.gap || to != tc.to {
			t.Errorf("%+v: got gap %d target %d valid %v", tc, gap, to, ok)
		}
	}

	for _, at := range []image.Point{p.top.Sub(image.Pt(1, 0)), p.top.Add(image.Pt(30, 3*layerPitch+4))} {
		if _, _, ok := p.drop(at, 0); ok {
			t.Errorf("accepted outside drop at %v", at)
		}
	}
	// Moving to this page's bottom must not jump into the following page.
	p.first, p.rows = 7, 5
	if _, to, _ := p.drop(p.top.Add(image.Pt(0, 5*layerPitch)), 7); to != 3 {
		t.Fatal(to)
	}
}
