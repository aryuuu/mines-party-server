package main

import (
	"fmt"

	"github.com/aryuuu/mines-party-server/minesweeper"
)

func main() {
	field := minesweeper.New(8, 8, 10)

	fmt.Println(field.String())
}
