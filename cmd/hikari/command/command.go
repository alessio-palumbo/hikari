package command

import (
	"fmt"
	"io"
	"strconv"
	"time"

	hlist "github.com/alessio-palumbo/hikari/cmd/hikari/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	"github.com/alessio-palumbo/lifxlan-go/pkg/messages"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
	"github.com/alessio-palumbo/lifxprotocol-go/gen/protocol/enums"
	"github.com/charmbracelet/bubbles/list"
)

var commands = []Command{
	{
		ID:          "power_on",
		Name:        "Power On",
		Description: "Turn the device on",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return messages.SetPowerOn(), nil
		},
		ParamTypes: []paramType{},
	},
	{
		ID:          "power_off",
		Name:        "Power Off",
		Description: "Turn the device off",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return messages.SetPowerOff(), nil
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

			return messages.SetColor(h, s, b, k, d, enums.LightWaveformLIGHTWAVEFORMSAW), err
		},
		ParamTypes: []paramType{
			{Name: "hue", Type: "float64", Required: false, Description: "Hue (0-360)", Validator: HueValidator},
			{Name: "saturation", Type: "float64", Required: false, Description: "Saturation (0-100)", Validator: PercentageValidator},
			{Name: "brightness", Type: "float64", Required: false, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "kelvin", Type: "uint16", Required: false, Description: "Kelvin (1500-9000)", Validator: KelvinValidator},
			{Name: "duration", Type: "duration", Required: false, Description: "Transition seconds", Validator: DurationValidator},
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

			return messages.SetColor(nil, nil, b, nil, d, enums.LightWaveformLIGHTWAVEFORMSAW), err
		},
		ParamTypes: []paramType{
			{Name: "brightness", Type: "float64", Required: true, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "duration", Type: "duration", Required: false, Description: "Transition seconds", Validator: DurationValidator},
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
	return style.ListTitle.Render(i.Name)
}

func (i Item) NewParams() list.Model {
	return newParamsList(i.ParamTypes)
}

func NewList() list.Model {
	padFunc := rightPadder(commands, func(c Command) int { return len(c.Name) })
	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		item, ok := listItem.(Item)
		if !ok {
			return
		}

		fn := style.ListItem.Render
		if index == m.Index() {
			action := "[S]end"
			if len(item.ParamTypes) > 0 {
				action = "[E]dit"
			}

			fn = func(s ...string) string {
				return style.ListSelected.Render(padFunc(s[0])) + style.ActionFocused.Render(action)
			}
		}

		fmt.Fprint(w, fn(item.Name))
	}

	d := hlist.NewDelegate(renderFunc)

	f := func(i Command) list.Item { return Item(i) }
	l := hlist.New(commands, f, d)
	l.SetHeight(len(commands) * 2)

	return l
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
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Duration(0), fmt.Errorf("invalid value for duration")
	}
	return time.Duration(v) * time.Second, nil
}
