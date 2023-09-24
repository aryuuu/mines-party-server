package utils

import "math/rand"

// GenerateRandomInt generates a random integer between min and max: [min, max)
func GenerateRandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}
