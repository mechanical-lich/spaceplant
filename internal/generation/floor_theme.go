package generation

import "github.com/mechanical-lich/spaceplant/internal/utility"

// Layout constants define the macro shape of a floor.
const (
	LayoutRingSpokes   = "ring_spokes"   // circle hub + 4 hallway arms + budded rooms off arms
	LayoutGrid         = "grid"          // cross/grid halls with many small budded rooms
	LayoutIndustrialRing = "industrial_ring" // outer ring + inner maintenance tunnels, open bays
	LayoutOpenBays     = "open_bays"     // few large rooms, minimal budding
	LayoutRectangle    = "rectangle"     // corner rooms + connecting halls + central circle
)

// RoomWeight pairs a room tag with its relative spawn weight.
// Higher weight = more likely to be selected for a budded room.
type RoomWeight struct {
	Tag    string
	Weight int
}

// FloorTheme describes everything about a floor's generation:
// its macro layout, which room types to populate, and how many
// budded rooms to attempt.
type FloorTheme struct {
	Name                string
	Layout              string
	BudCount            int // max rooms PlaceRooms will attempt
	SecondaryStairCount int // how many secondary stair pairs connect this floor to the one above
	RoomWeights         []RoomWeight
}

// pickRoomTag selects a room tag from the theme's weighted table.
func (t *FloorTheme) pickRoomTag() string {
	if len(t.RoomWeights) == 0 {
		return ""
	}
	total := 0
	for _, rw := range t.RoomWeights {
		total += rw.Weight
	}
	n := utility.GetRandom(0, total)
	for _, rw := range t.RoomWeights {
		n -= rw.Weight
		if n <= 0 {
			return rw.Tag
		}
	}
	return t.RoomWeights[len(t.RoomWeights)-1].Tag
}

// --- Theme definitions ---

var ThemeEngineering = FloorTheme{
	Name:                "Engineering & Systems",
	Layout:              LayoutIndustrialRing,
	BudCount:            60,
	SecondaryStairCount: 4, // heavy service access between decks
	RoomWeights: []RoomWeight{
		{"reactor_core", 5},
		{"engineering_workshop", 15},
		{"maintenance_bay", 15},
		{"life_support_control", 10},
		{"water_waste_processing", 8},
		{"cargo_hold", 12},
		{"fuel_storage", 8},
		{"eva_bay", 10},
		{"utility_corridor", 17},
	},
}

var ThemeLogistics = FloorTheme{
	Name:                "Logistics & Industry",
	Layout:              LayoutOpenBays,
	BudCount:            0,
	SecondaryStairCount: 3, // cargo routes need multiple vertical connections
	RoomWeights: []RoomWeight{
		{"manufacturing_hangar", 10},
		{"robotics_bay", 10},
		{"freight_airlock", 8},
		{"freight_sorting", 15},
		{"storage_vault", 20},
		{"customs_inspection", 12},
		{"cargo_hold", 25},
	},
}

var ThemeHabitation = FloorTheme{
	Name:                "Habitation",
	Layout:              LayoutGrid,
	BudCount:            100,
	SecondaryStairCount: 2, // residents mostly use the main column
	RoomWeights: []RoomWeight{
		{"crew_quarters", 35},
		{"officers_suite", 15},
		{"guest_cabin", 10},
		{"family_apartment", 10},
		{"laundry", 8},
		{"childcare", 5},
		{"captains_quarters", 3},
	},
}

var ThemeCommerceSocial = FloorTheme{
	Name:                "Commerce & Social",
	Layout:              LayoutGrid,
	BudCount:            80,
	SecondaryStairCount: 2, // some back-access for staff and deliveries
	RoomWeights: []RoomWeight{
		{"mess_hall", 20},
		{"bar_cantina", 15},
		{"market", 15},
		{"shop", 15},
		{"recreation", 10},
		{"library", 8},
		{"bank", 5},
		{"administration", 5},
		{"meditation", 7},
	},
}

var ThemeScience = FloorTheme{
	Name:                "Science & Research",
	Layout:              LayoutRingSpokes,
	BudCount:            60,
	SecondaryStairCount: 1, // controlled access, one secondary route
	RoomWeights: []RoomWeight{
		{"general_lab", 20},
		{"biolab", 12},
		{"chemistry_lab", 12},
		{"fabrication_lab", 10},
		{"observatory", 8},
		{"medical_research", 10},
		{"hydroponics_lab", 12},
		{"data_center", 8},
		{"medical_bay", 8},
	},
}

var ThemeCommand = FloorTheme{
	Name:                "Operations & Command",
	Layout:              LayoutRingSpokes,
	BudCount:            50,
	SecondaryStairCount: 0, // command deck — main column only, no back routes
	RoomWeights: []RoomWeight{
		{"bridge", 8},
		{"mission_control", 12},
		{"comms_relay", 10},
		{"navigation", 10},
		{"docking_control", 10},
		{"security_office", 15},
		{"medical_bay", 10},
		{"brig", 8},
		{"courtroom", 5},
		{"executive_suite", 12},
	},
}

// FloorStack is the ordered list of themes from Z=0 (bottom) to top.
// Adjust order and length to change the station layout.
var FloorStack = []FloorTheme{
	ThemeEngineering,
	ThemeLogistics,
	ThemeHabitation,
	ThemeCommerceSocial,
	ThemeScience,
	ThemeCommand,
}
