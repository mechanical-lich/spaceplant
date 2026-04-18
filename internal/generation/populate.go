package generation

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// roomFurniture maps a room tag to the blueprint names that can appear in it.
// Each entry is a slice of blueprint names; a random subset will be placed.
var roomFurniture = map[string][]string{
	"crew_quarters": {"bunk_bed", "bunk_bed", "locker", "locker", "desk"},
	"officers_suite": {"single_bed", "wardrobe", "desk", "couch"},
	"captains_quarters": {"single_bed", "desk", "wardrobe", "safe", "bookshelf"},
	"guest_cabin": {"single_bed", "folding_table", "locker"},
	"family_apartment": {"bunk_bed", "single_bed", "kitchenette", "table", "chair", "chair", "storage_unit"},
	"childcare": {"sleeping_cot", "sleeping_cot", "storage_unit", "table"},
	"laundry": {"storage_unit", "sorting_bin", "folding_table"},

	"bridge": {"captains_chair", "console_bank", "console_bank", "tactical_display", "comm_array", "map_table"},
	"mission_control": {"operator_console", "operator_console", "operator_console", "holo_display", "table"},
	"comms_relay": {"antenna_rack", "signal_processor", "signal_processor", "operator_console"},
	"navigation": {"navigation_console", "map_table", "holo_display"},
	"docking_control": {"docking_monitor", "docking_monitor", "status_light", "control_panel"},
	"security_office": {"weapons_locker", "monitoring_wall", "evidence_locker", "desk"},
	"executive_suite": {"single_bed", "desk", "couch", "holo_display", "safe"},

	"reactor_core": {"containment_vessel", "control_panel", "coolant_pipe", "radiation_shielding", "catwalk"},
	"engineering_workshop": {"workbench", "workbench", "tool_rack", "welding_rig", "parts_bin"},
	"maintenance_bay": {"tool_cart", "tool_rack", "diagnostic_panel", "maintenance_platform", "shelving"},
	"life_support_control": {"environmental_panel", "control_panel", "control_panel"},
	"water_waste_processing": {"filtration_tank", "filtration_tank", "diagnostic_panel"},
	"cargo_hold": {"crate", "crate", "crate", "pallet", "pallet_rack", "cargo_net"},
	"fuel_storage": {"fuel_tank", "fuel_tank", "pressure_gauge", "control_panel"},
	"eva_bay": {"suit_rack", "suit_rack", "airlock_controls", "tool_rack"},
	"utility_corridor": {"junction_box", "coolant_pipe", "diagnostic_panel"},

	"manufacturing_hangar": {"workbench", "crane", "welding_rig", "tool_rack", "spare_parts_shelf"},
	"robotics_bay": {"charging_station", "charging_station", "workbench", "tool_rack"},
	"freight_airlock": {"airlock_controls", "crate", "crate", "pallet"},
	"freight_sorting": {"conveyor_belt", "sorting_bin", "sorting_bin", "crate", "crate"},
	"storage_vault": {"secure_locker", "secure_locker", "shelving", "safe"},
	"customs_inspection": {"desk", "secure_locker", "evidence_locker"},

	"mess_hall": {"table", "table", "bench", "bench", "serving_line", "kitchenette"},
	"bar_cantina": {"bar_counter", "bar_stool", "bar_stool", "bar_stool", "table"},
	"market": {"shelving", "shelving", "table", "storage_unit"},
	"shop": {"shelving", "workbench", "storage_unit", "secure_locker"},
	"recreation": {"table", "table", "bench", "bench", "bookshelf"},
	"library": {"bookshelf", "bookshelf", "bookshelf", "table", "chair"},
	"bank": {"safe", "secure_locker", "desk"},
	"administration": {"desk", "desk", "filing_cabinet", "table"},
	"meditation": {"bench", "bench", "planter", "planter"},

	"general_lab": {"lab_bench", "lab_bench", "sample_rack", "analysis_instrument", "fume_hood"},
	"biolab": {"bio_cabinet", "specimen_tank", "specimen_tank", "incubator", "decontamination_shower"},
	"chemistry_lab": {"lab_bench", "reagent_cabinet", "fume_hood", "centrifuge"},
	"fabrication_lab": {"3d_printer", "workbench", "tool_rack", "spare_parts_shelf"},
	"observatory": {"telescope_mount", "sensor_console", "holo_display"},
	"medical_research": {"lab_bench", "centrifuge", "microscope", "specimen_freezer"},
	"hydroponics_lab": {"grow_rack", "grow_rack", "nutrient_tank", "nutrient_tank"},
	"data_center": {"server_rack", "server_rack", "server_rack", "diagnostic_panel"},

	"medical_bay": {"exam_bed", "exam_bed", "vitals_monitor", "diagnostic_console", "drug_cabinet"},
	"surgery": {"operating_table", "surgical_light", "sterilization_unit", "anesthesia_console"},
	"quarantine": {"exam_bed", "decontamination_shower", "ppe_station"},
	"pharmacy": {"drug_cabinet", "drug_cabinet", "shelving"},
	"morgue": {"morgue_slab", "morgue_slab", "specimen_freezer"},
	"cryo": {"cryo_pod", "cryo_pod", "cryo_pod", "monitoring_wall"},

	"brig": {"bench", "bench", "secure_locker"},
	"courtroom": {"bench", "bench", "bench", "table", "desk"},
	"interrogation": {"table", "chair", "chair"},
	"forensics": {"lab_bench", "evidence_locker", "analysis_instrument"},

	"life_pod_bay":       {"life_pod_console", "suit_rack", "suit_rack", "airlock_controls"},
	"self_destruct_room": {"self_destruct_console", "radiation_shielding", "control_panel"},
}

// PopulateRooms places furniture entities inside each tagged room.
// It is called after GenerateFloors returns FloorResults.
func PopulateRooms(l *world.Level, results []FloorResult) {
	for _, fr := range results {
		for i, room := range fr.Rooms {
			if room.Tag == "" {
				continue
			}
			populateRoom(l, fr.Z, room, fr.PlacementHints[i])
		}
	}
}

// populateRoom places a random selection of furniture from the tag's list
// into the interior tiles of the room (avoiding walls).
// hints constrains where specific blueprints are placed; nil hints = fully random.
func populateRoom(l *world.Level, z int, room Room, hints []PlacementHint) {
	blueprints, ok := roomFurniture[room.Tag]
	if !ok || len(blueprints) == 0 {
		return
	}

	// Interior bounds (inset by 1 to avoid walls).
	x1 := room.X + 1
	y1 := room.Y + 1
	x2 := room.X + room.Width - 2
	y2 := room.Y + room.Height - 2

	if x2 <= x1 || y2 <= y1 {
		return // room too small
	}

	// Place up to half the room's interior area in furniture, capped at blueprint count.
	interiorArea := (x2 - x1 + 1) * (y2 - y1 + 1)
	maxItems := interiorArea / 4
	if maxItems > len(blueprints) {
		maxItems = len(blueprints)
	}
	if maxItems == 0 {
		maxItems = 1
	}
	numItems := utility.GetRandom(1, maxItems+1)

	// Build a blueprint→region lookup from hints.
	// Hinted blueprints are placed first (guaranteed); remaining slots filled randomly.
	hintMap := make(map[string]PlacementRegion, len(hints))
	hintOrder := make([]string, 0, len(hints))
	hintSet := make(map[string]bool, len(hints))
	for _, h := range hints {
		hintMap[h.Blueprint] = h.Region
		hintOrder = append(hintOrder, h.Blueprint)
		hintSet[h.Blueprint] = true
	}

	// Non-hinted blueprints shuffled for random fill.
	var rest []string
	for _, bp := range blueprints {
		if !hintSet[bp] {
			rest = append(rest, bp)
		}
	}
	utility.Shuffle(rest)

	ordered := append(hintOrder, rest...)

	placed := 0
	for _, bp := range ordered {
		if placed >= numItems {
			break
		}

		region := RegionAnywhere
		if r, ok := hintMap[bp]; ok {
			region = r
		}

		tx, ty := pickPositionInRegion(x1, y1, x2, y2, region)
		if tx < 0 {
			tx = utility.GetRandom(x1, x2+1)
			ty = utility.GetRandom(y1, y2+1)
		}

		tile := l.Level.GetTilePtr(tx, ty, z)
		if tile == nil || tile.IsSolid() {
			continue
		}
		if l.Level.GetEntityAt(tx, ty, z) != nil {
			continue
		}
		if adjacentToDoor(l, tx, ty, z) {
			continue
		}

		e, err := factory.Create(bp, tx, ty)
		if err != nil {
			// Blueprint doesn't exist yet — skip silently.
			continue
		}
		e.GetComponent("Position").(*component.PositionComponent).SetPosition(tx, ty, z)
		l.Level.AddEntity(e)
		placed++
	}
}

// pickPositionInRegion returns a random position within the given region of the interior
// bounds [x1,x2] x [y1,y2]. Returns (-1,-1) if the region is degenerate.
func pickPositionInRegion(x1, y1, x2, y2 int, region PlacementRegion) (int, int) {
	var rx1, ry1, rx2, ry2 int
	switch region {
	case RegionNorthWall:
		rx1, ry1, rx2, ry2 = x1, y1, x2, y1
	case RegionSouthWall:
		rx1, ry1, rx2, ry2 = x1, y2, x2, y2
	case RegionWestWall:
		rx1, ry1, rx2, ry2 = x1, y1, x1, y2
	case RegionEastWall:
		rx1, ry1, rx2, ry2 = x2, y1, x2, y2
	case RegionCenter:
		iw := x2 - x1 + 1
		ih := y2 - y1 + 1
		rx1 = x1 + iw/3
		rx2 = x1 + 2*iw/3
		ry1 = y1 + ih/3
		ry2 = y1 + 2*ih/3
		if rx2 < rx1 {
			rx2 = rx1
		}
		if ry2 < ry1 {
			ry2 = ry1
		}
	default: // RegionAnywhere
		rx1, ry1, rx2, ry2 = x1, y1, x2, y2
	}
	if rx2 < rx1 || ry2 < ry1 {
		return -1, -1
	}
	return utility.GetRandom(rx1, rx2+1), utility.GetRandom(ry1, ry2+1)
}

// adjacentToDoor returns true if any of the four cardinal neighbors of (tx,ty,z)
// has a door entity, keeping a 1-tile clearance in front of every door.
func adjacentToDoor(l *world.Level, tx, ty, z int) bool {
	cardinals := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	var buf []*ecs.Entity
	for _, d := range cardinals {
		buf = buf[:0]
		l.Level.GetEntitiesAt(tx+d[0], ty+d[1], z, &buf)
		for _, e := range buf {
			if e.HasComponent(component.Door) {
				return true
			}
		}
	}
	return false
}
