package utility

import (
	"math/rand"
)

func GetRandom(low int, high int) int {
	if high < low {
		return low
	}
	if low == high {
		return low
	}
	return (rand.Intn((high - low))) + low
}
