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
	isStarted  bool
	cells      [][]*Cell
}

func New(row, col, mines int) *Field {
	field := &Field{
		row:        row,
		col:        col,
		minesCount: mines,
		isStarted:  false,
		cells:      generateCells(row, col),
	}
	// TODO: do this initialization after the first cell is opened

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

func (f Field) GetCells() [][]*Cell {
	return f.cells
}

func (f Field) GetRow() int {
	return f.row
}

func (f Field) GetCol() int {
	return f.col
}

// OpenCell opens the cell at the given position.
func (f *Field) OpenCell(row, col int) (*Cell, error) {
	cell := f.cells[row][col]

	cell.Open()
	if !f.isStarted {
		genesisCoordinate := Location{
			row: row,
			col: col,
		}
		// TODO: do quick open cell to at least the adjacent cells
		f.isStarted = true
		f.generateMines(genesisCoordinate)
		f.setAdjacentMinesCount()
	}

	// TODO: do something if the cell is a mine?
	return cell, nil
}

// FlagCell flags the cell at the given position.
// TODO: maybe consider doing the flag x mines count check?
func (f *Field) FlagCell(row, col int) (*Cell, error) {
	cell := f.cells[row][col]

	if cell.isOpen {
		return nil, errors.New("cannot flag an open cell")
	}

	cell.Flag()

	return cell, nil
}

// QuickOpenCell opens the cell at the given position and all the adjacent cells if the number of adjacent flagged cells is equal to the number of adjacent mines.
// TODO: finish function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) QuickOpenCell(pos Location) error {
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
func (f *Field) generateMines(genesisCoordinate Location) error {
	minesLocations, err := f.generateMinesLocations(genesisCoordinate, f.minesCount)
	if err != nil {
		return err
	}

	for _, loc := range minesLocations {
		f.cells[loc.row][loc.col].isMine = true
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
func (f Field) generateMinesLocations(genesisCoordinate Location, mines int) ([]Location, error) {
	cellCount := f.row * f.col
	if mines > cellCount {
		return nil, errors.New("too many mines")
	}

	// TODO: make sure the genesis coordinate is not a mine
	neutralSeries := map[int]bool{
		genesisCoordinate.row*f.col + genesisCoordinate.col:         true,
		genesisCoordinate.row*f.col + genesisCoordinate.col - 1:     true,
		genesisCoordinate.row*f.col + genesisCoordinate.col + 1:     true,
		(genesisCoordinate.row-1)*f.col + genesisCoordinate.col:     true,
		(genesisCoordinate.row-1)*f.col + genesisCoordinate.col - 1: true,
		(genesisCoordinate.row-1)*f.col + genesisCoordinate.col + 1: true,
		(genesisCoordinate.row+1)*f.col + genesisCoordinate.col:     true,
		(genesisCoordinate.row+1)*f.col + genesisCoordinate.col - 1: true,
		(genesisCoordinate.row+1)*f.col + genesisCoordinate.col + 1: true,
	}

	minesLocations := make([]Location, mines)
	for i := 0; i < mines; i++ {
		randomLoc := 0
		for {
			randomLoc = utils.GenerateRandomInt(0, cellCount)
			if !neutralSeries[randomLoc] {
				break
			}
		}
		minesLocations[i] = Location{
			row: randomLoc / f.col,
			col: randomLoc % f.col,
		}
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
	if c.isOpen {
		if c.isMine {
			return "X"
		}
		return strconv.Itoa(int(c.adjacentMines))
	}

	if c.isFlagged {
		return "F"
	}

	return " "
}

func (c *Cell) IsOpen() bool {
	return c.isOpen
}

func (c *Cell) Open() {
	c.isOpen = true
}

// TODO: maybe consider doing the flag x mines count check?
func (c *Cell) Flag() {
	c.isFlagged = true
}

type Location struct {
	row int
	col int
}
