package generation

import "github.com/mechanical-lich/spaceplant/internal/world"

// PlacementRegion constrains where in a room a piece of furniture should be placed.
type PlacementRegion int

const (
	RegionAnywhere  PlacementRegion = iota
	RegionNorthWall                  // row adjacent to north interior wall
	RegionSouthWall                  // row adjacent to south interior wall
	RegionEastWall                   // column adjacent to east interior wall
	RegionWestWall                   // column adjacent to west interior wall
	RegionCenter                     // middle third of both axes
)

// PlacementHint tells populateRoom to place a specific blueprint in a specific region.
type PlacementHint struct {
	Blueprint string
	Region    PlacementRegion
}

// RoomGenerator sculpts the interior of a room and returns placement hints for the
// populate pass. doorDir is the direction from the hallway into the room (the direction
// the room grew); zero doorDir means the room has no bud door.
type RoomGenerator interface {
	Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint
}

// RoomSizeRange defines the preferred width and height bounds for a room tag.
type RoomSizeRange struct {
	MinW, MaxW, MinH, MaxH int
}

// defaultRoomSize is used for any tag that has no explicit entry in roomSizes.
var defaultRoomSize = RoomSizeRange{7, 11, 6, 9}

// roomSizes defines preferred dimensions per tag. Sized so each generator has
// enough interior space to be effective without feeling wastefully large.
var roomSizes = map[string]RoomSizeRange{
	// Habitation — sleeping rooms need space for a partition
	"crew_quarters":     {9, 13, 8, 11},
	"officers_suite":    {10, 13, 9, 11},
	"captains_quarters": {11, 15, 10, 12},
	"guest_cabin":       {6, 9, 5, 8},  // no partition, intentionally compact
	"family_apartment":  {11, 15, 10, 12},
	"childcare":         {9, 12, 8, 11},
	"laundry":           {6, 9, 5, 8},
	"executive_suite":   {11, 15, 10, 12},

	// Command — bridge needs platform room; others are focused workspaces
	"bridge":          {11, 15, 10, 13},
	"mission_control": {10, 14, 9, 12},
	"comms_relay":     {8, 11, 7, 10},
	"navigation":      {8, 11, 7, 10},
	"docking_control": {9, 12, 8, 11},
	"security_office": {8, 11, 7, 10},

	// Engineering — reactor and cargo need generous space; corridors are tight
	"reactor_core":           {10, 14, 9, 12},
	"engineering_workshop":   {9, 13, 8, 11},
	"maintenance_bay":        {8, 12, 7, 10},
	"life_support_control":   {9, 12, 8, 11},
	"water_waste_processing": {9, 12, 8, 11},
	"cargo_hold":             {11, 16, 9, 13},
	"fuel_storage":           {9, 12, 8, 11},
	"eva_bay":                {9, 13, 8, 11},
	"utility_corridor":       {5, 8, 4, 7},

	// Logistics — hangars are large open bays
	"manufacturing_hangar": {13, 18, 11, 14},
	"robotics_bay":         {9, 13, 8, 11},
	"freight_airlock":      {9, 12, 8, 11},
	"freight_sorting":      {11, 15, 9, 12},
	"storage_vault":        {9, 13, 8, 11},
	"customs_inspection":   {7, 10, 6, 9},

	// Commerce & Social — communal spaces are larger; intimate shops are smaller
	"mess_hall":      {12, 16, 10, 13},
	"bar_cantina":    {10, 14, 9, 12},
	"market":         {10, 14, 9, 12},
	"shop":           {7, 10, 6, 9},
	"recreation":     {10, 14, 9, 12},
	"library":        {10, 13, 8, 11},
	"bank":           {7, 10, 6, 9},
	"administration": {9, 13, 8, 11},
	"meditation":     {8, 11, 7, 10},

	// Science — labs need bench-row space; observatory wants near-square
	"general_lab":      {9, 13, 8, 11},
	"biolab":           {9, 13, 8, 11},
	"chemistry_lab":    {9, 13, 8, 11},
	"fabrication_lab":  {9, 12, 8, 11},
	"observatory":      {10, 13, 10, 13},
	"medical_research": {9, 12, 8, 11},
	"hydroponics_lab":  {10, 14, 9, 12},
	"data_center":      {10, 14, 8, 11},

	// Medical — bed-row rooms need width for parallel rows
	"medical_bay": {11, 15, 9, 12},
	"surgery":     {9, 12, 8, 11},
	"quarantine":  {10, 13, 8, 11},
	"pharmacy":    {7, 10, 6, 9},
	"morgue":      {10, 13, 8, 11},
	"cryo":        {11, 15, 9, 12},

	// Justice & Security
	"brig":          {8, 11, 7, 10},
	"courtroom":     {10, 14, 9, 12},
	"interrogation": {5, 8, 5, 8},
	"forensics":     {8, 11, 7, 10},
}

// RoomSizeFor returns the preferred size range for the given tag,
// falling back to defaultRoomSize if none is defined.
func RoomSizeFor(tag string) RoomSizeRange {
	if s, ok := roomSizes[tag]; ok {
		return s
	}
	return defaultRoomSize
}

// roomGenerators is the dispatch table mapping room tag → generator.
var roomGenerators = map[string]RoomGenerator{
	// Habitation
	"crew_quarters":     &sleepingRoomGenerator{hasPartition: true, bedBlueprint: "bunk_bed", farItems: []string{"bunk_bed"}, nearItems: []string{"locker", "locker", "desk"}},
	"officers_suite":    &sleepingRoomGenerator{hasPartition: true, bedBlueprint: "single_bed", farItems: []string{"wardrobe"}, nearItems: []string{"desk", "couch"}},
	"captains_quarters": &sleepingRoomGenerator{hasPartition: true, bedBlueprint: "single_bed", farItems: []string{"safe", "bookshelf"}, nearItems: []string{"desk", "wardrobe"}},
	"guest_cabin":       &sleepingRoomGenerator{hasPartition: false, bedBlueprint: "single_bed", nearItems: []string{"locker"}},
	"family_apartment":  &sleepingRoomGenerator{hasPartition: true, bedBlueprint: "bunk_bed", farItems: []string{"single_bed"}, nearItems: []string{"kitchenette", "table", "chair", "chair"}},
	"childcare":         &sleepingRoomGenerator{hasPartition: false, bedBlueprint: "sleeping_cot", farItems: []string{"sleeping_cot"}, nearItems: []string{"storage_unit", "table"}},
	"laundry":           &workshopGenerator{sideItems: []string{"storage_unit", "sorting_bin"}, farItems: []string{"sorting_bin"}, nearItems: []string{"folding_table"}},

	// Command
	"bridge":          &bridgeGenerator{},
	"mission_control": &controlRoomGenerator{farItems: []string{"operator_console", "operator_console", "operator_console"}, centerItems: []string{"holo_display", "table"}},
	"comms_relay":     &controlRoomGenerator{farItems: []string{"antenna_rack", "signal_processor"}, sideItems: []string{"signal_processor", "operator_console"}},
	"navigation":      &controlRoomGenerator{farItems: []string{"navigation_console"}, centerItems: []string{"map_table", "holo_display"}},
	"docking_control": &controlRoomGenerator{farItems: []string{"docking_monitor", "docking_monitor"}, sideItems: []string{"status_light", "control_panel"}},
	"security_office": &controlRoomGenerator{farItems: []string{"monitoring_wall", "weapons_locker"}, sideItems: []string{"evidence_locker"}, nearItems: []string{"desk"}},
	"executive_suite": &sleepingRoomGenerator{hasPartition: true, bedBlueprint: "single_bed", farItems: []string{"safe"}, nearItems: []string{"desk", "couch", "holo_display"}},

	// Engineering
	"reactor_core":           &controlRoomGenerator{centerItems: []string{"containment_vessel"}, farItems: []string{"control_panel"}, sideItems: []string{"coolant_pipe", "radiation_shielding", "catwalk"}},
	"engineering_workshop":   &workshopGenerator{sideItems: []string{"workbench", "workbench"}, farItems: []string{"welding_rig"}, nearItems: []string{"tool_rack", "parts_bin"}},
	"maintenance_bay":        &workshopGenerator{sideItems: []string{"tool_rack", "shelving"}, farItems: []string{"maintenance_platform"}, nearItems: []string{"tool_cart", "diagnostic_panel"}},
	"life_support_control":   &controlRoomGenerator{farItems: []string{"environmental_panel"}, sideItems: []string{"control_panel", "control_panel"}},
	"water_waste_processing": &storageRoomGenerator{wallItems: []string{"filtration_tank", "filtration_tank"}, centerItems: []string{"diagnostic_panel"}},
	"cargo_hold":             &storageRoomGenerator{wallItems: []string{"crate", "crate", "pallet"}, centerItems: []string{"pallet_rack", "cargo_net"}},
	"fuel_storage":           &storageRoomGenerator{wallItems: []string{"fuel_tank", "fuel_tank", "pressure_gauge"}, centerItems: []string{"control_panel"}},
	"eva_bay":                &controlRoomGenerator{farItems: []string{"suit_rack", "suit_rack"}, nearItems: []string{"airlock_controls", "tool_rack"}},
	"utility_corridor":       &storageRoomGenerator{wallItems: []string{"coolant_pipe", "junction_box"}, centerItems: []string{"diagnostic_panel"}},

	// Logistics
	"manufacturing_hangar": &workshopGenerator{sideItems: []string{"workbench", "welding_rig"}, farItems: []string{"crane"}, nearItems: []string{"tool_rack", "spare_parts_shelf"}},
	"robotics_bay":         &workshopGenerator{sideItems: []string{"charging_station", "charging_station"}, farItems: []string{"workbench"}, nearItems: []string{"tool_rack"}},
	"freight_airlock":      &controlRoomGenerator{nearItems: []string{"airlock_controls"}, sideItems: []string{"crate", "crate"}, centerItems: []string{"pallet"}},
	"freight_sorting":      &socialRoomGenerator{centerItems: []string{"conveyor_belt"}, sideItems: []string{"sorting_bin", "sorting_bin"}, nearItems: []string{"crate", "crate"}},
	"storage_vault":        &storageRoomGenerator{wallItems: []string{"secure_locker", "secure_locker", "shelving"}, centerItems: []string{"safe"}},
	"customs_inspection":   &controlRoomGenerator{farItems: []string{"evidence_locker"}, nearItems: []string{"desk", "secure_locker"}},

	// Commerce & Social
	"mess_hall":      &socialRoomGenerator{centerItems: []string{"table", "table", "bench", "bench"}, farItems: []string{"serving_line", "kitchenette"}},
	"bar_cantina":    &socialRoomGenerator{farItems: []string{"bar_counter", "bar_stool", "bar_stool", "bar_stool"}, centerItems: []string{"table"}},
	"market":         &socialRoomGenerator{sideItems: []string{"shelving", "shelving"}, centerItems: []string{"table"}, nearItems: []string{"storage_unit"}},
	"shop":           &controlRoomGenerator{sideItems: []string{"shelving", "shelving"}, nearItems: []string{"workbench", "storage_unit", "secure_locker"}},
	"recreation":     &socialRoomGenerator{centerItems: []string{"table", "table", "bench", "bench", "bookshelf"}},
	"library":        &storageRoomGenerator{wallItems: []string{"bookshelf", "bookshelf", "bookshelf"}, centerItems: []string{"table", "chair"}},
	"bank":           &controlRoomGenerator{farItems: []string{"safe", "secure_locker"}, nearItems: []string{"desk"}},
	"administration": &socialRoomGenerator{centerItems: []string{"desk", "desk", "table"}, sideItems: []string{"filing_cabinet"}},
	"meditation":     &socialRoomGenerator{sideItems: []string{"planter", "planter"}, centerItems: []string{"bench", "bench"}},

	// Science & Research
	"general_lab":      &labStyleGenerator{sideItems: []string{"lab_bench", "sample_rack"}, farItems: []string{"analysis_instrument", "fume_hood"}},
	"biolab":           &labStyleGenerator{sideItems: []string{"bio_cabinet", "specimen_tank", "specimen_tank"}, farItems: []string{"incubator", "decontamination_shower"}},
	"chemistry_lab":    &labStyleGenerator{sideItems: []string{"lab_bench", "reagent_cabinet"}, farItems: []string{"fume_hood", "centrifuge"}},
	"fabrication_lab":  &labStyleGenerator{sideItems: []string{"workbench", "tool_rack"}, farItems: []string{"3d_printer", "spare_parts_shelf"}},
	"observatory":      &controlRoomGenerator{centerItems: []string{"telescope_mount"}, sideItems: []string{"sensor_console"}, farItems: []string{"holo_display"}},
	"medical_research": &labStyleGenerator{sideItems: []string{"lab_bench", "centrifuge"}, farItems: []string{"microscope", "specimen_freezer"}},
	"hydroponics_lab":  &labStyleGenerator{sideItems: []string{"grow_rack", "grow_rack", "nutrient_tank"}, farItems: []string{"nutrient_tank"}},
	"data_center":      &storageRoomGenerator{wallItems: []string{"server_rack", "server_rack", "server_rack"}, centerItems: []string{"diagnostic_panel"}},

	// Medical
	"medical_bay": &rowBedsGenerator{bedBlueprint: "exam_bed", farItems: []string{"diagnostic_console", "drug_cabinet"}, nearItems: []string{"vitals_monitor"}},
	"surgery":     &controlRoomGenerator{centerItems: []string{"operating_table"}, farItems: []string{"surgical_light", "sterilization_unit"}, sideItems: []string{"anesthesia_console"}},
	"quarantine":  &rowBedsGenerator{bedBlueprint: "exam_bed", nearItems: []string{"decontamination_shower", "ppe_station"}},
	"pharmacy":    &storageRoomGenerator{wallItems: []string{"drug_cabinet", "drug_cabinet", "shelving"}},
	"morgue":      &rowBedsGenerator{bedBlueprint: "morgue_slab", farItems: []string{"specimen_freezer"}},
	"cryo":        &rowBedsGenerator{bedBlueprint: "cryo_pod", farItems: []string{"monitoring_wall"}},

	// Justice & Security
	"brig":          &socialRoomGenerator{sideItems: []string{"bench", "bench"}, farItems: []string{"secure_locker"}},
	"courtroom":     &socialRoomGenerator{centerItems: []string{"bench", "bench", "bench", "table"}, farItems: []string{"desk"}},
	"interrogation": &socialRoomGenerator{centerItems: []string{"table", "chair", "chair"}},
	"forensics":     &labStyleGenerator{sideItems: []string{"lab_bench", "evidence_locker"}, farItems: []string{"analysis_instrument"}},
}

// ApplyRoomGenerators runs the registered generator for each tagged room, carving
// sub-geometry and collecting placement hints. Returns a map of room index → hints.
func ApplyRoomGenerators(l *world.Level, z int, rooms []Room) map[int][]PlacementHint {
	hints := make(map[int][]PlacementHint)
	for i, room := range rooms {
		gen, ok := roomGenerators[room.Tag]
		if !ok {
			continue
		}
		if h := gen.Generate(l, room, z, room.DoorDir); len(h) > 0 {
			hints[i] = h
		}
	}
	return hints
}

// =============================================================================
// Helpers
// =============================================================================

type interiorBounds struct{ x1, y1, x2, y2 int }

func roomInterior(room Room) interiorBounds {
	return interiorBounds{
		x1: room.X + 1,
		y1: room.Y + 1,
		x2: room.X + room.Width - 2,
		y2: room.Y + room.Height - 2,
	}
}

// dirToFarWallRegion returns the wall region farthest from the door.
// doorDir points from the hallway into the room.
func dirToFarWallRegion(doorDir [2]int) PlacementRegion {
	switch {
	case doorDir[1] == -1:
		return RegionNorthWall
	case doorDir[1] == 1:
		return RegionSouthWall
	case doorDir[0] == -1:
		return RegionWestWall
	case doorDir[0] == 1:
		return RegionEastWall
	default:
		return RegionAnywhere
	}
}

// dirToNearWallRegion returns the wall region closest to the door (the door wall itself).
func dirToNearWallRegion(doorDir [2]int) PlacementRegion {
	switch dirToFarWallRegion(doorDir) {
	case RegionNorthWall:
		return RegionSouthWall
	case RegionSouthWall:
		return RegionNorthWall
	case RegionEastWall:
		return RegionWestWall
	case RegionWestWall:
		return RegionEastWall
	default:
		return RegionAnywhere
	}
}

// doorDirPerpendicularWalls returns the two wall regions perpendicular to the door axis.
func doorDirPerpendicularWalls(doorDir [2]int) (PlacementRegion, PlacementRegion) {
	if doorDir[1] != 0 { // door on north or south wall
		return RegionEastWall, RegionWestWall
	}
	return RegionNorthWall, RegionSouthWall
}

// hintsForItems creates PlacementHints distributing blueprints to regions in round-robin order.
func hintsForItems(blueprints []string, regions ...PlacementRegion) []PlacementHint {
	if len(regions) == 0 || len(blueprints) == 0 {
		return nil
	}
	hints := make([]PlacementHint, len(blueprints))
	for i, bp := range blueprints {
		hints[i] = PlacementHint{Blueprint: bp, Region: regions[i%len(regions)]}
	}
	return hints
}

// carvePartitionWall carves a partial wall dividing the room interior roughly 2/3
// from the door toward the far wall. Leaves a 1-tile passage gap at one end.
func carvePartitionWall(l *world.Level, ib interiorBounds, z int, doorDir [2]int) {
	iw := ib.x2 - ib.x1 + 1
	ih := ib.y2 - ib.y1 + 1
	if iw < 6 || ih < 5 {
		return
	}
	if doorDir[1] != 0 {
		var partY int
		if doorDir[1] == 1 {
			partY = ib.y1 + ih*2/3
		} else {
			partY = ib.y1 + ih/3
		}
		for x := ib.x1; x <= ib.x2-1; x++ {
			l.SetTileTypeAt(x, partY, z, world.TypeWall)
		}
	} else {
		var partX int
		if doorDir[0] == 1 {
			partX = ib.x1 + iw*2/3
		} else {
			partX = ib.x1 + iw/3
		}
		for y := ib.y1; y <= ib.y2-1; y++ {
			l.SetTileTypeAt(partX, y, z, world.TypeWall)
		}
	}
}

// =============================================================================
// Generator implementations
// =============================================================================

// bridgeGenerator carves a command platform ring and places consoles on the far wall.
type bridgeGenerator struct{}

func (g *bridgeGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	ib := roomInterior(room)
	iw := ib.x2 - ib.x1 + 1
	ih := ib.y2 - ib.y1 + 1

	if iw >= 5 && ih >= 5 {
		px1 := ib.x1 + 1
		py1 := ib.y1 + 1
		px2 := ib.x2 - 1
		py2 := ib.y2 - 1

		if px2-px1 >= 2 && py2-py1 >= 2 {
			for x := px1; x <= px2; x++ {
				for y := py1; y <= py2; y++ {
					if x == px1 || x == px2 || y == py1 || y == py2 {
						l.SetTileTypeAt(x, y, z, world.TypeWall)
					} else {
						l.SetTileTypeAt(x, y, z, world.TypeFloor)
					}
				}
			}
			midX := (px1 + px2) / 2
			midY := (py1 + py2) / 2
			l.SetTileTypeAt(midX, py1, z, world.TypeFloor)
			l.SetTileTypeAt(midX, py2, z, world.TypeFloor)
			l.SetTileTypeAt(px1, midY, z, world.TypeFloor)
			l.SetTileTypeAt(px2, midY, z, world.TypeFloor)
		}
	}

	far := dirToFarWallRegion(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems([]string{"captains_chair", "map_table"}, RegionCenter)...)
	hints = append(hints, hintsForItems([]string{"console_bank", "tactical_display", "comm_array"}, far)...)
	return hints
}

// sleepingRoomGenerator divides the room with an optional partition wall,
// places beds in the far alcove and storage items near the entrance.
type sleepingRoomGenerator struct {
	hasPartition bool
	bedBlueprint string
	farItems     []string
	nearItems    []string
}

func (g *sleepingRoomGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	ib := roomInterior(room)
	if g.hasPartition {
		carvePartitionWall(l, ib, z, doorDir)
	}
	far := dirToFarWallRegion(doorDir)
	near := dirToNearWallRegion(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems([]string{g.bedBlueprint}, far)...)
	hints = append(hints, hintsForItems(g.farItems, far)...)
	hints = append(hints, hintsForItems(g.nearItems, near)...)
	return hints
}

// labStyleGenerator places benches on the two side walls (perpendicular to the door axis)
// and specialty equipment on the far wall, creating rows with an open central aisle.
type labStyleGenerator struct {
	sideItems []string
	farItems  []string
}

func (g *labStyleGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	ib := roomInterior(room)
	iw := ib.x2 - ib.x1 + 1
	ih := ib.y2 - ib.y1 + 1
	if iw < 4 || ih < 4 {
		return nil
	}
	sideA, sideB := doorDirPerpendicularWalls(doorDir)
	far := dirToFarWallRegion(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems(g.sideItems, sideA, sideB)...)
	hints = append(hints, hintsForItems(g.farItems, far)...)
	return hints
}

// controlRoomGenerator places equipment on the far wall, optional center pieces,
// side items, and items near the entrance.
type controlRoomGenerator struct {
	farItems    []string
	centerItems []string
	sideItems   []string
	nearItems   []string
}

func (g *controlRoomGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	far := dirToFarWallRegion(doorDir)
	near := dirToNearWallRegion(doorDir)
	sideA, sideB := doorDirPerpendicularWalls(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems(g.farItems, far)...)
	hints = append(hints, hintsForItems(g.centerItems, RegionCenter)...)
	hints = append(hints, hintsForItems(g.sideItems, sideA, sideB)...)
	hints = append(hints, hintsForItems(g.nearItems, near)...)
	return hints
}

// storageRoomGenerator distributes items around all four walls and optionally
// places items in the center (e.g. crates, racks).
type storageRoomGenerator struct {
	wallItems   []string
	centerItems []string
}

func (g *storageRoomGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	var hints []PlacementHint
	hints = append(hints, hintsForItems(g.wallItems, RegionNorthWall, RegionSouthWall, RegionEastWall, RegionWestWall)...)
	hints = append(hints, hintsForItems(g.centerItems, RegionCenter)...)
	return hints
}

// rowBedsGenerator places beds on the two side walls (creating parallel rows),
// equipment on the far wall, and items near the entrance.
type rowBedsGenerator struct {
	bedBlueprint string
	farItems     []string
	nearItems    []string
}

func (g *rowBedsGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	sideA, sideB := doorDirPerpendicularWalls(doorDir)
	far := dirToFarWallRegion(doorDir)
	near := dirToNearWallRegion(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems([]string{g.bedBlueprint, g.bedBlueprint}, sideA, sideB)...)
	hints = append(hints, hintsForItems(g.farItems, far)...)
	hints = append(hints, hintsForItems(g.nearItems, near)...)
	return hints
}

// workshopGenerator places workbenches/racks on the side walls, specialty equipment
// on the far wall, and utility items near the entrance.
type workshopGenerator struct {
	sideItems []string
	farItems  []string
	nearItems []string
}

func (g *workshopGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	sideA, sideB := doorDirPerpendicularWalls(doorDir)
	far := dirToFarWallRegion(doorDir)
	near := dirToNearWallRegion(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems(g.sideItems, sideA, sideB)...)
	hints = append(hints, hintsForItems(g.farItems, far)...)
	hints = append(hints, hintsForItems(g.nearItems, near)...)
	return hints
}

// socialRoomGenerator places seating and tables in the center, service equipment
// on the far wall, decorative items on side walls, and staff items near the entrance.
type socialRoomGenerator struct {
	centerItems []string
	farItems    []string
	sideItems   []string
	nearItems   []string
}

func (g *socialRoomGenerator) Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint {
	far := dirToFarWallRegion(doorDir)
	near := dirToNearWallRegion(doorDir)
	sideA, sideB := doorDirPerpendicularWalls(doorDir)
	var hints []PlacementHint
	hints = append(hints, hintsForItems(g.centerItems, RegionCenter)...)
	hints = append(hints, hintsForItems(g.farItems, far)...)
	hints = append(hints, hintsForItems(g.sideItems, sideA, sideB)...)
	hints = append(hints, hintsForItems(g.nearItems, near)...)
	return hints
}
