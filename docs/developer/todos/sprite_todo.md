# Sprite TODO

Items that are currently using placeholder sprites or have no sprite at all and need original art.

Weapons with real sprites use the `weapons` sheet. All named sprite positions below are from blueprint `AppearanceComponent` values.

---

## Weapons — Ranged

All currently on the `weapons` sheet with placeholder positions (borrowing existing weapon sprites).

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `pulse_rifle` | M41A Pulse Rifle | Colonial Marine bullpup rifle. Caseless drum mag, digital round counter. |
| `smartgun` | M56 Smartgun | Body-mounted gyro-stabilized MG. Arm brace, large drum, optical sight. |
| `tactical_shotgun` | Tactical Shotgun | Military pump-action. Polymer frame, extended tube, pistol grip. |
| `heavy_pistol` | Heavy Pistol | Large-frame revolver-style .44. Wide barrel, chunky grip. |
| `rivet_gun` | Rivet Gun | Industrial tool. Boxy, yellow/grey body, wide muzzle. |

---

## Weapons — Melee

All currently on the `weapons` sheet with placeholder positions.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `combat_knife` | Combat Knife | Fast, light blade. Slashing. |
| `wrench` | Wrench | Maintenance tool. Bludgeoning. |
| `crowbar` | Crowbar | Heavy pry bar. Bludgeoning. |
| `fire_axe` | Fire Axe | Large axe, red handle. Slashing. |
| `stun_baton` | Stun Baton | Electric baton, glowing tip. Electric damage. Applies slowed on hit. |
| `plasma_cutter` | Plasma Cutter | Heavy industrial cutter. Wide grip, glowing pink/magenta beam emitter. |
| `electric_prod` | Electric Prod | Long pole with pronged tip, yellow arc markings. |

---

## Ammo

All three ammo types currently share `SpriteX: 0` on the `ammo` sheet — 5.56 and 12g need unique sprites.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `5_56_rounds` | 5.56 Magazine | Rifle magazine, rectangular. Currently shares 9mm sprite. |
| `12g_shells` | 12g Shells | Shotgun shell tube, orange/red. Currently shares 9mm sprite. |
| `10mm_caseless_drum` | 10mm Caseless Drum | Wide cylindrical drum mag, grey/blue. Currently shares 9mm sprite. |
| `44_rounds` | .44 Speedloader | Circular speedloader with large rounds. Currently shares 9mm sprite. |
| `rivet_pack` | Rivet Pack | Strip/coil of steel rivets, industrial yellow. Currently shares 9mm sprite. |
| `12g_slugs` | 12g Slugs | Tube of solid slugs, darker than buckshot shells. Currently shares 9mm sprite. |

---

## Armor

New armor items use placeholder positions borrowed from existing armor sheet sprites.

| Blueprint ID | Display Name | Slot | Notes |
|---|---|---|---|
| `tactical_vest` | Tactical Vest | torso | Plate carrier with front/back trauma plates. OD green or tan. | <done>
| `marine_armor` | Marine Battle Armor | torso | Full composite plates, Colonial Marine markings, dark olive. | <done>
| `combat_trousers` | Combat Trousers | legs | Armored knee/thigh panels over canvas. Matches marine armor palette. | <done>
| `marine_helmet` | Marine Helmet | head | Composite combat helmet, integrated HUD rail, visor mount. | <done>
| `eva_helmet` | EVA Helmet | head | Sealed bubble visor, white/grey, built-in lamp on forehead. | <done>
| `combat_boots` | Combat Boots | feet | Heavy steel-toe composite boots, black. Matches marine armor. | <done>

---

## Misc Items

| Blueprint ID | Display Name | Resource | Notes |
|---|---|---|---|
| `blue_keycard` | Blue Keycard | `keys` | Card with blue stripe. Sheet likely needs more keycard colors. |
| `health` | Med Kit | `items` | Red cross medical kit. |

---

## Environment — Furniture

No sprites exist. Needs a `furniture` sprite sheet (or additions to `environment` sheet). All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `bunk_bed` | Bunk Bed | Two-tier bed, metal frame. |
| `single_bed` | Bed | Single mattress, white/blue. |
| `sleeping_cot` | Sleeping Cot | Simple folding cot. |
| `desk` | Desk | Flat work surface with small monitor. |
| `locker` | Locker | Tall narrow metal locker. |
| `wardrobe` | Wardrobe | Wider wooden cabinet. |
| `couch` | Couch | Two-seat sofa, cushioned. |
| `table` | Table | Generic flat table. |
| `bench` | Bench | Long seating bench, no back. |
| `chair` | Chair | Simple upright chair. |
| `bar_stool` | Bar Stool | Tall stool, round seat. |
| `bookshelf` | Bookshelf | Tall shelf full of books/binders. |
| `storage_unit` | Storage Unit | Closed cabinet with drawers. |
| `folding_table` | Folding Table | Lightweight collapsible table. |
| `kitchenette` | Kitchenette | Small counter with sink/hotplate. |
| `privacy_curtain` | Privacy Curtain | Hanging curtain rail. |
| `porthole` | Porthole | Round window, stars visible. |
| `bar_counter` | Bar Counter | Low bar surface, wood-toned. |
| `serving_line` | Serving Line | Cafeteria-style food counter. |
| `planter` | Planter | Square pot with green plant. |

---

## Environment — Consoles & Displays

No sprites exist. Needs a `consoles` sprite sheet or additions to `environment`. All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `console_bank` | Console Bank | Wide multi-screen workstation. |
| `tactical_display` | Tactical Display | Vertical screen with grid/map. |
| `navigation_console` | Navigation Console | Angled console with star chart. |
| `comm_array` | Comm Array | Dish/antenna emitter unit. |
| `captains_chair` | Captain's Chair | High-backed command seat. |
| `map_table` | Map Table | Flat holographic table display. |
| `holo_display` | Holo-Display | Upright hologram emitter. |
| `control_panel` | Control Panel | Generic button/switch panel. |
| `docking_monitor` | Docking Monitor | Screen showing bay status. |
| `signal_processor` | Signal Processor | Rack-mounted signal unit. |
| `antenna_rack` | Antenna Rack | Vertical rack with small dishes. |
| `status_light` | Status Light | Indicator light column. |
| `junction_box` | Junction Box | Electrical panel, yellow markings. |
| `operator_console` | Operator Console | Standard seated workstation. |
| `environmental_panel` | Environmental Control Panel | Air/life-support readout. |
| `monitoring_wall` | Monitoring Wall | Multi-screen wall array. |

---

## Environment — Lab Equipment

No sprites exist. Needs a `lab` sprite sheet or additions to `environment`. All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `lab_bench` | Lab Bench | Long white bench with equipment. |
| `specimen_tank` | Specimen Tank | Glass tank with green liquid. |
| `fume_hood` | Fume Hood | Ventilated enclosure, transparent front. |
| `centrifuge` | Centrifuge | Cylindrical spinning device. |
| `microscope` | Microscope | Classic scope on stand. |
| `server_rack` | Server Rack | Tall black rack with blinking lights. |
| `grow_rack` | Grow Rack | Shelf with glowing plant trays. |
| `containment_cabinet` | Containment Cabinet | Orange biohazard cabinet. |
| `sample_rack` | Sample Rack | Test tube/vial rack. |
| `3d_printer` | 3D Printer | Boxy printer with open frame. |
| `telescope_mount` | Telescope Mount | Large scope on pivot. |
| `reagent_cabinet` | Reagent Cabinet | Cabinet with chemical bottles. |
| `incubator` | Incubator | Warm chamber with racks inside. |
| `nutrient_tank` | Nutrient Tank | Green-tinted cylindrical tank. |
| `analysis_instrument` | Analysis Instrument | Generic scientific analyzer. |

---

## Environment — Medical

No sprites exist. Needs a `medical` sprite sheet or additions to `environment`. All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `exam_bed` | Exam Bed | Blue padded examination table. |
| `operating_table` | Operating Table | Surgical table with arm extensions. |
| `surgical_light` | Surgical Light | Overhead articulated lamp. |
| `vitals_monitor` | Vitals Monitor | Screen with waveform readout. |
| `cryo_pod` | Cryo Pod | Upright frost-covered capsule. |
| `life_support_rig` | Life Support Rig | Mechanical support frame with tubes. |
| `decontamination_shower` | Decontamination Shower | Standing shower enclosure. |
| `drug_cabinet` | Drug Cabinet | White wall cabinet with lock. |
| `diagnostic_console` | Diagnostic Console | Medical-specific workstation. |
| `specimen_freezer` | Specimen Freezer | White freezer unit. |
| `sterilization_unit` | Sterilization Unit | Autoclave/UV sterilizer box. |
| `anesthesia_console` | Anesthesia Console | Gas/IV delivery unit. |
| `morgue_slab` | Morgue Slab | Cold metal table, clinical grey. |
| `ppe_station` | PPE Station | Wall-mounted gown/glove dispenser. |

---

## Environment — Industrial

No sprites exist. Needs an `industrial` sprite sheet or additions to `environment`. All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `reactor_core` | Reactor Core | Glowing orange cylindrical core. |
| `workbench` | Workbench | Heavy wood/metal work surface. |
| `welding_rig` | Welding Rig | Torch stand with tank. |
| `hoist` | Hoist | Chain/cable lifting block. |
| `conveyor_belt` | Conveyor Belt | Moving belt section. |
| `crane` | Crane | Overhead lifting arm. |
| `airlock_controls` | Airlock Controls | Panel with red/green cycle buttons. |
| `coolant_pipe` | Coolant Pipe | Blue insulated pipe section. |
| `catwalk` | Catwalk | Metal grating walkway section. |
| `suit_rack` | EVA Suit Rack | Wall mount with EVA suit hanging. |
| `pressure_gauge` | Pressure Gauge | Round dial gauge. |
| `fuel_tank` | Fuel Tank | Orange cylindrical storage tank. |
| `filtration_tank` | Filtration Tank | Blue-grey tank with piping. |
| `tool_cart` | Tool Cart | Rolling cart with tools. |
| `charging_station` | Charging Station | Floor unit with green charge ports. |
| `turret_mount` | Turret Mount | Fixed gun emplacement base. |
| `diagnostic_panel` | Diagnostic Panel | Wall-mounted readout panel. |
| `radiation_shielding` | Radiation Shielding | Yellow/black striped panel. |
| `containment_vessel` | Containment Vessel | Sealed glowing sphere/drum. |
| `maintenance_platform` | Maintenance Platform | Raised grating platform section. |

---

## Environment — Storage

No sprites exist. Needs a `storage` sprite sheet or additions to `environment`. All 32×48.

| Blueprint ID | Display Name | Notes |
|---|---|---|
| `crate` | Crate | Wooden/metal shipping crate. |
| `pallet` | Pallet | Flat wooden pallet. |
| `shelving` | Shelving | Open metal shelf unit. |
| `parts_bin` | Parts Bin | Open-top bin with small parts. |
| `cargo_net` | Cargo Net | Rope/strap net bundle. |
| `secure_locker` | Secure Locker | Reinforced locked cabinet. |
| `safe` | Safe | Heavy steel safe, dial lock. |
| `vault_door` | Vault Door | Large armored door panel. |
| `sorting_bin` | Sorting Bin | Color-coded open container. |
| `tool_rack` | Tool Rack | Wall-mounted tools on pegs. |
| `weapons_locker` | Weapons Locker | Secure red-marked cabinet. |
| `evidence_locker` | Evidence Locker | Tagged sealed container cabinet. |
| `spare_parts_shelf` | Spare Parts Shelf | Shelf with mechanical parts. |
| `pallet_rack` | Pallet Rack | Tall warehouse racking. |
