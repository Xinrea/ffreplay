package model

import (
	"log"

	"github.com/Xinrea/ffreplay/pkg/object"
	"gopkg.in/yaml.v3"
)

type TimelineData struct {
	Name      string `yaml:"name"`
	StartTick int64  `yaml:"-"`
	// Caster is the entity that handles all skills in the timeline
	Events []*Event `yaml:"events"`
}

func LoadTimelineRaw(raw string) *TimelineData {
	var tl TimelineData

	err := yaml.Unmarshal([]byte(raw), &tl)
	if err != nil {
		log.Fatal(err)
	}

	return &tl
}

func (t *TimelineData) Init(
	tick int64,
) *TimelineData {
	t.StartTick = tick

	return t
}

func (t *TimelineData) Reset() {
	t.StartTick = 0
}

func (t *TimelineData) OffsetTickOf(index int) int64 {
	return t.Events[index].OffsetTick()
}

func (t *TimelineData) ProgressedTickOf(index int, currentTick int64) int64 {
	return t.Events[index].ProgressedTick(t.StartTick, currentTick)
}

type RangeType int

const (
	RangeTypeRect RangeType = iota
	RangeTypeCircle
	RangeTypeFan
	RangeTypeRing
)

type SkillTemplateConfigure struct {
	ID          int64
	Name        string
	Cast        int64
	Range       RangeType
	RangeOpt    object.ObjectOption
	Anchor      int
	Width       int
	Height      int
	Radius      int
	InnerRadius int
	Angle       float64
}
