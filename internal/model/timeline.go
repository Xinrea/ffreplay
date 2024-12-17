package model

import (
	"github.com/yohamta/donburi/ecs"
)

type TimelineData struct {
	Name      string
	BeginTick int64
	Events    []*Event
}

type Event struct {
	Offset int64
	Action func(ecs *ecs.ECS)
	Done   bool
}

func (t *TimelineData) Reset() {
	for _, e := range t.Events {
		e.Done = false
	}
	t.BeginTick = 0
}
