package input

import (
	"fmt"
	"strings"

	"github.com/alessio-palumbo/hikari/cmd/hikari/internal/utils"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	unselectedOption = " "
	selectedOption   = "✔"
	upArrow          = "↑"
	downArrow        = "↓"
)

type SingleSelectModel struct {
	options []string
	cursor  int
	inline  bool
	padFunc func(string) string
}

func NewInputSingleSelect(options []string, width int) SingleSelectModel {
	padFunc := utils.RightPadder(options, func(o string) int { return len(o) }, width-len(unselectedOption))
	return SingleSelectModel{options: options, padFunc: padFunc}
}

func NewInputSingleSelectInline(options []string, width int) SingleSelectModel {
	arrowSuffixWidth := 1
	padFunc := utils.RightPadder(options, func(o string) int { return len(o) }, width-arrowSuffixWidth)
	return SingleSelectModel{options: options, inline: true, padFunc: padFunc}
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
	if m.inline {
		return m.inlineView()
	}

	var b strings.Builder
	for i, opt := range m.options {
		selected := unselectedOption
		fn := style.ActionActive.Render
		if m.cursor == i {
			selected = selectedOption
			fn = style.ActionSelected.Render
		}
		fmt.Fprintln(&b, fn(selected, m.padFunc(opt)))
	}
	return b.String()
}

func (m SingleSelectModel) Value() string {
	return m.options[m.cursor]
}

func (m SingleSelectModel) inlineView() string {
	var b strings.Builder
	for i, opt := range m.options {
		if m.cursor == i {
			fn := style.ActionSelected.Render
			if i == len(m.options)-1 {
				fmt.Fprint(&b, fn(upArrow, m.padFunc(opt)))
			} else {
				fmt.Fprint(&b, fn(downArrow, m.padFunc(opt)))
			}
		}
	}
	return b.String()
}
