package minesweeper

import (
	"errors"

	"github.com/aryuuu/mines-party-server/utils"
)

type Field struct {
	cells [][]Cell
}

func New(row, col, mines int) *Field {
	field := &Field{
		cells: generateCells(row, col),
	}
	field.generateMines(row, col, mines)

	return field
}

// OpenCell opens the cell at the given position.
// TODO: finish this function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) OpenCell(pos Coordinate) (Cell, error) {
	cell := f.cells[pos.x][pos.y]

	// TODO: check if this actually updates the value inside the cells array
	cell.isOpen = true
	return cell, nil
}

// FlagCell flags the cell at the given position.
// TODO: finish function
func (f *Field) FlagCell(pos Coordinate) error {
	return nil
}

// QuickOpenCell opens the cell at the given position and all the adjacent cells if the number of adjacent flagged cells is equal to the number of adjacent mines.
// TODO: finish function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) QuickOpenCell(pos Coordinate) error {
	return nil
}

// generateCells generates row * col cells.
func generateCells(row, col int) [][]Cell {
	cells := make([][]Cell, row)
	for i := range cells {
		cells[i] = make([]Cell, col)
	}
	return cells
}

// generateMines generates mines randomly.
func (f *Field) generateMines(row, col, mineCount int) error {
	minesLocations, err := generateMinesLocations(row, col, mineCount)
	if err != nil {
		return err
	}

	for _, loc := range minesLocations {
		f.cells[loc.x][loc.y].isMine = true
	}

	return nil
}

// generateMinesLocations generates mines locations randomly.
func generateMinesLocations(row, col, mines int) ([]Coordinate, error) {
	cellCount := row * col
	if mines > cellCount {
		return nil, errors.New("too many mines")
	}

	minesLocations := make([]Coordinate, mines)
	for i := 0; i < mines; i++ {
		randomLoc := utils.GenerateRandomInt(0, cellCount)
		minesLocations = append(minesLocations, Coordinate{
			x: randomLoc / col,
			y: randomLoc % col,
		})
	}

	return minesLocations, nil
}

type Cell struct {
	isMine        bool
	isOpen        bool
	isFlagged     bool
	adjacentMines uint8
}

func (c Cell) GetValue() string {
	if c.isFlagged {
		return "F"
	}
	if c.isOpen {
		if c.isMine {
			return "X"
		}
		return string(c.adjacentMines)
	}
	return "0"
}

type Coordinate struct {
	x int
	y int
}
