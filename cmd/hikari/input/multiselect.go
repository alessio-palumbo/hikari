package input

import (
	"fmt"
	"strings"

	"github.com/alessio-palumbo/hikari/cmd/hikari/internal/utils"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	unselectedCheckbox = "[ ]"
	selectedCheckbox   = "[✔]"
)

type MultiSelectItem struct {
	Label   string
	Checked bool
}

type MultiSelectModel struct {
	items   []MultiSelectItem
	cursor  int
	padFunc func(string) string
	width   int
}

func NewMultiSelect(options []string, width int) MultiSelectModel {
	padFunc := utils.RightPadder(options, func(o string) int { return len(o) }, width-len(unselectedCheckbox))
	items := make([]MultiSelectItem, len(options))
	for i, opt := range options {
		items[i] = MultiSelectItem{Label: opt}
	}
	return MultiSelectModel{items: items, padFunc: padFunc, width: width}
}

func (m MultiSelectModel) Update(msg tea.Msg) (Input, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ", "t":
			m.items[m.cursor].Checked = !m.items[m.cursor].Checked
		}
	}
	return m, nil
}

func (m MultiSelectModel) View() string {
	var b strings.Builder
	for i, item := range m.items {
		checkbox := unselectedCheckbox
		if item.Checked {
			checkbox = selectedCheckbox
		}
		fn := style.ActionActive.Render
		if m.cursor == i {
			fn = style.ActionSelected.Render
		}
		fmt.Fprintln(&b, fn(checkbox, m.padFunc(item.Label)))
	}
	return b.String()
}

func (m MultiSelectModel) Value() string {
	var labels []string
	for _, l := range m.items {
		if l.Checked {
			labels = append(labels, l.Label)
		}
	}
	return strings.Join(labels, ",")
}
