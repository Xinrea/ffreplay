package skills

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func NewCyclonicBreak() model.Skill {
	return model.Skill{
		ID:        SKillCyclonicBreak,
		Name:      "Cyclonic Break",
		StartTick: -1,
		Cast:      0,
		Recast:    0,
		SkillEvents: &model.TimelineData{
			Events: []model.Event{
				// first event is a direct damage
				{
					Offset:      0,
					DisplayTime: 200,
					EffectRange: object.NewFanObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, 20, 20*METER),
					Finish: func(ecs *ecs.ECS, rangeObj object.Object, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int) {
						casterPos := component.Sprite.Get(caster).Instances[casterInstance].Object.Position()
						targetPos := component.Sprite.Get(target).Instances[targetInstance].Object.Position()
						radian := targetPos.Sub(casterPos).Radian()
						rangeObj.UpdateRotate(radian)
					},
				},
				// at the same time, postion the range object to the target
				{
					Offset:      0,
					DisplayTime: 2000,
					EffectRange: object.NewFanObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, 20, 20*METER),
					Finish: func(ecs *ecs.ECS, rangeObj object.Object, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int) {
						casterPos := component.Sprite.Get(caster).Instances[casterInstance].Object.Position()
						targetPos := component.Sprite.Get(target).Instances[targetInstance].Object.Position()
						radian := targetPos.Sub(casterPos).Radian()
						rangeObj.UpdateRotate(radian)
					},
				},
				{
					Offset:      2000,
					DisplayTime: 2000,
					EffectRange: object.NewFanObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, 20, 20*METER),
					Finish: func(ecs *ecs.ECS, rangeObj object.Object, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int) {
						casterPos := component.Sprite.Get(caster).Instances[casterInstance].Object.Position()
						targetPos := component.Sprite.Get(target).Instances[targetInstance].Object.Position()
						radian := targetPos.Sub(casterPos).Radian()
						rangeObj.UpdateRotate(radian)
					},
				},
				{
					Offset:      4000,
					DisplayTime: 2000,
					EffectRange: object.NewFanObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, 20, 20*METER),
					Finish: func(ecs *ecs.ECS, rangeObj object.Object, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int) {
						casterPos := component.Sprite.Get(caster).Instances[casterInstance].Object.Position()
						targetPos := component.Sprite.Get(target).Instances[targetInstance].Object.Position()
						radian := targetPos.Sub(casterPos).Radian()
						rangeObj.UpdateRotate(radian)
					},
				},
			},
		},
	}
}

// func (s *CyclonicBreak) Start(tick int64) {
// 	s.startTick = tick
// 	// initialize skill
// 	if s.isDuplicated {
// 		// using previous target position
// 		return
// 	}
// 	casterPos := component.Sprite.Get(s.caster).Instances[s.casterInstance].Object.Position()
// 	targetPos := component.Sprite.Get(s.target).Instances[0].Object.Position()
// 	radian := targetPos.Sub(casterPos).Radian()
// 	s.rangeObj.Translate(casterPos)
// 	s.rangeObj.Rotate(radian)
// 	if !s.isDuplicated {
// 		caster := component.Sprite.Get(s.caster).Instances[s.casterInstance]
// 		another := *s
// 		another.startTick = -1
// 		another.isDuplicated = true
// 		caster.CastBehind(&another)
// 	}
// }

// func (s *CyclonicBreak) Instance(ecs *ecs.ECS, casterInstance int, caster, target *donburi.Entry) model.Skill {
// 	rangeObj := object.NewFanObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, 22.5, 20*METER)
// 	// using a copy
// 	return &CyclonicBreak{
// 		DefaultSkill: DefaultSkill{
// 			id:          s.id,
// 			name:        s.name,
// 			startTick:   s.startTick,
// 			castTime:    s.castTime,
// 			displayTime: s.displayTime,

// 			ecs:            ecs,
// 			caster:         caster,
// 			casterInstance: casterInstance,
// 			target:         target,
// 			rangeObj:       rangeObj,
// 		},
// 	}
// }
