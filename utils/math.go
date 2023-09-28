package utils

import (
	"math/rand"
	"time"
)

// GenerateRandomInt generates a random integer between min and max: [min, max)
func GenerateRandomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}
