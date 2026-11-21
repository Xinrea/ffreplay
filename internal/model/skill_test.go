package model_test

import (
	"testing"

	"github.com/Xinrea/ffreplay/internal/model"
)

func TestParseActionRecordsRejectsShortSheets(t *testing.T) {
	_, err := model.ParseActionRecords(nil)
	if err == nil {
		t.Fatal("empty sheet parsed")
	}

	headers := [][]string{{"key"}, {"type"}, {"index"}}

	_, err = model.ParseActionRecords(headers[:2])
	if err == nil {
		t.Fatal("short header parsed")
	}

	_, err = model.ParseActionRecords(append(headers, []string{"1", "only-two"}))
	if err == nil {
		t.Fatal("short row parsed")
	}
}

func TestParseActionRecordsReadsGCDFlag(t *testing.T) {
	entries, err := model.ParseActionRecords([][]string{
		{"key"},
		{"type"},
		{"index"},
		{"7", "Fire", "", "", "2"},
		{"8", "Sprint", "", "", "1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("entries: %+v", entries)
	}

	if entries[0].ID != 7 || entries[0].Name != "Fire" || !entries[0].IsGCD {
		t.Fatalf("gcd row: %+v", entries[0])
	}

	if entries[1].ID != 8 || entries[1].Name != "Sprint" || entries[1].IsGCD {
		t.Fatalf("weaponskill row: %+v", entries[1])
	}
}
