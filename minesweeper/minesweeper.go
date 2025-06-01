package minesweeper

import (
	"strconv"

	"github.com/aryuuu/mines-party-server/utils"
)

const (
	DEFAULT_MINE_POINT = -50
	DEFAULT_CELL_POINT = 1
	DEFAULT_FLAG_POINT = 0
	DEFAULT_ROW        = 20
	DEFAULT_COL        = 40
	DEFAULT_MINE_COUNT = 45
)

type Field struct {
	row        int
	col        int
	minesCount int
	// TODO: consider moving this somewhere else
	openCells int
	isStarted bool
	cells     [][]*Cell

	cellScore     int
	mineScore     int
	countColdOpen bool
}

type FieldBuilder struct {
	field *Field
}

func NewFieldBuilder() *FieldBuilder {
	return &FieldBuilder{
		field: &Field{
			row:           DEFAULT_ROW,
			col:           DEFAULT_COL,
			minesCount:    DEFAULT_MINE_COUNT,
			openCells:     0,
			isStarted:     false,
			cells:         [][]*Cell{},
			cellScore:     DEFAULT_CELL_POINT,
			mineScore:     DEFAULT_MINE_POINT,
			countColdOpen: false,
		},
	}
}

type difficultyConfig struct {
	col  int
	row  int
	mine int
}

var difficultyMap = map[string]difficultyConfig{
	"easy": {
		col:  5,
		row:  5,
		mine: 10,
	},
	"medium": {
		col:  10,
		row:  10,
		mine: 20,
	},
	"hard": {
		col:  20,
		row:  10,
		mine: 30,
	},
}

func (fb *FieldBuilder) WithDifficulty(diff string) *FieldBuilder {
	cfg := difficultyMap["hard"]
	if val, ok := difficultyMap[diff]; ok {
		cfg = val
	}
	fb.field.row = cfg.row
	fb.field.col = cfg.col
	fb.field.minesCount = cfg.mine
	return fb
}

func (fb *FieldBuilder) WithRow(row int) *FieldBuilder {
	fb.field.row = row
	return fb
}

func (fb *FieldBuilder) WithCol(col int) *FieldBuilder {
	fb.field.col = col
	return fb
}

func (fb *FieldBuilder) WithMinesCount(val int) *FieldBuilder {
	fb.field.minesCount = val
	return fb
}

func (fb *FieldBuilder) WithCellScore(val int) *FieldBuilder {
	fb.field.cellScore = val
	return fb
}

func (fb *FieldBuilder) WithMineScore(val int) *FieldBuilder {
	fb.field.mineScore = val
	return fb
}

func (fb *FieldBuilder) WithCountColdOpen(val bool) *FieldBuilder {
	fb.field.countColdOpen = val
	return fb
}

func (fb *FieldBuilder) Build() *Field {
	fb.field.cells = generateCells(fb.field.row, fb.field.col)
	return fb.field
}

func NewField(row, col, mines int) *Field {
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

func (f Field) IsCleared() bool {
	return f.openCells == f.row*f.col-f.minesCount
}

func (f Field) IsClearedForReal() bool {
	return f.GetOpenCellCount() == f.row*f.col-f.minesCount
}

func (f Field) GetOpenCellCount() int {
	// TODO: actually count the open cells
	result := 0
	for _, row := range f.cells {
		for _, cell := range row {
			if cell.isOpen {
				result++
			}
		}
	}
	return result
}

func (f Field) GetCells() [][]*Cell {
	return f.cells
}

func (f Field) GetCellString() *[][]string {
	result := make([][]string, f.row)
	for i, row := range f.cells {
		result[i] = make([]string, f.col)
		for j, cell := range row {
			result[i][j] = cell.GetValue()
		}
	}
	return &result
}

func (f Field) GetCellStringBare() *[][]string {
	result := make([][]string, f.row)
	for i, row := range f.cells {
		result[i] = make([]string, f.col)
		for j, cell := range row {
			result[i][j] = cell.GetValueBare()
		}
	}
	return &result
}

func (f Field) GetRow() int {
	return f.row
}

func (f Field) GetCol() int {
	return f.col
}

// OpenCell opens the cell at the given position.
func (f *Field) OpenCell(row, col int, playerID string) (int, error) {
	cell := f.cells[row][col]

	isColdOpen := f.openCells == 0
	isOpen := cell.isOpen
	points := 0

	if cell.isFlagged {
		return points, ErrOpenFlaggedCell
	}

	cell.Open(playerID)

	if !f.isStarted {
		f.isStarted = true
		genesisCoordinate := Location{
			row: row,
			col: col,
		}
		f.generateMines(genesisCoordinate)
		f.setAdjacentMinesCount()
	}

	if cell.isMine {
		points += DEFAULT_MINE_POINT
		return points, ErrOpenMine
	}

	if !isOpen {
		f.openCells++
		points++
	}

	adjacentFlagCount := f.getAdjacentFlagCount(row, col)
	var errQuickOpen error
	var quickOpenPoints int
	if int(cell.adjacentMines) == adjacentFlagCount {
		quickOpenPoints, errQuickOpen = f.QuickOpenCell(row, col, playerID)
		points += quickOpenPoints
	}

	if isColdOpen && !f.countColdOpen {
		return 0, errQuickOpen
	}
	return points, errQuickOpen
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

// ToggleFlagCell flags the cell at the given position.
// TODO: maybe consider doing the flag x mines count check?
func (f *Field) ToggleFlagCell(row, col int, playerID string) (*Cell, error) {
	cell := f.cells[row][col]

	if cell.isOpen {
		return nil, ErrFlagOpenedCell
	}

	cell.Flag(playerID)

	return cell, nil
}

// QuickOpenCell opens the cell at the given position and all the adjacent cells if the number of adjacent flagged cells is equal to the number of adjacent mines.
// TODO: finish function, see if we need to return more than just error (maybe all the newly open cell?)
func (f *Field) QuickOpenCell(row, col int, playerID string) (int, error) {
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

	points := 0
	for len(locationsToOpen) > 0 {
		loc := locationsToOpen[0]
		locationsToOpen = locationsToOpen[1:]

		if loc.row < 0 || loc.row >= f.row || loc.col < 0 || loc.col >= f.col {
			continue
		}

		cell := f.cells[loc.row][loc.col]

		if cell.isMine && !cell.isFlagged {
			points = DEFAULT_MINE_POINT
			return points, ErrOpenMine
		}

		if cell.isFlagged || cell.isOpen {
			continue
		}

		cell.Open(playerID)
		f.openCells++
		points += DEFAULT_CELL_POINT

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

	return points, nil
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
		return nil, ErrTooManyMines
	}

	cellHasMine := make(map[int]bool)
	for i := genesisCoordinate.row - 1; i <= genesisCoordinate.row+1; i++ {
		for j := genesisCoordinate.col - 1; j <= genesisCoordinate.col+1; j++ {
			cellHasMine[i*f.col+j] = true
		}
	}

	minesLocations := make([]Location, mines)
	for i := 0; i < mines; i++ {
		randomLoc := 0
		for {
			randomLoc = utils.GenerateRandomInt(0, cellCount)
			if !cellHasMine[randomLoc] {
				break
			}
		}

		cellHasMine[randomLoc] = true

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
	openerID      string
	flaggerID     string
}

func (c Cell) GetValueBare() string {
	if c.isMine {
		return "X"
	}

	result := strconv.Itoa(int(c.adjacentMines))

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

func (c *Cell) IsMine() bool {
	return c.isMine
}

func (c *Cell) Open(openerID string) {
	c.isOpen = true
	c.openerID = openerID
}

// TODO: maybe consider doing the flag x mines count check?
func (c *Cell) Flag(playerID string) {
	c.isFlagged = !c.isFlagged
	c.flaggerID = playerID
}

type Location struct {
	row int
	col int
}
