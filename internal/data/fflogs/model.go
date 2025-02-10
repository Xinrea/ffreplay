package fflogs

import (
	"encoding/json"
	"net/http"
)

type PlayerDetail struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Server string `json:"server"`
}

type PlayerDetails struct {
	Tanks   []PlayerDetail `json:"tanks"`
	Healers []PlayerDetail `json:"healers"`
	DPS     []PlayerDetail `json:"dps"`
}

type Credentials struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type BearerAuthTransport struct {
	Token string
}

type ReportFight struct {
	ID           int
	Name         string
	StartTime    float64
	EndTime      float64
	EnemyNPCs    []ReportFightNPC `json:"enemyNPCs"`
	FriendlyPets []ReportFightNPC `json:"friendlyPets"`
	Maps         []struct {
		ID int
	}
	PhaseTransitions []struct {
		ID        int
		StartTime int64
	}
}

type ReportFightNPC struct {
	GameID        int64
	ID            int64
	InstanceCount int
}

type GameMap struct {
	ID         int64   `json:"id"`
	FileName   string  `json:"filename"`
	SizeFactor float64 `json:"sizeFactor"`
	OffsetX    int     `json:"offsetX"`
	OffsetY    int     `json:"offsetY"`
}

func (t *BearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	return http.DefaultTransport.RoundTrip(req)
}

type FFLogsEvent struct {
	Timestamp           int64         `json:"timestamp"`
	LocalTick           int64         `json:"-"`
	Type                EventType     `json:"type"`
	Fight               int64         `json:"fight"`
	SourceID            *int64        `json:"sourceID,omitempty"`
	Gear                []interface{} `json:"gear,omitempty"`
	Auras               []Aura        `json:"auras,omitempty"`
	Level               *int64        `json:"level,omitempty"`
	SimulatedCrit       *float64      `json:"simulatedCrit,omitempty"`
	SimulatedDirectHit  *float64      `json:"simulatedDirectHit,omitempty"`
	Strength            *int64        `json:"strength,omitempty"`
	Dexterity           *int64        `json:"dexterity,omitempty"`
	Vitality            *int64        `json:"vitality,omitempty"`
	Intelligence        *int64        `json:"intelligence,omitempty"`
	Mind                *int64        `json:"mind,omitempty"`
	Piety               *int64        `json:"piety,omitempty"`
	Attack              *int64        `json:"attack,omitempty"`
	DirectHit           interface{}   `json:"directHit,omitempty"` // Can be bool or int64
	CriticalHit         *int64        `json:"criticalHit,omitempty"`
	AttackMagicPotency  *int64        `json:"attackMagicPotency,omitempty"`
	HealMagicPotency    *int64        `json:"healMagicPotency,omitempty"`
	Determination       *int64        `json:"determination,omitempty"`
	SkillSpeed          *int64        `json:"skillSpeed,omitempty"`
	SpellSpeed          *int64        `json:"spellSpeed,omitempty"`
	Tenacity            *int64        `json:"tenacity,omitempty"`
	TargetID            *int64        `json:"targetID,omitempty"`
	Ability             *Ability      `json:"ability,omitempty"`
	SourceResources     *Resources    `json:"sourceResources,omitempty"`
	TargetResources     *Resources    `json:"targetResources,omitempty"`
	HitType             *int64        `json:"hitType,omitempty"`
	Amount              *int64        `json:"amount,omitempty"`
	UnmitigatedAmount   *int64        `json:"unmitigatedAmount,omitempty"`
	Multiplier          *float64      `json:"multiplier,omitempty"`
	PacketID            *int64        `json:"packetID,omitempty"`
	Unpaired            *bool         `json:"unpaired,omitempty"`
	ExtraAbilityGameID  *int64        `json:"extraAbilityGameID,omitempty"`
	ExtraInfo           *int64        `json:"extraInfo,omitempty"`
	Duration            *int64        `json:"duration,omitempty"`
	Value               *int64        `json:"value,omitempty"`
	Bars                *int64        `json:"bars,omitempty"`
	Melee               *bool         `json:"melee,omitempty"`
	Overheal            *int64        `json:"overheal,omitempty"`
	Tick                *bool         `json:"tick,omitempty"`
	FinalizedAmount     *float64      `json:"finalizedAmount,omitempty"`
	Simulated           *bool         `json:"simulated,omitempty"`
	ExpectedAmount      *int64        `json:"expectedAmount,omitempty"`
	ExpectedCritRate    *int64        `json:"expectedCritRate,omitempty"`
	ActorPotencyRatio   *float64      `json:"actorPotencyRatio,omitempty"`
	GuessAmount         *float64      `json:"guessAmount,omitempty"`
	DirectHitPercentage *float64      `json:"directHitPercentage,omitempty"`
	BonusPercent        *int64        `json:"bonusPercent,omitempty"`
	Stack               *int64        `json:"stack,omitempty"`
	GaugeID             *string       `json:"gaugeID,omitempty"`
	Data1               *string       `json:"data1,omitempty"`
	Data2               *string       `json:"data2,omitempty"`
	Data3               *string       `json:"data3,omitempty"`
	Data4               *string       `json:"data4,omitempty"`
	TargetInstance      *int64        `json:"targetInstance,omitempty"`
	Absorbed            *int64        `json:"absorbed,omitempty"`
	AttackerID          *int64        `json:"attackerID,omitempty"`
	Absorb              *int64        `json:"absorb,omitempty"`
	SourceInstance      *int64        `json:"sourceInstance,omitempty"`
	MapID               *int          `json:"mapID,omitempty"`
	Icon                *int          `json:"icon,omitempty"`
	X                   *int64        `json:"x,omitempty"`
	Y                   *int64        `json:"y,omitempty"`
	SourceMarker        *int          `json:"sourceMarker,omitempty"`
	TargetMarker        *int          `json:"targetMarker,omitempty"`
	Buffs               string        `json:"buffs,omitempty"`
}

func (e FFLogsEvent) ToJson() string {
	jsonStr, _ := json.Marshal(e)

	return string(jsonStr)
}

type Aura struct {
	Source  int64  `json:"source"`
	Ability int64  `json:"ability"`
	Stacks  int64  `json:"stacks"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
}

type Resources struct {
	HitPoints    int64  `json:"hitPoints"`
	MaxHitPoints int64  `json:"maxHitPoints"`
	Mp           int64  `json:"mp"`
	MaxMP        int64  `json:"maxMP"`
	Tp           int64  `json:"tp"`
	MaxTP        int64  `json:"maxTP"`
	X            int64  `json:"x"`
	Y            int64  `json:"y"`
	Facing       int64  `json:"facing"`
	Absorb       *int64 `json:"absorb,omitempty"`
}

type EventType string

const (
	Absorbed           EventType = "absorbed"
	Applybuff          EventType = "applybuff"
	Refreshbuff        EventType = "refreshbuff"
	RefreshDebuff      EventType = "refreshdebuff"
	Applybuffstack     EventType = "applybuffstack"
	Applydebuff        EventType = "applydebuff"
	Begincast          EventType = "begincast"
	Calculateddamage   EventType = "calculateddamage"
	Calculatedheal     EventType = "calculatedheal"
	Cast               EventType = "cast"
	Combatantinfo      EventType = "combatantinfo"
	TDamage            EventType = "damage"
	THeal              EventType = "heal"
	Death              EventType = "death"
	Gaugeupdate        EventType = "gaugeupdate"
	Limitbreakupdate   EventType = "limitbreakupdate"
	Removebuff         EventType = "removebuff"
	RemoveDebuff       EventType = "removedebuff"
	Removebuffstack    EventType = "removebuffstack"
	MapChange          EventType = "mapchange"
	WorldMarkerRemoved EventType = "worldmarkerremoved"
	WorldMarkerPlaced  EventType = "worldmarkerplaced"
	DungeonStart       EventType = "dungeonstart"
	Tether             EventType = "tether"
)

type Actor struct {
	GameID  int64  `json:"gameID"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	SubType string `json:"subType"`
}
