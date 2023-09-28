package main

import (
	"fmt"
	"log"

	"github.com/aryuuu/mines-party-server/minesweeper"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

var (
	// Available spinners
	cellWidth  = 5
	cellHeight = 3
	greyColor  = lipgloss.Color("241")
	cyanColor  = lipgloss.Color("69")
	// need BG, grey border, DONE
	closedCellStyle = lipgloss.NewStyle().
			Width(cellWidth).
			Height(cellHeight).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder()).
			Background(greyColor)
	// no BG, grey border, DONE
	openedCellStyle = lipgloss.NewStyle().
			Width(cellWidth).
			Height(cellHeight).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(greyColor)
	// no BG, cyan border, DONE
	focusedOpenedCellStyle = lipgloss.NewStyle().
				Width(cellWidth).
				Height(cellHeight).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(cyanColor)
	// need BG, cyan border
	focusedClosedCellStyle = lipgloss.NewStyle().
				Width(cellWidth).
				Height(cellHeight).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(cyanColor).
				Background(greyColor)
)

type mainModel struct {
	field  *minesweeper.Field
	cursor struct{ row, col int }
}

func newModel() mainModel {
	m := mainModel{
		field: minesweeper.New(8, 8, 10),
		cursor: struct {
			row int
			col int
		}{row: 0, col: 0},
	}
	return m
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	// return tea.Batch(m.timer.Init(), m.spinner.Tick)
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down", "j":
			// TODO: go down one cell if possible
			if m.cursor.row < m.field.GetRow()-1 {
				m.cursor.row++
			}
		case "up", "k":
			// TODO: go up one cell if possible
			if m.cursor.row > 0 {
				m.cursor.row--
			}
		case "left", "h":
			// TODO: go left one cell if possible
			if m.cursor.col > 0 {
				m.cursor.col--
			}
		case "right", "l":
			// TODO: go right one cell if possible
			if m.cursor.col < m.field.GetCol()-1 {
				m.cursor.col++
			}
		case "space":
			// TODO: open cell
		case "ctrl+space", "shift+space":
			// TODO: flag/unflag cell
		}

		// switch m.state {
		// // update whichever model is focused
		// case spinnerView:
		// 	m.spinner, cmd = m.spinner.Update(msg)
		// 	cmds = append(cmds, cmd)
		// default:
		// 	m.timer, cmd = m.timer.Update(msg)
		// 	cmds = append(cmds, cmd)
		// }
		// case spinner.TickMsg:
		// 	m.spinner, cmd = m.spinner.Update(msg)
		// 	cmds = append(cmds, cmd)
		// case timer.TickMsg:
		// 	m.timer, cmd = m.timer.Update(msg)
		// 	cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	cellStrings := []string{}
	for i, row := range m.field.GetCells() {
		rowStrings := []string{}
		for j, cell := range row {
			if cell.IsOpen() {
				if m.cursor.row == i && m.cursor.col == j {
					rowStrings = append(rowStrings, focusedOpenedCellStyle.Render(fmt.Sprintf("%4s", cell.GetValue())))
				} else {
					rowStrings = append(rowStrings, openedCellStyle.Render(fmt.Sprintf("%4s", cell.GetValue())))
				}
			} else {
				if m.cursor.row == i && m.cursor.col == j {
					rowStrings = append(rowStrings, focusedClosedCellStyle.Render(fmt.Sprintf("%4s", cell.GetValue())))
				} else {
					rowStrings = append(rowStrings, closedCellStyle.Render(fmt.Sprintf("%4s", cell.GetValue())))
				}
			}
		}
		cellStrings = append(cellStrings, lipgloss.JoinHorizontal(lipgloss.Top, rowStrings...))
	}

	return lipgloss.JoinVertical(lipgloss.Top, cellStrings...)
}

// func (m mainModel) currentFocusedModel() string {
// 	if m.state == timerView {
// 		return "timer"
// 	}
// 	return "spinner"
// }

// func (m *mainModel) Next() {
// 	if m.index == len(spinners)-1 {
// 		m.index = 0
// 	} else {
// 		m.index++
// 	}
// }

// func (m *mainModel) resetSpinner() {
// 	m.spinner = spinner.New()
// 	m.spinner.Style = spinnerStyle
// 	m.spinner.Spinner = spinners[m.index]
// }

func main() {
	p := tea.NewProgram(newModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
