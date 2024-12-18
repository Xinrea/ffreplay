package model

type Skill struct {
	ID        int64
	Name      string
	StartTick int64
	Cast      int64
	Recast    int64

	SkillEvents *TimelineData
}
