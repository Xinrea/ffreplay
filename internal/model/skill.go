package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
)

type SkillTimeOption struct {
	StartTick   int64
	CastTime    int64
	DisplayTime int64
}

type SkillData struct {
	Time      SkillTimeOption
	GameSkill GameSkill
}

type GameSkill interface {
	Name() string
	Caster() *donburi.Entry
	Target() *donburi.Entry
	Update()
	Range() object.Object
	InRange(target *donburi.Entry) bool
	Effect(target *donburi.Entry)
}
