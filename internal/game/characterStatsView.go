package game

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

const (
	csModalW = 700
	csModalH = 490
)

// CharacterStatsView is the Shift+I character sheet modal.
type CharacterStatsView struct {
	Visible bool
	player  *ecs.Entity

	modal *minui.Modal
	tabs  *minui.TabPanel

	// Tab 1 — Overview
	overviewArea *minui.ScrollingTextArea

	// Tab 2 — Equipment
	equipList       *minui.ListBox
	equipImg        *minui.ImageWidget
	equipDesc       *minui.ScrollingTextArea
	equipUnequipBtn *minui.Button
	equipIDs        []string // parallel to list items (slot names)

	// Tab 3 — Skills
	skillList *minui.ListBox
	skillDesc *minui.ScrollingTextArea
	skillIDs  []string

	// Tab 4 — Inventory
	invList            *minui.ListBox
	invNearbyList      *minui.ListBox
	invImg             *minui.ImageWidget
	invDesc            *minui.ScrollingTextArea
	invEquip           *minui.Button
	invDrop            *minui.Button
	invTake            *minui.Button
	invGive            *minui.Button
	invNearbyLabel     *minui.Label
	invItems           []*ecs.Entity
	invNearbyItems     []*ecs.Entity
	invNearbyItemSlots []string // parallel to invNearbyItems: body-part slot name if equipped, "" if in bag
	invNearbyEntity    *ecs.Entity
	// which list was last clicked: "player" or "nearby"
	invFocus string
}

func NewCharacterStatsView(player *ecs.Entity) *CharacterStatsView {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - csModalW) / 2
	my := (cfg.ScreenHeight - csModalH) / 2

	v := &CharacterStatsView{player: player}

	v.modal = minui.NewModal("cs_modal", "Character Sheet", csModalW, csModalH)
	v.modal.SetPosition(mx, my)
	v.modal.Closeable = true
	v.modal.OnClose = func() { v.Visible = false }

	v.tabs = minui.NewTabPanel("cs_tabs", csModalW-20, csModalH-80)
	v.tabs.SetPosition(10, 10)
	v.modal.AddChild(v.tabs)

	v.buildOverviewTab()
	v.buildEquipmentTab()
	v.buildSkillsTab()
	v.buildInventoryTab()

	return v
}

func (v *CharacterStatsView) buildOverviewTab() {
	panel := minui.NewPanel("cs_overview_panel")
	panel.SetPosition(0, v.tabs.TabHeight)
	panel.SetSize(csModalW-20, csModalH-80-v.tabs.TabHeight)

	v.overviewArea = minui.NewScrollingTextArea("cs_overview", csModalW-40, csModalH-80-v.tabs.TabHeight-20)
	v.overviewArea.SetPosition(10, 10)
	v.overviewArea.LineHeight = 16
	panel.AddChild(v.overviewArea)

	v.tabs.AddTab("overview", "Overview", panel)
}

func (v *CharacterStatsView) buildEquipmentTab() {
	panelH := csModalH - 80 - v.tabs.TabHeight
	panel := minui.NewPanel("cs_equip_panel")
	panel.SetPosition(0, v.tabs.TabHeight)
	panel.SetSize(csModalW-20, panelH)

	const btnH = 28
	const btnGap = 8
	listH := panelH - 20 - btnH - btnGap

	v.equipList = minui.NewListBox("cs_equip_list", []string{})
	v.equipList.SetPosition(10, 10)
	v.equipList.SetSize(200, listH)
	v.equipList.Layout()
	v.equipList.OnSelect = func(idx int, _ string) { v.refreshEquipDesc(idx) }
	panel.AddChild(v.equipList)

	v.equipUnequipBtn = minui.NewButton("cs_equip_unequip", "Unequip")
	v.equipUnequipBtn.SetPosition(10, 10+listH+btnGap)
	v.equipUnequipBtn.SetSize(200, btnH)
	v.equipUnequipBtn.OnClick = func() { v.onEquipUnequip() }
	panel.AddChild(v.equipUnequipBtn)

	const imgSize = 64
	const imgGap = 10

	v.equipImg = minui.NewImageWidget("cs_equip_img", imgSize, imgSize)
	v.equipImg.SetPosition(220, 10)
	panel.AddChild(v.equipImg)

	v.equipDesc = minui.NewScrollingTextArea("cs_equip_desc", csModalW-240, panelH-20-imgSize-imgGap)
	v.equipDesc.SetPosition(220, 10+imgSize+imgGap)
	v.equipDesc.LineHeight = 14
	panel.AddChild(v.equipDesc)

	v.tabs.AddTab("equipment", "Equipment", panel)
}

func (v *CharacterStatsView) buildSkillsTab() {
	panel := minui.NewPanel("cs_skills_panel")
	panel.SetPosition(0, v.tabs.TabHeight)
	panel.SetSize(csModalW-20, csModalH-80-v.tabs.TabHeight)

	v.skillList = minui.NewListBox("cs_skill_list", []string{})
	v.skillList.SetPosition(10, 10)
	v.skillList.SetSize(200, csModalH-80-v.tabs.TabHeight-20)
	v.skillList.Layout()
	v.skillList.OnSelect = func(idx int, _ string) { v.refreshSkillDesc(idx) }
	panel.AddChild(v.skillList)

	v.skillDesc = minui.NewScrollingTextArea("cs_skill_desc", csModalW-240, csModalH-80-v.tabs.TabHeight-20)
	v.skillDesc.SetPosition(220, 10)
	v.skillDesc.LineHeight = 14
	panel.AddChild(v.skillDesc)

	v.tabs.AddTab("skills", "Skills", panel)
}

func (v *CharacterStatsView) buildInventoryTab() {
	panelH := csModalH - 80 - v.tabs.TabHeight
	panel := minui.NewPanel("cs_inv_panel")
	panel.SetPosition(0, v.tabs.TabHeight)
	panel.SetSize(csModalW-20, panelH)

	const listW = 170
	const midX = listW + 20 // left edge of center column
	const midW = csModalW - 20 - listW*2 - 40
	const nearbyX = csModalW - 20 - listW - 10

	const imgSizeW = 64
	const imgSizeH = 96
	const imgGap = 10
	const btnH = 28
	btnY := panelH - btnH - 10

	// --- Player label + list ---
	playerLbl := minui.NewLabel("cs_inv_player_lbl", "Your Inventory")
	playerLbl.SetPosition(10, 10)
	playerLbl.SetSize(listW, 18)
	panel.AddChild(playerLbl)

	v.invList = minui.NewListBox("cs_inv_list", []string{})
	v.invList.SetPosition(10, 30)
	v.invList.SetSize(listW, btnY-34)
	v.invList.Layout()
	v.invList.OnSelect = func(idx int, _ string) {
		v.invFocus = "player"
		v.invNearbyList.SelectedIndex = -1
		v.refreshInvDesc(idx)
	}
	panel.AddChild(v.invList)

	// Equip/Use and Drop under player list
	v.invEquip = minui.NewButton("cs_inv_equip", "Equip/Use")
	v.invEquip.SetPosition(10, btnY)
	v.invEquip.SetSize(80, btnH)
	v.invEquip.OnClick = v.onInvEquip
	panel.AddChild(v.invEquip)

	v.invDrop = minui.NewButton("cs_inv_drop", "Drop")
	v.invDrop.SetPosition(95, btnY)
	v.invDrop.SetSize(listW-85, btnH)
	v.invDrop.OnClick = v.onInvDrop
	panel.AddChild(v.invDrop)

	// --- Centre: image + description ---
	v.invImg = minui.NewImageWidget("cs_inv_img", imgSizeW, imgSizeH)
	v.invImg.SetPosition(midX+(midW-imgSizeW)/2, 10)
	panel.AddChild(v.invImg)

	v.invDesc = minui.NewScrollingTextArea("cs_inv_desc", midW, panelH-20-imgSizeH-imgGap)
	v.invDesc.SetPosition(midX, 10+imgSizeH+imgGap)
	v.invDesc.LineHeight = 14
	panel.AddChild(v.invDesc)

	// Take / Give transfer buttons in the centre column
	v.invTake = minui.NewButton("cs_inv_take", "← Take")
	v.invTake.SetPosition(midX, btnY)
	v.invTake.SetSize((midW/2)-2, btnH)
	v.invTake.OnClick = v.onInvTake
	v.invTake.SetVisible(false)
	panel.AddChild(v.invTake)

	v.invGive = minui.NewButton("cs_inv_give", "Give →")
	v.invGive.SetPosition(midX+midW/2+2, btnY)
	v.invGive.SetSize((midW/2)-2, btnH)
	v.invGive.OnClick = v.onInvGive
	v.invGive.SetVisible(false)
	panel.AddChild(v.invGive)

	// --- Nearby entity label + list ---
	v.invNearbyLabel = minui.NewLabel("cs_inv_nearby_lbl", "")
	v.invNearbyLabel.SetPosition(nearbyX, 10)
	v.invNearbyLabel.SetSize(listW, 18)
	v.invNearbyLabel.SetVisible(false)
	panel.AddChild(v.invNearbyLabel)

	v.invNearbyList = minui.NewListBox("cs_inv_nearby_list", []string{})
	v.invNearbyList.SetPosition(nearbyX, 30)
	v.invNearbyList.SetSize(listW, btnY-34)
	v.invNearbyList.Layout()
	v.invNearbyList.OnSelect = func(idx int, _ string) {
		v.invFocus = "nearby"
		v.invList.SelectedIndex = -1
		v.refreshNearbyDesc(idx)
	}
	v.invNearbyList.SetVisible(false)
	panel.AddChild(v.invNearbyList)

	v.tabs.AddTab("inventory", "Inventory", panel)
}

// Open refreshes all tabs and shows the modal.
func (v *CharacterStatsView) Open() {
	v.refreshOverview()
	v.refreshEquipmentList()
	v.refreshSkillList()
	v.refreshInventoryList()
	v.modal.SetVisible(true)
	v.Visible = true
}

// OpenToInventory opens the modal with the Inventory tab active.
func (v *CharacterStatsView) OpenToInventory() {
	v.Open()
	v.tabs.SetActiveTab("inventory")
}

// -----------------------------------------------------------------------
// Inventory tab helpers
// -----------------------------------------------------------------------

func (v *CharacterStatsView) refreshInventoryList() {
	bag := playerBag(v.player)
	v.invItems = make([]*ecs.Entity, 0, len(bag))
	names := make([]string, 0, len(bag))
	for _, item := range bag {
		v.invItems = append(v.invItems, item)
		names = append(names, itemName(item))
	}
	v.invList.SetItems(names)
	v.invDesc.Clear()
	v.invDesc.AddText("Select an item.")
}

func (v *CharacterStatsView) refreshInvDesc(idx int) {
	v.invDesc.Clear()
	v.invImg.Image = nil
	if idx < 0 || idx >= len(v.invItems) {
		return
	}
	item := v.invItems[idx]
	v.invDesc.AddText(itemName(item))
	if desc := itemDesc(item); desc != "" {
		v.invDesc.AddText("")
		for _, line := range mlge_text.Wrap(desc, 38, 0) {
			v.invDesc.AddText(line)
		}
	}

	// Set item sprite image
	v.invImg.Image = itemSpriteImage(item)
	if item.HasComponent(component.Item) {
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		if ic.Effect == "heal" {
			v.invDesc.AddText("")
			v.invDesc.AddText(fmt.Sprintf("Use to heal %d HP", ic.Value))
		} else if ic.Effect == "skill_chip" && item.HasComponent(component.SkillChip) {
			scc := item.GetComponent(component.SkillChip).(*component.SkillChipComponent)
			sd := skill.Get(scc.SkillId)
			v.invDesc.AddText("")
			if sd != nil {
				v.invDesc.AddText(fmt.Sprintf("Teaches: %s", sd.Name))
				if sd.Description != "" {
					v.invDesc.AddText("")
					for _, line := range mlge_text.Wrap(sd.Description, 38, 0) {
						v.invDesc.AddText(line)
					}
				}
				if skill.HasSkill(v.player, scc.SkillId) {
					v.invDesc.AddText("")
					v.invDesc.AddText("(already learned)")
				}
			}
		} else if ic.Slot != component.BagSlot {
			v.invDesc.AddText("")
			v.invDesc.AddText(fmt.Sprintf("Slot: %s", ic.Slot))
		}
	}
}

func (v *CharacterStatsView) onInvEquip() {
	idx := v.invList.SelectedIndex
	if idx < 0 || idx >= len(v.invItems) {
		return
	}
	item := v.invItems[idx]
	if item.HasComponent(component.Item) {
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		if ic.Effect == "heal" {
			entityhelpers.HealBodyParts(v.player, ic.Value)
			playerRemoveItem(v.player, item)
		} else if ic.Effect == "skill_chip" && item.HasComponent(component.SkillChip) {
			scc := item.GetComponent(component.SkillChip).(*component.SkillChipComponent)
			if skill.HasSkill(v.player, scc.SkillId) {
				sd := skill.Get(scc.SkillId)
				name := scc.SkillId
				if sd != nil {
					name = sd.Name
				}
				message.AddMessage("You already know " + name + ".")
			} else {
				skill.Apply(v.player, scc.SkillId)
				playerRemoveItem(v.player, item)
				sd := skill.Get(scc.SkillId)
				name := scc.SkillId
				if sd != nil {
					name = sd.Name
				}
				message.AddMessage("You slot the chip. You have learned " + name + "!")
			}
		} else if ic.Slot != component.BagSlot {
			playerEquipItem(v.player, item)
			skill.SyncEquippedSkills(v.player)
		}
	}
	v.refreshInventoryList()
}

func (v *CharacterStatsView) onInvDrop() {
	idx := v.invList.SelectedIndex
	if idx < 0 || idx >= len(v.invItems) {
		return
	}
	item := v.invItems[idx]
	pc := v.player.GetComponent("Position").(*component.PositionComponent)
	eventsystem.EventManager.SendEvent(eventsystem.DropItemEventData{
		X:    pc.GetX(),
		Y:    pc.GetY(),
		Z:    pc.GetZ(),
		Item: item,
	})
	playerRemoveItem(v.player, item)
	v.refreshInventoryList()
}

func (v *CharacterStatsView) onInvTake() {
	idx := v.invNearbyList.SelectedIndex
	if idx < 0 || idx >= len(v.invNearbyItems) || v.invNearbyEntity == nil {
		return
	}
	item := v.invNearbyItems[idx]
	slot := v.invNearbyItemSlots[idx]
	if slot != "" {
		// Unequip from the body-part slot first (moves it to the bag).
		if v.invNearbyEntity.HasComponent(component.BodyInventory) {
			v.invNearbyEntity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Unequip(slot)
		}
	}
	nearbyRemoveItem(v.invNearbyEntity, item)
	playerAddItem(v.player, item)
	v.refreshInventoryList()
	v.refreshNearbyList()
}

func (v *CharacterStatsView) onInvGive() {
	idx := v.invList.SelectedIndex
	if idx < 0 || idx >= len(v.invItems) || v.invNearbyEntity == nil {
		return
	}
	item := v.invItems[idx]
	playerRemoveItem(v.player, item)
	nearbyAddItem(v.invNearbyEntity, item)
	v.refreshInventoryList()
	v.refreshNearbyList()
}

// SetNearbyEntity sets the entity whose inventory is shown in the nearby panel.
// Pass nil to hide it.
func (v *CharacterStatsView) SetNearbyEntity(e *ecs.Entity) {
	v.invNearbyEntity = e
	v.refreshNearbyList()
}

func (v *CharacterStatsView) refreshNearbyList() {
	if v.invNearbyEntity == nil {
		v.invNearbyItems = nil
		v.invNearbyItemSlots = nil
		v.invNearbyList.SetItems([]string{})
		v.invNearbyLabel.Text = ""
		v.invNearbyList.SetVisible(false)
		v.invNearbyLabel.SetVisible(false)
		v.invTake.SetVisible(false)
		v.invGive.SetVisible(false)
		return
	}
	v.invNearbyList.SetVisible(true)
	v.invNearbyLabel.SetVisible(true)
	v.invTake.SetVisible(true)
	v.invGive.SetVisible(true)
	name := "Nearby"
	if v.invNearbyEntity.HasComponent(component.Description) {
		name = v.invNearbyEntity.GetComponent(component.Description).(*component.DescriptionComponent).Name
	}
	v.invNearbyLabel.Text = name

	v.invNearbyItems = nil
	v.invNearbyItemSlots = nil
	names := []string{}

	// When dead, also show equipped items so they can be looted.
	isDead := v.invNearbyEntity.HasComponent(component.Dead)
	if isDead && v.invNearbyEntity.HasComponent(component.BodyInventory) {
		inv := v.invNearbyEntity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for slot, item := range inv.Equipped {
			if item == nil {
				continue
			}
			v.invNearbyItems = append(v.invNearbyItems, item)
			v.invNearbyItemSlots = append(v.invNearbyItemSlots, slot)
			names = append(names, "[eq] "+itemName(item))
		}
	}

	for _, item := range nearbyBag(v.invNearbyEntity) {
		v.invNearbyItems = append(v.invNearbyItems, item)
		v.invNearbyItemSlots = append(v.invNearbyItemSlots, "")
		names = append(names, itemName(item))
	}

	v.invNearbyList.SetItems(names)
}

func (v *CharacterStatsView) refreshNearbyDesc(idx int) {
	v.invDesc.Clear()
	v.invImg.Image = nil
	if idx < 0 || idx >= len(v.invNearbyItems) {
		return
	}
	item := v.invNearbyItems[idx]
	v.invDesc.AddText(itemName(item))
	if desc := itemDesc(item); desc != "" {
		v.invDesc.AddText("")
		for _, line := range mlge_text.Wrap(desc, 28, 0) {
			v.invDesc.AddText(line)
		}
	}
	v.invImg.Image = itemSpriteImage(item)
}

// -----------------------------------------------------------------------
// Refresh helpers
// -----------------------------------------------------------------------

func (v *CharacterStatsView) refreshOverview() {
	v.overviewArea.Clear()
	if v.player == nil {
		return
	}

	// Name
	name := "Unknown"
	if v.player.HasComponent(component.Description) {
		dc := v.player.GetComponent(component.Description).(*component.DescriptionComponent)
		name = dc.Name
	}
	v.overviewArea.AddText("Name: " + name)
	v.overviewArea.AddText("")

	// Stats + active skill modifiers
	if v.player.HasComponent(component.Stats) {
		sc := v.player.GetComponent(component.Stats).(*component.StatsComponent)
		stats := []struct {
			name string
			val  int
			key  string
		}{
			{"PH", sc.PH, "ph"},
			{"AG", sc.AG, "ag"},
			{"MA", sc.MA, "ma"},
			{"CL", sc.CL, "cl"},
			{"LD", sc.LD, "ld"},
			{"CS", sc.CS, "cs"},
			{"HTCS", sc.HTCS, "htcs"},
		}
		v.overviewArea.AddText("── Stats ──────────────")
		for _, s := range stats {
			mods := v.collectStatMods(s.key)
			var line string
			switch s.key {
			case "cs":
				eff := 20 + sc.AG/2 + sc.CS
				if mods == "" {
					line = fmt.Sprintf("  %s: %d  [eff: %d%%]", s.name, s.val, eff)
				} else {
					line = fmt.Sprintf("  %s: %d  [eff: %d%%]  (%s)", s.name, s.val, eff, mods)
				}
			case "htcs":
				eff := 20 + sc.PH/2 + sc.HTCS
				if mods == "" {
					line = fmt.Sprintf("  %s: %d  [eff: %d%%]", s.name, s.val, eff)
				} else {
					line = fmt.Sprintf("  %s: %d  [eff: %d%%]  (%s)", s.name, s.val, eff, mods)
				}
			default:
				if mods == "" {
					line = fmt.Sprintf("  %s: %d", s.name, s.val)
				} else {
					line = fmt.Sprintf("  %s: %d  (%s)", s.name, s.val, mods)
				}
			}
			v.overviewArea.AddText(line)
		}
		v.overviewArea.AddText("")
	}

	// Class
	if v.player.HasComponent(component.Class) {
		cc := v.player.GetComponent(component.Class).(*component.ClassComponent)
		v.overviewArea.AddText("── Class ───────────────")
		for _, classID := range cc.Classes {
			def := class.Get(classID)
			if def == nil {
				continue
			}
			v.overviewArea.AddText("  " + def.Name)
			for _, line := range mlge_text.Wrap(def.Description, 45, 0) {
				v.overviewArea.AddText("    " + line)
			}
		}
		v.overviewArea.AddText("")
	}

	// Background
	if v.player.HasComponent(component.Background) {
		bc := v.player.GetComponent(component.Background).(*component.BackgroundComponent)
		def := background.Get(bc.BackgroundID)
		if def != nil {
			v.overviewArea.AddText("── Background ──────────")
			v.overviewArea.AddText("  " + def.Name)
			for _, line := range mlge_text.Wrap(def.Description, 45, 0) {
				v.overviewArea.AddText("    " + line)
			}
		}
	}
}

// collectStatMods returns a short "+2 Brawler, -1 …" string for a stat key.
func (v *CharacterStatsView) collectStatMods(statKey string) string {
	if !v.player.HasComponent(component.Skill) {
		return ""
	}
	sc := v.player.GetComponent(component.Skill).(*component.SkillComponent)
	var parts []string
	for _, sID := range sc.Skills {
		def := skill.Get(sID)
		if def == nil {
			continue
		}
		for _, m := range def.StatMods {
			if strings.EqualFold(m.Stat, statKey) {
				sign := "+"
				if m.Delta < 0 {
					sign = ""
				}
				parts = append(parts, fmt.Sprintf("%s%d %s", sign, m.Delta, def.Name))
			}
		}
	}
	return strings.Join(parts, ", ")
}

func (v *CharacterStatsView) onEquipUnequip() {
	idx := v.equipList.SelectedIndex
	if idx < 0 || idx >= len(v.equipIDs) {
		return
	}
	slot := v.equipIDs[idx]
	if v.player.HasComponent(component.BodyInventory) {
		v.player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Unequip(slot)
		skill.SyncEquippedSkills(v.player)
	} else if v.player.HasComponent(component.Inventory) {
		inv := v.player.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.Unequip(component.ItemSlot(slot))
	}
	v.refreshEquipmentList()
}

func (v *CharacterStatsView) refreshEquipmentList() {
	equipped := playerEquipped(v.player)
	v.equipIDs = v.equipIDs[:0]
	items := []string{}
	for slot, item := range equipped {
		if item == nil {
			continue
		}
		v.equipIDs = append(v.equipIDs, slot)
		items = append(items, fmt.Sprintf("%s: %s", slot, itemName(item)))
	}
	v.equipList.SetItems(items)
	v.equipDesc.Clear()
	v.equipDesc.AddText("Select an item to see details.")
}

func (v *CharacterStatsView) refreshEquipDesc(idx int) {
	v.equipDesc.Clear()
	v.equipImg.Image = nil
	if idx < 0 || idx >= len(v.equipIDs) {
		return
	}
	equipped := playerEquipped(v.player)
	item := equipped[v.equipIDs[idx]]
	if item == nil {
		return
	}

	v.equipDesc.AddText(itemName(item))
	if desc := itemDesc(item); desc != "" {
		v.equipDesc.AddText("")
		for _, line := range mlge_text.Wrap(desc, 38, 0) {
			v.equipDesc.AddText(line)
		}
	}

	// Set item sprite image
	v.equipImg.Image = itemSpriteImage(item)

	// Skills granted by this item
	if item.HasComponent(component.ItemSkills) {
		isc := item.GetComponent(component.ItemSkills).(*component.ItemSkillsComponent)
		if len(isc.Skills) > 0 {
			v.equipDesc.AddText("")
			v.equipDesc.AddText("Grants skills:")
			for _, sID := range isc.Skills {
				sname := sID
				if sd := skill.Get(sID); sd != nil {
					sname = sd.Name
				}
				v.equipDesc.AddText("  • " + sname)
			}
		}
	}
}

func (v *CharacterStatsView) refreshSkillList() {
	if !v.player.HasComponent(component.Skill) {
		v.skillList.SetItems([]string{})
		v.skillIDs = nil
		return
	}
	sc := v.player.GetComponent(component.Skill).(*component.SkillComponent)
	v.skillIDs = make([]string, len(sc.Skills))
	names := make([]string, len(sc.Skills))
	for i, sID := range sc.Skills {
		v.skillIDs[i] = sID
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		names[i] = name
	}
	v.skillList.SetItems(names)
	v.skillDesc.Clear()
	v.skillDesc.AddText("Select a skill to see its description.")
}

func (v *CharacterStatsView) refreshSkillDesc(idx int) {
	v.skillDesc.Clear()
	if idx < 0 || idx >= len(v.skillIDs) {
		return
	}
	sd := skill.Get(v.skillIDs[idx])
	if sd == nil {
		return
	}
	v.skillDesc.AddText(sd.Name)
	v.skillDesc.AddText("")
	for _, line := range mlge_text.Wrap(sd.Description, 38, 0) {
		v.skillDesc.AddText(line)
	}

	// Source (item skill vs class skill vs background skill)
	if v.player.HasComponent(component.Skill) {
		sc := v.player.GetComponent(component.Skill).(*component.SkillComponent)
		for _, is := range sc.ItemSkills {
			if is == v.skillIDs[idx] {
				v.skillDesc.AddText("")
				v.skillDesc.AddText("(from equipped item)")
				break
			}
		}
	}

	// Stat mods
	if len(sd.StatMods) > 0 {
		v.skillDesc.AddText("")
		v.skillDesc.AddText("Stat modifiers:")
		for _, m := range sd.StatMods {
			sign := "+"
			if m.Delta < 0 {
				sign = ""
			}
			v.skillDesc.AddText(fmt.Sprintf("  %s%d %s", sign, m.Delta, m.Stat))
		}
	}
}

// -----------------------------------------------------------------------
// Update / Draw
// -----------------------------------------------------------------------

func (v *CharacterStatsView) Update() {
	if !v.Visible {
		return
	}
	v.modal.Update()
}

func (v *CharacterStatsView) Draw(screen *ebiten.Image) {
	if !v.Visible {
		return
	}
	v.modal.Draw(screen)
}

// -----------------------------------------------------------------------
// Helper functions
// -----------------------------------------------------------------------

// itemSpriteImage returns the sprite image for an item, or nil if unavailable.
func itemSpriteImage(item *ecs.Entity) *ebiten.Image {
	if !item.HasComponent(component.Appearance) {
		return nil
	}
	ac := item.GetComponent(component.Appearance).(*component.AppearanceComponent)
	cfg := config.Global()
	return resource.GetSubImage(ac.Resource, ac.SpriteX, ac.SpriteY, cfg.TileSizeW, cfg.TileSizeH)
}
