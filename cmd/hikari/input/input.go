package input

import tea "github.com/charmbracelet/bubbletea"

type InputType int

const (
	InputText InputType = iota
	InputSingleSelect
	InputMultiSelect
)

type Input interface {
	Update(tea.Msg) (Input, tea.Cmd)
	View() string
	Value() string
}
