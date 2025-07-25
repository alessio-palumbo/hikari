package input

import (
	"fmt"
	"strings"

	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	tea "github.com/charmbracelet/bubbletea"
)

type SingleSelectModel struct {
	options []string
	cursor  int
}

func NewInputSingleSelect(options []string) SingleSelectModel {
	return SingleSelectModel{options: options}
}

func (m SingleSelectModel) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m SingleSelectModel) View() string {
	var b strings.Builder
	for i, opt := range m.options {
		selected := "  "
		fn := style.ActionActive.Render
		if m.cursor == i {
			selected = "✔ "
			fn = style.ActionSelected.Render
		}
		fmt.Fprintf(&b, "%s%s\n", fn(selected), fn(opt))
	}
	return b.String()
}

func (m SingleSelectModel) Value() string {
	return m.options[m.cursor]
}
