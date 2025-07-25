package style

import "github.com/charmbracelet/lipgloss"

var (
	SelectedTextColor   = lipgloss.AdaptiveColor{Light: "#ee6ff8", Dark: "#ee6ff8"}
	SelectedBorderColor = lipgloss.AdaptiveColor{Light: "#f793ff", Dark: "#ad58b4"}
	ListColor           = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}
	StatusColor         = lipgloss.Color("#888888")
	HelpColor           = lipgloss.Color("#626262")
	TitleBackground     = lipgloss.Color("#5f5fd7")
	TitleColor          = lipgloss.Color("#ffffd5")
	ListTitleColor      = lipgloss.Color("#aa38c7")
	ActionActiveColor   = lipgloss.Color("#008eff")
	TODO1Color          = lipgloss.Color("#874bfd")
	TODO2Color          = lipgloss.Color("#ee6ff8")
	TODO3Color          = lipgloss.Color("#58b7ff")
)

var (
	ListSelected = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.Border{Left: "┃"}, false, false, false, true).
			BorderForeground(SelectedBorderColor).
			Foreground(SelectedTextColor).
			PaddingLeft(1)

	ListItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(ListColor).
			PaddingLeft(2)

	ActionSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(SelectedTextColor)

	ActionActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ActionActiveColor)

	ActionBlurred = lipgloss.NewStyle().
			Bold(true).
			Foreground(ListColor).
			Faint(true)

	Status = lipgloss.NewStyle().Foreground(StatusColor)

	Help = lipgloss.NewStyle().Foreground(HelpColor)

	Spinner = lipgloss.NewStyle().
		Bold(true).
		Foreground(TitleBackground)

	Title = lipgloss.NewStyle().
		Background(TitleBackground).
		Foreground(TitleColor).
		Padding(0, 1)

	ListTitle = lipgloss.NewStyle().
			Background(ListTitleColor).
			Foreground(TitleColor).
			Padding(0, 1).
			MarginLeft(2)

	SelectedDevice = lipgloss.NewStyle().
			Bold(true).
			Foreground(ListColor)

	SelectedBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(SelectedBorderColor).
			MarginLeft(2)
)
