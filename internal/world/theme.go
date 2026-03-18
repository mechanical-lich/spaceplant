package world

import "image/color"

type Theme struct {
	OpenForgroundColor       color.Color
	OpenBackgroundColor      color.Color
	BackgroundColor          color.Color
	ForgroundColor           color.Color
	SecondaryBackgroundColor color.Color
	SecondaryForgroundColor  color.Color
}

func NewDefaultTheme() Theme {
	return Theme{
		OpenForgroundColor:       color.RGBA{R: 0, G: 0, B: 0, A: 100},
		OpenBackgroundColor:      color.RGBA{R: 0, G: 0, B: 0, A: 255},
		BackgroundColor:          color.RGBA{R: 0, G: 0, B: 0, A: 255},
		ForgroundColor:           color.RGBA{R: 125, G: 125, B: 125, A: 255},
		SecondaryBackgroundColor: color.RGBA{R: 25, G: 25, B: 75, A: 255},
		SecondaryForgroundColor:  color.RGBA{R: 125, G: 125, B: 125, A: 255},
	}
}

// TileForgroundColor returns the foreground color for a tile based on its type.
func (t Theme) TileForgroundColor(tileType int) color.Color {
	switch tileType {
	case TypeOpen:
		return t.OpenForgroundColor
	case TypeWall:
		return t.ForgroundColor
	case TypeFloor:
		return t.ForgroundColor
	case TypeDoor, TypeMaintenanceTunnelDoor, TypeStairsUp, TypeStairsDown,
		TypeMaintenanceTunnelWall, TypeMaintenanceTunnelFloor:
		return t.SecondaryForgroundColor
	default:
		return t.ForgroundColor
	}
}

// TileBackgroundColor returns the background color for a tile based on its type.
func (t Theme) TileBackgroundColor(tileType int) color.Color {
	switch tileType {
	case TypeOpen:
		return t.OpenBackgroundColor
	case TypeWall, TypeFloor:
		return t.BackgroundColor
	case TypeDoor, TypeMaintenanceTunnelDoor, TypeStairsUp, TypeStairsDown,
		TypeMaintenanceTunnelWall, TypeMaintenanceTunnelFloor:
		return t.SecondaryBackgroundColor
	default:
		return t.BackgroundColor
	}
}
