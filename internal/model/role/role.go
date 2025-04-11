package role

type RoleType int

const (
	MT RoleType = iota
	ST
	H1
	H2
	D1
	D2
	D3
	D4
	Paladin
	Warrior
	DarkKnight
	Gunbreaker
	WhiteMage
	Scholar
	Astrologian
	Sage
	Monk
	Dragoon
	Ninja
	Samurai
	Reaper
	Viper
	Bard
	Machinist
	Dancer
	BlackMage
	Summoner
	RedMage
	Pictomancer
	Boss
	NPC
	Pet
	LimitBreak
	Special
)

var roleToStringMap = map[RoleType]string{
	MT:          "MT",
	ST:          "ST",
	H1:          "H1",
	H2:          "H2",
	D1:          "D1",
	D2:          "D2",
	D3:          "D3",
	D4:          "D4",
	Paladin:     "Paladin",
	Warrior:     "Warrior",
	DarkKnight:  "DarkKnight",
	Gunbreaker:  "Gunbreaker",
	WhiteMage:   "WhiteMage",
	Scholar:     "Scholar",
	Astrologian: "Astrologian",
	Sage:        "Sage",
	Monk:        "Monk",
	Dragoon:     "Dragoon",
	Ninja:       "Ninja",
	Samurai:     "Samurai",
	Reaper:      "Reaper",
	Viper:       "Viper",
	Bard:        "Bard",
	Machinist:   "Machinist",
	Dancer:      "Dancer",
	BlackMage:   "BlackMage",
	Summoner:    "Summoner",
	RedMage:     "RedMage",
	Pictomancer: "Pictomancer",
	Boss:        "Boss",
	NPC:         "NPC",
	Pet:         "Pet",
	LimitBreak:  "LimitBreak",
	Special:     "Special",
}

func (r RoleType) String() string {
	if str, ok := roleToStringMap[r]; ok {
		return str
	}

	return "Unknown"
}

var stringToRoleMap = map[string]RoleType{
	"MT":          MT,
	"ST":          ST,
	"H1":          H1,
	"H2":          H2,
	"D1":          D1,
	"D2":          D2,
	"D3":          D3,
	"D4":          D4,
	"Paladin":     Paladin,
	"Warrior":     Warrior,
	"DarkKnight":  DarkKnight,
	"Gunbreaker":  Gunbreaker,
	"WhiteMage":   WhiteMage,
	"Scholar":     Scholar,
	"Astrologian": Astrologian,
	"Sage":        Sage,
	"Monk":        Monk,
	"Dragoon":     Dragoon,
	"Ninja":       Ninja,
	"Samurai":     Samurai,
	"Reaper":      Reaper,
	"Viper":       Viper,
	"Bard":        Bard,
	"Machinist":   Machinist,
	"Dancer":      Dancer,
	"BlackMage":   BlackMage,
	"Summoner":    Summoner,
	"RedMage":     RedMage,
	"Pictomancer": Pictomancer,
	"Boss":        Boss,
	"NPC":         NPC,
	"Pet":         Pet,
	"LimitBreak":  LimitBreak,
	"Special":     Special,
}

func StringToRole(s string) RoleType {
	if role, ok := stringToRoleMap[s]; ok {
		return role
	}

	return -1
}

var SpecialBoss = map[int64]bool{
	17827: true,
	17841: true,
}
