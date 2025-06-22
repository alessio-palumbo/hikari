package style

import "github.com/charmbracelet/lipgloss"

var (
	SelectedColor   = lipgloss.AdaptiveColor{Light: "#ee6ff8", Dark: "#ee6ff8"}
	SelectedBorder  = lipgloss.AdaptiveColor{Light: "#f793ff", Dark: "#ad58b4"}
	ListColor       = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}
	StatusColor     = lipgloss.Color("#888888")
	HelpColor       = lipgloss.Color("#626262")
	TitleBackground = lipgloss.Color("#5f5fd7")
	TitleColor      = lipgloss.Color("#ffffd5")
	ListTitleColor  = lipgloss.Color("#f25d94")
	TODO1Color      = lipgloss.Color("#874bfd")
	TODO3Color      = lipgloss.Color("#25a065")
)

var (
	ListSelected = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.Border{Left: ">"}, false, false, false, true).
			BorderForeground(SelectedBorder).
			Foreground(SelectedColor).
			PaddingLeft(1)

	ListItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(ListColor).
			PaddingLeft(2)

	Status = lipgloss.NewStyle().Foreground(StatusColor)

	Help = lipgloss.NewStyle().Foreground(HelpColor)

	Title = lipgloss.NewStyle().
		Background(TitleBackground).
		Foreground(TitleColor).
		Padding(0, 1)

	ListTitle = lipgloss.NewStyle().
			MarginLeft(2).
			Underline(true)
)
