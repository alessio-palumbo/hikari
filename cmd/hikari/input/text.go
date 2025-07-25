package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextModel struct {
	textinput.Model
}

func NewInputText(width, maxChars int, placeholder string) TextModel {
	i := textinput.New()
	i.Prompt = ""
	i.Width = width
	i.CharLimit = maxChars
	i.Placeholder = placeholder
	i.Focus()
	return TextModel{i}
}

func (m TextModel) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m TextModel) View() string {
	return m.Model.View()
}

func (m TextModel) Value() string {
	return m.Model.Value()
}
