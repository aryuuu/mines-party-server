package main

import (
	"fmt"
	"math"

	"github.com/aryuuu/mines-party-server/minesweeper"
)

func main() {
	x := 1534236469
	fmt.Println(math.MaxInt32)
	if x > math.MaxInt32 || x < math.MinInt32 {
		fmt.Println(0)
	}
	field := minesweeper.NewField(8, 8, 10)

	field.OpenCell(0, 0)
	fmt.Println(field.String())
}
