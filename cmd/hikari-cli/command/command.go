package command

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

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065"))
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	// helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

var commands = []Command{
	{
		ID:          "power_on",
		Name:        "Power On",
		Description: "Turn the device on",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return client.SetPowerOn(), nil
		},
		ParamTypes: []paramType{},
	},
	{
		ID:          "power_off",
		Name:        "Power Off",
		Description: "Turn the device off",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return client.SetPowerOff(), nil
		},
		ParamTypes: []paramType{},
	},
	{
		ID:          "set_color",
		Name:        "Set Color",
		Description: "Change device color (HSB + Kelvin)",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			var (
				h, s, b *float64
				k       *uint16
				d       time.Duration
				err     error
			)

			for _, p := range params {
				if v := p.GetValue(); v != "" {
					switch p.Name {
					case "hue":
						h, err = parseFloat64Input(v)
					case "saturation":
						s, err = parseFloat64Input(v)
					case "brightness":
						b, err = parseFloat64Input(v)
					case "kelvin":
						k, err = parseUint16Input(v)
					case "duration":
						d, err = parseDurationInput(v)
					}
				}
			}

			return client.SetColor(h, s, b, k, d), err
		},
		ParamTypes: []paramType{
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
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			var (
				b   *float64
				d   time.Duration
				err error
			)

			for _, p := range params {
				if v := p.GetValue(); v != "" {
					switch p.Name {
					case "brightness":
						b, err = parseFloat64Input(v)
					case "duration":
						d, err = parseDurationInput(v)
					}
				}
			}

			return client.SetColor(nil, nil, b, nil, d), err
		},
		ParamTypes: []paramType{
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
	Handler     func(args ...ParamItem) (*protocol.Message, error)
	ParamTypes  []paramType
}

// Item implements list.Item interface.
type Item Command

func (i Item) FilterValue() string {
	return i.Name
}

func (i Item) Title() string {
	return titleStyle.Render(i.Name)
}

func (i Item) NewParams() list.Model {
	return newParamsList(i.ParamTypes)
}

func NewList() list.Model {
	items := make([]list.Item, len(commands))
	for i, c := range commands {
		items[i] = Item(c)
	}

	// l := list.New(items, commandDelegate{}, 20, 14)
	l := list.New(items, commandDelegate{}, 0, len(items)*2)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true) // Enable filtering by command name/description
	// l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	// l.Styles.HelpStyle = helpStyle
	return l
}

type commandDelegate struct{}

func (d commandDelegate) Height() int                             { return 1 }
func (d commandDelegate) Spacing() int                            { return 0 }
func (d commandDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d commandDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s - %s", item.Name, item.Description)

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
