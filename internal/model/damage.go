package model

type DamageType int

const (
	Physical DamageType = iota
	Magical
	Direct
)

type Damage struct {
	Type   DamageType
	Amount int
}

type Heal struct {
	Amount int
}
