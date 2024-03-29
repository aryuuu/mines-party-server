package minesweeper

import "errors"

var (
	ErrCannotOpenFlaggedCell = errors.New("cannot open a flagged cell")
	ErrFlagOpenedCell        = errors.New("cannot flag an opened cell")
	ErrCannotFlagOpenedMine  = errors.New("cannot flag an opened mine")
	ErrCannotOpenMine        = errors.New("cannot open a mine")
	ErrOpenOpenedCell        = errors.New("cannot open an open cell")
	ErrOpenFlaggedCell       = errors.New("cannot open a flagged cell")
	ErrOpenMine              = errors.New("opened a mine")
	ErrTooManyMines          = errors.New("too many mines")
)
