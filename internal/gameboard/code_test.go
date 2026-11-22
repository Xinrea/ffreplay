package gameboard_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Xinrea/ffreplay/internal/gameboard"
)

// External "Full Party" fixture from https://github.com/mczub/xiv-strat-board.
const fullParty = "[stgy:aGz4kwGPfaf8h3GsnyGY8RcjsvIuShQZmcZtFzkdlUwjlvyIRWEM51OI5bb5wPoW9i5fdIqflSXuQux" +
	"O-ssbT0x7z7vVNICXYJLPw7BJrkNEkkdL1PsxTGiDSFSfaQVelWWN705StQ-Cfi25ZbbkEPv2nSuoIdeqFF85" +
	"54DetpRZgJT+LXbHdZD2nCqwbMyhNX2kFAsLY-RNiNgl+BPNiNOPOLG+NrovN-mG+Qy8MQrjKFGrLUELwIq-]"

func TestExternalCode(t *testing.T) {
	b, err := gameboard.Decode(fullParty)
	if err != nil {
		t.Fatal(err)
	}

	if b.Name != "Full Party" || b.Background != 2 || len(b.Objects) != 17 ||
		b.Objects[0].Type != 64 ||
		b.Objects[16].Type != 86 {
		t.Fatal(b)
	}
}

func TestWireRoundTrip(t *testing.T) {
	boards := []gameboard.Board{{Name: "空", Background: 1}, {Name: "分摊", Background: 4, Objects: []gameboard.Object{
		{Type: 100, Text: "南集合", X: -300, Y: 1200, Flags: 9, Angle: -30, RGBA: [4]byte{2, 3, 4, 5}, Scale: 100},
		{Type: 12, Params: [3]uint16{5130, 42, 3}},
	}}}
	for _, b := range boards {
		code, err := gameboard.Encode(b)
		if err != nil {
			t.Fatal(err)
		}

		got, err := gameboard.Decode(code)
		if err != nil || !reflect.DeepEqual(got, b) {
			t.Fatal(err, got, b)
		}
	}
}

func TestRejectCorruptCode(t *testing.T) {
	for _, s := range []string{
		"#ffplan=abcd", fullParty[:20] + "]",
		strings.Replace(fullParty, "kwGP", "kwGQ", 1), fullParty[:19] + "!" + fullParty[20:],
	} {
		_, err := gameboard.Decode(s)
		if err == nil {
			t.Fatal("accepted", s)
		}
	}
}
