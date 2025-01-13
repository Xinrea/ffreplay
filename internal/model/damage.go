package model

type DamageType int

const (
	Physical DamageType = 128
	Magical  DamageType = 1024
	Special  DamageType = 32
)

type Damage struct {
	Type   DamageType
	Amount int
}

type Heal struct {
	Amount int
}
