package stationconfig

// Config holds player-configurable station generation parameters.
type Config struct {
	CrewCapacity        int  // crew spawned per floor; also controls crew_quarters room count
	ScienceLabCount     int  // general_lab rooms placed and researchers spawned on science floor
	MedCount            int  // medical_bay rooms placed and medics spawned on science floor
	EngineeringCapacity int  // engineering_workshop rooms placed on engineering floor
	SecurityCapacity    int  // security_office rooms placed on command floor
	LifePodBayCount     int  // life_pod_bay rooms placed on engineering floor
	SelfDestructEnabled bool // whether self_destruct_room is placed
}

var active = defaults()

func defaults() Config {
	return Config{
		CrewCapacity:        10,
		ScienceLabCount:     3,
		MedCount:            2,
		EngineeringCapacity: 4,
		SecurityCapacity:    2,
		LifePodBayCount:     2,
		SelfDestructEnabled: true,
	}
}

// Set replaces the active station config.
func Set(c Config) { active = c }

// Get returns the active station config.
func Get() Config { return active }

// Reset restores defaults.
func Reset() { active = defaults() }
