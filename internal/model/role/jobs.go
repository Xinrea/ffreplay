package role

// JobOption is one selectable playable job in editor UIs.
type JobOption struct {
	Role  RoleType
	Label string
}

// PlayableJobOptions returns every concrete FFXIV job (excludes party slots and non-player roles).
func PlayableJobOptions() []JobOption {
	roles := []RoleType{
		Paladin, Warrior, DarkKnight, Gunbreaker,
		WhiteMage, Scholar, Astrologian, Sage,
		Monk, Dragoon, Ninja, Samurai, Reaper, Viper,
		Bard, Machinist, Dancer,
		BlackMage, Summoner, RedMage, Pictomancer,
	}
	opts := make([]JobOption, 0, len(roles))
	for _, r := range roles {
		opts = append(opts, JobOption{Role: r, Label: r.String()})
	}
	return opts
}

// FindJobOption returns the option for role, or nil if it is not a playable job.
func FindJobOption(r RoleType) *JobOption {
	for _, opt := range PlayableJobOptions() {
		if opt.Role == r {
			copy := opt
			return &copy
		}
	}
	return nil
}
