package world

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
)

// Tile type indices — populated after LoadTileDefinitions.
var (
	TypeOpen                  int
	TypeWall                  int
	TypeFloor                 int
	TypeDoor                  int
	TypeStairsUp              int
	TypeStairsDown            int
	TypeMaintenanceTunnelWall int
	TypeMaintenanceTunnelFloor int
	TypeMaintenanceTunnelDoor int
)

func LoadTileDefinitions(path string) error {
	if err := rlworld.LoadTileDefinitions(path); err != nil {
		return err
	}
	TypeOpen = rlworld.TileNameToIndex["open"]
	TypeWall = rlworld.TileNameToIndex["wall"]
	TypeFloor = rlworld.TileNameToIndex["floor"]
	TypeDoor = rlworld.TileNameToIndex["door"]
	TypeStairsUp = rlworld.TileNameToIndex["stairs_up"]
	TypeStairsDown = rlworld.TileNameToIndex["stairs_down"]
	TypeMaintenanceTunnelWall = rlworld.TileNameToIndex["maintenance_tunnel_wall"]
	TypeMaintenanceTunnelFloor = rlworld.TileNameToIndex["maintenance_tunnel_floor"]
	TypeMaintenanceTunnelDoor = rlworld.TileNameToIndex["maintenance_tunnel_door"]
	return nil
}
