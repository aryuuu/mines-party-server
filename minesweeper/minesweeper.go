package minesweeper

import (
	"errors"
	"strconv"

	"github.com/aryuuu/mines-party-server/utils"
)

type Field struct {
	row        int
	col        int
	minesCount int
	cells      [][]*Cell
}

func New(row, col, mines int) *Field {
	field := &Field{
		row:        row,
		col:        col,
		minesCount: mines,
		cells:      generateCells(row, col),
	}
	field.generateMines()
	field.setAdjacentMinesCount()

	return field
}

func (f Field) String() string {
	result := ""
	for _, row := range f.cells {
		for _, cell := range row {
			result += cell.GetValueBare()
		}
		result += "\n"
	}
	return result
}

// OpenCell opens the cell at the given position.
// TODO: finish this function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) OpenCell(pos Coordinate) (*Cell, error) {
	cell := f.cells[pos.x][pos.y]

	cell.Open()
	// TODO: do something if the cell is a mine?
	return cell, nil
}

// FlagCell flags the cell at the given position.
// TODO: maybe consider doing the flag x mines count check?
func (f *Field) FlagCell(pos Coordinate) (*Cell, error) {
	cell := f.cells[pos.x][pos.y]

	if cell.isOpen {
		return nil, errors.New("cannot flag an open cell")
	}

	cell.Flag()

	return cell, nil
}

// QuickOpenCell opens the cell at the given position and all the adjacent cells if the number of adjacent flagged cells is equal to the number of adjacent mines.
// TODO: finish function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) QuickOpenCell(pos Coordinate) error {
	return nil
}

// generateCells generates row * col cells.
func generateCells(row, col int) [][]*Cell {
	cells := make([][]*Cell, row)
	for i := range cells {
		cells[i] = make([]*Cell, col)
		for j := range cells[i] {
			cells[i][j] = &Cell{}
		}
	}
	return cells
}

// generateMines generates mines randomly.
func (f *Field) generateMines() error {
	minesLocations, err := generateMinesLocations(f.row, f.col, f.minesCount)
	if err != nil {
		return err
	}

	for _, loc := range minesLocations {
		f.cells[loc.x][loc.y].isMine = true
	}

	return nil
}

func (f *Field) setAdjacentMinesCount() {
	for i, row := range f.cells {
		for j, cell := range row {
			if cell.isMine {
				continue
			}

			cell.adjacentMines = uint8(f.getAdjacentMinesCount(i, j))
		}
	}
}

func (f *Field) getAdjacentMinesCount(row, col int) int {
	result := 0

	// A B C
	// D X E
	// F G H
	if row > 0 {
		// A
		if col > 0 && f.cells[row-1][col-1].isMine {
			result++
		}
		// B
		if f.cells[row-1][col].isMine {
			result++
		}
		// C
		if col < f.col-1 && f.cells[row-1][col+1].isMine {
			result++
		}
	}

	// D
	if col > 0 && f.cells[row][col-1].isMine {
		result++
	}
	// E
	if col < f.col-1 && f.cells[row][col+1].isMine {
		result++
	}

	if row < f.row-1 {
		// F
		if col > 0 && f.cells[row+1][col-1].isMine {
			result++
		}
		// G
		if f.cells[row+1][col].isMine {
			result++
		}
		// H
		if col < f.col-1 && f.cells[row+1][col+1].isMine {
			result++
		}
	}

	return result
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

func (c Cell) GetValueBare() string {
	if c.isMine {
		return "X"
	}

	result := " "
	if c.adjacentMines > 0 {
		result = strconv.Itoa(int(c.adjacentMines))
	}

	return result
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

func (c *Cell) Open() {
	c.isOpen = true
}

// TODO: maybe consider doing the flag x mines count check?
func (c *Cell) Flag() {
	c.isFlagged = true
}

type Coordinate struct {
	x int
	y int
}
