package level

import "image/color"

type Theme struct {
	Wall                   []int
	WallTop                []int
	Floor                  []int
	Door                   []int
	MaintenanceTunnelFloor []int
	MaintenanceTunnelWall  []int
	MaintenanceTunnelTop   []int
	MaintenanceTunnelDoor  []int
	Open                   []int

	OpenForgroundColor       color.Color
	OpenBackgroundColor      color.Color
	BackgroundColor          color.Color
	ForgroundColor           color.Color
	SecondaryBackgroundColor color.Color
	SecondaryForgroundColor  color.Color
}

func NewDefaultTheme() Theme {
	return Theme{
		Wall:                   []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		WallTop:                []int{10},
		Floor:                  []int{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 16}, // Crappy weighted randomness
		Door:                   []int{11, 12},
		MaintenanceTunnelFloor: []int{16},
		MaintenanceTunnelWall:  []int{17},
		MaintenanceTunnelTop:   []int{15},
		MaintenanceTunnelDoor:  []int{18},

		Open:                     []int{19},
		OpenForgroundColor:       color.RGBA{R: 255, G: 255, B: 255, A: 100},
		OpenBackgroundColor:      color.RGBA{R: 100, G: 100, B: 150, A: 255},
		BackgroundColor:          color.RGBA{R: 0, G: 0, B: 0, A: 255},
		ForgroundColor:           color.RGBA{R: 125, G: 125, B: 125, A: 255},
		SecondaryBackgroundColor: color.RGBA{R: 25, G: 25, B: 75, A: 255},
		SecondaryForgroundColor:  color.RGBA{R: 125, G: 125, B: 125, A: 255},
	}
}
