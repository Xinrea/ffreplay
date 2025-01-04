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
)

func (r RoleType) String() string {
	switch r {
	case MT:
		return "MT"
	case ST:
		return "ST"
	case H1:
		return "H1"
	case H2:
		return "H2"
	case D1:
		return "D1"
	case D2:
		return "D2"
	case D3:
		return "D3"
	case D4:
		return "D4"
	case Paladin:
		return "Paladin"
	case Warrior:
		return "Warrior"
	case DarkKnight:
		return "DarkKnight"
	case Gunbreaker:
		return "Gunbreaker"
	case WhiteMage:
		return "WhiteMage"
	case Scholar:
		return "Scholar"
	case Astrologian:
		return "Astrologian"
	case Sage:
		return "Sage"
	case Monk:
		return "Monk"
	case Dragoon:
		return "Dragoon"
	case Ninja:
		return "Ninja"
	case Samurai:
		return "Samurai"
	case Reaper:
		return "Reaper"
	case Viper:
		return "Viper"
	case Bard:
		return "Bard"
	case Machinist:
		return "Machinist"
	case Dancer:
		return "Dancer"
	case BlackMage:
		return "BlackMage"
	case Summoner:
		return "Summoner"
	case RedMage:
		return "RedMage"
	case Pictomancer:
		return "Pictomancer"
	case Boss:
		return "Boss"
	case NPC:
		return "NPC"
	case Pet:
		return "Pet"
	default:
		return "Unknown"
	}
}

func StringToRole(s string) RoleType {
	switch s {
	case "MT":
		return MT
	case "ST":
		return ST
	case "H1":
		return H1
	case "H2":
		return H2
	case "D1":
		return D1
	case "D2":
		return D2
	case "D3":
		return D3
	case "D4":
		return D4
	case "Paladin":
		return Paladin
	case "Warrior":
		return Warrior
	case "DarkKnight":
		return DarkKnight
	case "Gunbreaker":
		return Gunbreaker
	case "WhiteMage":
		return WhiteMage
	case "Scholar":
		return Scholar
	case "Astrologian":
		return Astrologian
	case "Sage":
		return Sage
	case "Monk":
		return Monk
	case "Dragoon":
		return Dragoon
	case "Ninja":
		return Ninja
	case "Samurai":
		return Samurai
	case "Reaper":
		return Reaper
	case "Viper":
		return Viper
	case "Bard":
		return Bard
	case "Machinist":
		return Machinist
	case "Dancer":
		return Dancer
	case "BlackMage":
		return BlackMage
	case "Summoner":
		return Summoner
	case "RedMage":
		return RedMage
	case "Pictomancer":
		return Pictomancer
	case "Boss":
		return Boss
	case "NPC":
		return NPC
	case "Pet":
		return Pet
	default:
		return -1
	}
}
