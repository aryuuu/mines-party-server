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

	if cell.isFlagged {
		return cell, errors.New("cannot open a flagged cell")
	}

	cell.Open()
	if !f.isStarted {
		genesisCoordinate := Location{
			row: row,
			col: col,
		}
		f.isStarted = true
		f.generateMines(genesisCoordinate)
		f.setAdjacentMinesCount()
	}

	// TODO: do quick open cell to at least the adjacent cells
	adjacentFlagCount := f.getAdjacentFlagCount(row, col)
	if int(cell.adjacentMines) == adjacentFlagCount {
		// TODO: maybe return all cells?
		f.QuickOpenCell(row, col)
	}

	// TODO: do something if the cell is a mine?
	return cell, nil
}

func (f *Field) getAdjacentFlagCount(row, col int) int {
	result := 0

	for i := row - 1; i <= row+1; i++ {
		for j := col - 1; j <= col+1; j++ {
			if i == row && j == col {
				continue
			}

			if i < 0 || i >= f.row || j < 0 || j >= f.col {
				continue
			}

			if f.cells[i][j].isFlagged {
				result++
			}
		}
	}

	return result
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
func (f *Field) QuickOpenCell(row, col int) error {
	// flood fill rule:
	// if current cell is not a mine and has no adjacent mines, open all adjacent cells
	// if current cell is not a mine and has adjacent mines, open only the current cell
	// if current cell is a mine, do nothing
	// if current cell is flagged, do nothing
	// if current cell is open, open all adjacent cells that are not flagged

	locationsToOpen := []Location{}
	for i := row - 1; i <= row+1; i++ {
		for j := col - 1; j <= col+1; j++ {
			if i == row && j == col {
				continue
			}

			locationsToOpen = append(locationsToOpen, Location{
				row: i,
				col: j,
			})
		}
	}

	for len(locationsToOpen) > 0 {
		loc := locationsToOpen[0]
		locationsToOpen = locationsToOpen[1:]

		if loc.row < 0 || loc.row >= f.row || loc.col < 0 || loc.col >= f.col {
			continue
		}

		cell := f.cells[loc.row][loc.col]

		if cell.isMine || cell.isFlagged || cell.isOpen {
			continue
		}

		cell.Open()

		// TODO: also open when adjacentFlagCount == adjacentMinesCount
		if cell.adjacentMines == 0 {
			for i := loc.row - 1; i <= loc.row+1; i++ {
				for j := loc.col - 1; j <= loc.col+1; j++ {
					if i == loc.row && j == loc.col {
						continue
					}

					locationsToOpen = append(locationsToOpen, Location{
						row: i,
						col: j,
					})
				}
			}
		}
	}

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

	for i := row - 1; i <= row+1; i++ {
		for j := col - 1; j <= col+1; j++ {
			if i == row && j == col {
				continue
			}

			if i < 0 || i >= f.row || j < 0 || j >= f.col {
				continue
			}

			if f.cells[i][j].isMine {
				result++
			}
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

	neutralSeries := map[int]bool{}
	for i := genesisCoordinate.row - 1; i <= genesisCoordinate.row+1; i++ {
		for j := genesisCoordinate.col - 1; j <= genesisCoordinate.col+1; j++ {
			neutralSeries[i*f.col+j] = true
		}
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
