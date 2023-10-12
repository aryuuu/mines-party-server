package configs

import (
	"os"
	"strconv"
)

type constant struct {
	Capacity int
}

func initConstant() *constant {
	capacity, _ := strconv.Atoi(os.Getenv("CAPACITY"))

	result := &constant{
		Capacity: capacity,
	}

	return result
}
