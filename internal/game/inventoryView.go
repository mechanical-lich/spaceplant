package game

import (
	"fmt"
	"image/color"

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

func (view *InventoryView) Update() {
	if view.player != nil && view.Visible {
		cX, cY := ebiten.CursorPosition()

		if view.tab == 0 {
			inventoryX := view.X + 4.0
			inventoryY := view.Y + 48.0
			itemHeight := 16
			// Populate buttons for inventory/check for clicks
			inventory := view.player.GetComponent("Inventory").(*component.InventoryComponent)
			view.inventoryButtons = []Button{}
			for i, v := range inventory.Bag {
				d := v.GetComponent("Description").(*component.DescriptionComponent)
				b := Button{inventoryX, inventoryY + float64(15+(i*itemHeight)), 100, float64(itemHeight), d.Name}
				view.inventoryButtons = append(view.inventoryButtons, b)

				b2 := Button{inventoryX + 170, inventoryY + float64(15+(i*itemHeight)), 40, float64(itemHeight), "drop"}
				view.inventoryButtons = append(view.inventoryButtons, b2)

				if b.Within(cX, cY) {
					// TODO Temp use code
					if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
						item := v.GetComponent("Item").(*component.ItemComponent)
						if item.Effect == "heal" {
							view.player.GetComponent("Health").(*component.HealthComponent).Health += item.Value
							inventory.RemoveItem(v)
						} else if item.Slot != component.BagSlot {
							inventory.Equip(v)
						}
					}
				}

				if b2.Within(cX, cY) {
					// TODO Temp drop code
					if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
						pc := view.player.GetComponent("Position").(*component.PositionComponent)
						data := eventsystem.DropItemEventData{
							X:    pc.GetX(),
							Y:    pc.GetY(),
							Z:    pc.GetZ(),
							Item: v,
						}
						eventsystem.EventManager.SendEvent(data)
						inventory.RemoveItem(v)
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
			// Dialog background
			op := &ebiten.DrawImageOptions{}
			//Position
			op.GeoM.Translate(view.X, view.Y)

			screen.DrawImage(resource.Textures["inventory"], op)
			// Inventory
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

			//Divider
			ebitenutil.DrawRect(screen, view.X+215, view.Y+50, 2, view.Height-100, color.White)

			// Stats
			stats := view.player.GetComponent("Stats").(*component.StatsComponent)
			inventoryComponent := view.player.GetComponent("Inventory").(*component.InventoryComponent)

			statX := int(view.X) + 220
			statY := int(view.Y) + 100
			DrawStat(screen, statX, statY, "AC", stats.AC)
			DrawStat(screen, statX, statY+25, "STR", stats.Str)
			DrawStat(screen, statX, statY+50, "DEX", stats.Dex)
			DrawStat(screen, statX, statY+75, "INT", stats.Int)
			DrawStat(screen, statX, statY+100, "WIS", stats.Wis)

			// Equipment
			DrawEquipment(screen, statX, statY+150, "Head", inventoryComponent.Head)
			DrawEquipment(screen, statX, statY+175, "L Hand", inventoryComponent.LeftHand)
			DrawEquipment(screen, statX, statY+200, "R Hand", inventoryComponent.RightHand)
			DrawEquipment(screen, statX, statY+225, "Torso", inventoryComponent.Torso)
			DrawEquipment(screen, statX, statY+250, "Legs", inventoryComponent.Legs)
			DrawEquipment(screen, statX, statY+275, "Feet", inventoryComponent.Feet)

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
