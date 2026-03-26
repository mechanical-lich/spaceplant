package game

import (
	"fmt"
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
)

type InventoryView struct {
	Visible bool
	X       float64
	Y       float64
	Width   float64
	Height  float64

	inventoryButtons []Button
	player           *ecs.Entity
	tab              int // 0 - Inventory, 1 - Equipment
}

func NewInventoryView(player *ecs.Entity) *InventoryView {
	view := &InventoryView{}
	view.X = 375.0
	view.Y = 225.0
	view.Width = 500.0
	view.Height = 500.0
	view.player = player

	return view
}

// --- inventory helpers that work with either BodyInventoryComponent or InventoryComponent ---

func playerBag(player *ecs.Entity) []*ecs.Entity {
	if player.HasComponent(component.BodyInventory) {
		return player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Bag
	}
	if player.HasComponent(component.Inventory) {
		return player.GetComponent(component.Inventory).(*component.InventoryComponent).Bag
	}
	return nil
}

func playerRemoveItem(player *ecs.Entity, item *ecs.Entity) {
	if player.HasComponent(component.BodyInventory) {
		player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).RemoveItem(item)
		return
	}
	if player.HasComponent(component.Inventory) {
		player.GetComponent(component.Inventory).(*component.InventoryComponent).RemoveItem(item)
	}
}

func healBodyParts(entity *ecs.Entity, amount int) {
	if !entity.HasComponent(component.Body) {
		return
	}
	bc := entity.GetComponent(component.Body).(*component.BodyComponent)
	var damaged []string
	for name, part := range bc.Parts {
		if !part.Amputated && part.HP < part.MaxHP {
			damaged = append(damaged, name)
		}
	}
	if len(damaged) == 0 {
		return
	}
	perPart := amount / len(damaged)
	remainder := amount % len(damaged)
	for i, name := range damaged {
		part := bc.Parts[name]
		heal := perPart
		if i < remainder {
			heal++
		}
		part.HP += heal
		if part.HP > part.MaxHP {
			part.HP = part.MaxHP
		}
		if part.HP > 0 && part.Broken {
			part.Broken = false
		}
		bc.Parts[name] = part
	}
}

func playerEquipItem(player *ecs.Entity, item *ecs.Entity) {
	if player.HasComponent(component.BodyInventory) && player.HasComponent(component.Body) {
		inv := player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		bc := player.GetComponent(component.Body).(*component.BodyComponent)
		inv.AutoEquip(item, bc)
		return
	}
	if player.HasComponent(component.Inventory) {
		player.GetComponent(component.Inventory).(*component.InventoryComponent).Equip(item)
	}
}

// playerEquipped returns a slot→item map regardless of inventory type.
func playerEquipped(player *ecs.Entity) map[string]*ecs.Entity {
	if player.HasComponent(component.BodyInventory) {
		return player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Equipped
	}
	if player.HasComponent(component.Inventory) {
		inv := player.GetComponent(component.Inventory).(*component.InventoryComponent)
		m := map[string]*ecs.Entity{}
		if inv.Head != nil {
			m["Head"] = inv.Head
		}
		if inv.Torso != nil {
			m["Torso"] = inv.Torso
		}
		if inv.Legs != nil {
			m["Legs"] = inv.Legs
		}
		if inv.Feet != nil {
			m["Feet"] = inv.Feet
		}
		if inv.RightHand != nil {
			m["R Hand"] = inv.RightHand
		}
		if inv.LeftHand != nil {
			m["L Hand"] = inv.LeftHand
		}
		return m
	}
	return nil
}

func (view *InventoryView) Update() {
	if view.player != nil && view.Visible {
		cX, cY := ebiten.CursorPosition()

		if view.tab == 0 {
			inventoryX := view.X + 4.0
			inventoryY := view.Y + 48.0
			itemHeight := 16
			bag := playerBag(view.player)
			view.inventoryButtons = []Button{}
			for i, v := range bag {
				d := v.GetComponent("Description").(*component.DescriptionComponent)
				b := Button{inventoryX, inventoryY + float64(15+(i*itemHeight)), 100, float64(itemHeight), d.Name}
				view.inventoryButtons = append(view.inventoryButtons, b)

				b2 := Button{inventoryX + 170, inventoryY + float64(15+(i*itemHeight)), 40, float64(itemHeight), "drop"}
				view.inventoryButtons = append(view.inventoryButtons, b2)

				if b.Within(cX, cY) {
					if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
						item := v.GetComponent("Item").(*component.ItemComponent)
						if item.Effect == "heal" {
							healBodyParts(view.player, item.Value)
							playerRemoveItem(view.player, v)
						} else if item.Slot != component.BagSlot {
							playerEquipItem(view.player, v)
						}
					}
				}

				if b2.Within(cX, cY) {
					if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
						pc := view.player.GetComponent("Position").(*component.PositionComponent)
						data := eventsystem.DropItemEventData{
							X:    pc.GetX(),
							Y:    pc.GetY(),
							Z:    pc.GetZ(),
							Item: v,
						}
						eventsystem.EventManager.SendEvent(data)
						playerRemoveItem(view.player, v)
					}
				}
			}
		}
	}
}

func (view *InventoryView) Draw(screen *ebiten.Image) {
	if view.player != nil && view.Visible {
		cX, cY := ebiten.CursorPosition()
		if view.tab == 0 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(view.X, view.Y)
			screen.DrawImage(resource.Textures["inventory"], op)

			for _, v := range view.inventoryButtons {
				m := fmt.Sprintf("* %s", v.Text)
				if v.Text == "drop" {
					m = "drop"
				}
				if v.Within(cX, cY) {
					ebitenutil.DrawRect(screen, v.X, v.Y, v.Width, v.Height, color.RGBA{0, 50, 50, 200})
					mlge_text.Draw(screen, m, 16, int(v.X), int(v.Y), color.Black)
				} else {
					mlge_text.Draw(screen, m, 16, int(v.X), int(v.Y), color.White)
				}
			}

			ebitenutil.DrawRect(screen, view.X+215, view.Y+50, 2, view.Height-100, color.White)

			stats := view.player.GetComponent("Stats").(*component.StatsComponent)
			statX := int(view.X) + 220
			statY := int(view.Y) + 100
			DrawStat(screen, statX, statY, "AC", stats.AC)
			DrawStat(screen, statX, statY+25, "STR", stats.Str)
			DrawStat(screen, statX, statY+50, "DEX", stats.Dex)
			DrawStat(screen, statX, statY+75, "INT", stats.Int)
			DrawStat(screen, statX, statY+100, "WIS", stats.Wis)

			equipped := playerEquipped(view.player)
			keys := make([]string, 0, len(equipped))
			for k := range equipped {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for i, slot := range keys {
				DrawEquipment(screen, statX, statY+150+(i*25), slot, equipped[slot])
			}
		}
	}
}

func DrawStat(screen *ebiten.Image, x, y int, stat string, value int) {
	m := fmt.Sprintf("%s : %d", stat, value)
	mlge_text.Draw(screen, m, 24, x, y, color.White)
}

func DrawEquipment(screen *ebiten.Image, x, y int, slot string, item *ecs.Entity) {
	m := fmt.Sprintf("%s : %s", slot, "-")

	if item != nil {
		dc := item.GetComponent("Description").(*component.DescriptionComponent)
		m = fmt.Sprintf("%s : %s", slot, dc.Name)

		if item.HasComponent("Weapon") {
			wc := item.GetComponent("Weapon").(*component.WeaponComponent)
			m = fmt.Sprintf("%s : %s (%s + %d)", slot, dc.Name, wc.AttackDice, wc.AttackBonus)
		}

		if item.HasComponent("Armor") {
			wc := item.GetComponent("Armor").(*component.ArmorComponent)
			m = fmt.Sprintf("%s : %s (%d)", slot, dc.Name, wc.DefenseBonus)
		}
	}
	mlge_text.Draw(screen, m, 16, x, y, color.White)
}

type Button struct {
	X, Y, Width, Height float64
	Text                string
}

func (b Button) Within(x, y int) bool {
	if x > int(b.X) && x < int(b.X+b.Width) && y > int(b.Y) && y < int(b.Y+b.Height) {
		return true
	}
	return false
}
