package main

import (
	"fmt"

	"github.com/aryuuu/mines-party-server/minesweeper"
)

func main() {
	field := minesweeper.NewField(8, 8, 10)

	field.OpenCell(0, 0)
	fmt.Println(field.String())
}
