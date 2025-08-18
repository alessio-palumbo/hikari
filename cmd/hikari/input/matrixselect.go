package input

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MatrixCellSelected   = "█"
	MatrixCellUnselected = "·"
)

type MatrixSelectModel struct {
	matrix        [][]bool
	height, width int
	cursorX       int
	cursorY       int
}

func NewMatrixSelect(width, height int) MatrixSelectModel {
	matrix := make([][]bool, height)
	for i := range matrix {
		matrix[i] = make([]bool, width)
	}
	return MatrixSelectModel{matrix: matrix, width: width, height: height}
}

func (m MatrixSelectModel) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.cursorX > 0 {
				m.cursorX--
			}
		case "right", "l":
			if m.cursorX < m.width-1 {
				m.cursorX++
			}
		case "up", "k":
			if m.cursorY > 0 {
				m.cursorY--
			}
		case "down", "j":
			if m.cursorY < m.height-1 {
				m.cursorY++
			}
		case " ", "t":
			if m.matrix[m.cursorY][m.cursorX] {
				m.matrix[m.cursorY][m.cursorX] = false
			} else {
				m.matrix[m.cursorY][m.cursorX] = true
			}
		}
	}
	return m, nil
}

func (m MatrixSelectModel) View() string {
	var b strings.Builder
	for y := range m.height {
		for x := range m.width {
			cell := MatrixCellUnselected
			if m.matrix[y][x] {
				cell = MatrixCellSelected
			}

			if x == m.cursorX && y == m.cursorY {
				cell = "[" + cell + "]"
			} else {
				cell = " " + cell + " "
			}
			b.WriteString(cell)
		}
		b.WriteRune('\n')
	}
	return b.String()
}

func (m MatrixSelectModel) Value() string {
	var b strings.Builder
	for y := range m.height {
		for x := range m.width {
			b.WriteString(" ")
			if m.matrix[y][x] {
				b.WriteString(MatrixCellSelected)
			} else {
				b.WriteString(MatrixCellUnselected)
			}
			b.WriteString(" ")
		}
		b.WriteRune('\n')
	}
	return b.String()
}

func (m MatrixSelectModel) Reset() Input {
	for y := range m.height {
		for x := range m.width {
			m.matrix[y][x] = false
		}
	}
	m.cursorX = 0
	m.cursorY = 0
	return m
}
