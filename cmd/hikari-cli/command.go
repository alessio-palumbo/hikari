package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/alessio-palumbo/hikari/internal/protocol"
	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var commands = []Command{
	{
		ID:          "power_on",
		Name:        "Power On",
		Description: "Turn the device on",
		Handler: func(params ...ParamType) (*protocol.Message, error) {
			return client.SetPowerOn(), nil
		},
		ParamTypes: []ParamType{},
	},
	{
		ID:          "power_off",
		Name:        "Power Off",
		Description: "Turn the device off",
		Handler: func(params ...ParamType) (*protocol.Message, error) {
			return client.SetPowerOff(), nil
		},
		ParamTypes: []ParamType{},
	},
	{
		ID:          "set_color",
		Name:        "Set Color",
		Description: "Change device color (HSB + Kelvin)",
		Handler: func(params ...ParamType) (*protocol.Message, error) {
			var (
				h, s, b *float64
				k       *uint16
				d       time.Duration
				err     error
			)

			for _, p := range params {
				if v := p.Value; v != "" {
					switch p.Name {
					case "hue":
						h, err = parseFloat64Input(p.Value)
					case "saturation":
						s, err = parseFloat64Input(p.Value)
					case "brightness":
						b, err = parseFloat64Input(p.Value)
					case "kelvin":
						k, err = parseUint16Input(p.Value)
					case "duration":
						d, err = parseDurationInput(p.Value)
					}
				}
			}

			return client.SetColor(h, s, b, k, d), err
		},
		ParamTypes: []ParamType{
			{Name: "hue", Type: "float64", Required: false, Description: "Hue (0-360)", Validator: HueValidator},
			{Name: "saturation", Type: "float64", Required: false, Description: "Saturation (0-100)", Validator: PercentageValidator},
			{Name: "brightness", Type: "float64", Required: false, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "kelvin", Type: "uint16", Required: false, Description: "Kelvin (1500-9000)", Validator: KelvinValidator},
			{Name: "duration", Type: "duration", Required: false, Description: "Transition seconds"},
		},
	},
	{
		ID:          "set_brightness",
		Name:        "Set Brightness",
		Description: "Adjust device brightness",
		Handler: func(params ...ParamType) (*protocol.Message, error) {
			var (
				b   *float64
				d   time.Duration
				err error
			)

			for _, p := range params {
				if v := p.Value; v != "" {
					switch p.Name {
					case "brightness":
						b, err = parseFloat64Input(p.Value)
					case "duration":
						d, err = parseDurationInput(p.Value)
					}
				}
			}

			return client.SetColor(nil, nil, b, nil, d), err
		},
		ParamTypes: []ParamType{
			{Name: "brightness", Type: "float64", Required: true, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "duration", Type: "duration", Required: false, Description: "Transition seconds"},
		},
	},
}

// Command represents a backend command with metadata
type Command struct {
	ID          string
	Name        string
	Description string
	Handler     func(args ...ParamType) (*protocol.Message, error)
	ParamTypes  []ParamType
}

// Implement list.Item interface for Command
func (c Command) FilterValue() string {
	return c.Name + " " + c.Description
}

// CommandRegistry manages available commands
type CommandRegistry struct {
	commandList []Command
	commandsMap map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commandsMap: make(map[string]Command),
	}

	// Register all available commands
	for _, cmd := range commands {
		registry.commandsMap[cmd.ID] = cmd
	}

	return registry
}

func (r *CommandRegistry) GetCommand(id string) (Command, bool) {
	cmd, exists := r.commandsMap[id]
	return cmd, exists
}

func (r *CommandRegistry) ListCommands() []Command {
	return commands
}

func (r *CommandRegistry) ExecuteCommand(id string, args ...ParamType) (*protocol.Message, error) {
	cmd, exists := r.commandsMap[id]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", id)
	}

	msg, err := cmd.Handler(args...)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	// helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func NewCommandList() list.Model {
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = cmd
	}

	l := list.New(items, itemDelegate{}, 20, 14)
	l.Title = "Select a Command"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true) // Enable filtering by command name/description
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	// l.Styles.HelpStyle = helpStyle
	return l
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	cmd, ok := listItem.(Command)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s - %s", cmd.Name, cmd.Description)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func parseFloat64Input(s string) (*float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value for float64")
	}
	return &v, nil
}

func parseUint16Input(s string) (*uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid value for uint16")
	}
	vv := uint16(v)
	return &vv, nil
}

func parseDurationInput(s string) (time.Duration, error) {
	v, err := time.ParseDuration(s)
	if err != nil {
		return time.Duration(0), fmt.Errorf("invalid value for duration")
	}
	return v, nil
}
