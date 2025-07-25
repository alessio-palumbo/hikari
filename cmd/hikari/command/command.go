package command

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alessio-palumbo/hikari/cmd/hikari/input"
	hlist "github.com/alessio-palumbo/hikari/cmd/hikari/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	ctrl "github.com/alessio-palumbo/lifxlan-go/pkg/controller"
	"github.com/alessio-palumbo/lifxlan-go/pkg/matrix"
	"github.com/alessio-palumbo/lifxlan-go/pkg/messages"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
	"github.com/alessio-palumbo/lifxprotocol-go/gen/protocol/enums"
	"github.com/alessio-palumbo/lifxprotocol-go/gen/protocol/packets"
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
					if err != nil {
						return nil, err
					}
				}
			}

			return messages.SetColor(h, s, b, k, d, enums.LightWaveformLIGHTWAVEFORMSAW), err
		},
		ParamTypes: []paramType{
			{Name: "hue", InputType: input.InputText, Required: false, Description: "Hue (0-360)", Validator: HueValidator},
			{Name: "saturation", InputType: input.InputText, Required: false, Description: "Saturation (0-100)", Validator: PercentageValidator},
			{Name: "brightness", InputType: input.InputText, Required: false, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "kelvin", InputType: input.InputText, Required: false, Description: "Kelvin (1500-9000)", Validator: KelvinValidator},
			{Name: "duration", InputType: input.InputText, Required: false, Description: "Transition seconds", Validator: DurationValidator},
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
					if err != nil {
						return nil, err
					}
				}
			}

			return messages.SetColor(nil, nil, b, nil, d, enums.LightWaveformLIGHTWAVEFORMSAW), nil
		},
		ParamTypes: []paramType{
			{Name: "brightness", InputType: input.InputText, Required: true, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "duration", InputType: input.InputText, Required: false, Description: "Transition seconds", Validator: DurationValidator},
		},
	},
	{
		ID:          "waterfall_effect",
		Name:        "Waterfall Effect",
		Description: "Apply given colors sequentially row by row",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			var (
				mode, intervalMs, cycles int64 = 0, 100, 0
				colors                   []packets.LightHsbk
			)

			for _, p := range params {
				if v := p.GetValue(); v != "" {
					if p.Name == "colors" {
						for c := range strings.SplitSeq(v, ",") {
							colors = append(colors, packets.LightHsbk{
								Hue: colorNamesToHue[c], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
							})
						}
						continue
					}

					vv, err := parseInt64Input(v)
					if err != nil {
						return nil, err
					}

					switch p.Name {
					case "mode":
						mode = *vv
					case "send_interval":
						intervalMs = *vv
					case "cycles":
						cycles = *vv
					}
				}
			}

			return func() error {
				return matrix.Waterfall(m, send, intervalMs, int(cycles), matrix.ParseChainMode(int(mode)), colors...)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputText, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: EffectModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition", Validator: PositiveIntegerValidator},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "colors", InputType: input.InputMultiSelect, Required: true, Description: "Colors of the waterfall", Validator: ColorListValidator},
		},
	},
	{
		ID:          "snake_effect",
		Name:        "Snake Effect",
		Description: "Simulate a slithering snake",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			var (
				mode, intervalMs, cycles, size int64 = 0, 100, 0, 4
				color                          packets.LightHsbk
			)

			for _, p := range params {
				if v := p.GetValue(); v != "" {
					if p.Name == "color" {
						color = packets.LightHsbk{
							Hue: colorNamesToHue[v], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
						}
						continue
					}

					vv, err := parseInt64Input(v)
					if err != nil {
						return nil, err
					}

					switch p.Name {
					case "mode":
						mode = *vv
					case "send_interval":
						intervalMs = *vv
					case "cycles":
						cycles = *vv
					case "size":
						size = *vv
					}
				}
			}

			return func() error {
				return matrix.Snake(m, send, intervalMs, int(cycles), matrix.ParseChainMode(int(mode)), int(size), color)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputText, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: EffectModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition (default 100)", Validator: PositiveIntegerValidator},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "size", InputType: input.InputText, Required: false, Description: "The size of the snake (default 4)", Validator: PositiveIntegerValidator},
			{Name: "color", InputType: input.InputSingleSelect, Required: true, Description: "Color of the snake", Validator: ColorListValidator},
		},
	},
}

// Command represents a backend command with metadata
type Command struct {
	ID                  string
	Name                string
	Description         string
	Handler             func(args ...ParamItem) (*protocol.Message, error)
	MatrixEffectHandler func(m *matrix.Matrix, send matrix.SendFunc, args ...ParamItem) (func() error, error)
	EffectStopper       *atomic.Bool
	ParamTypes          []paramType
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

// StartMatrixEffect starts a matrix effect in a goroutine and returns handle to stop the effect.
// If validation fails it returns an error.
func (i Item) StartMatrixEffect(mProps ctrl.MatrixProperties, send matrix.SendFunc, args ...ParamItem) (*atomic.Bool, error) {
	m := matrix.New(int(mProps.Width), int(mProps.Height), int(mProps.ChainLength))
	sender, stopped := matrix.SendWithStop(send)
	f, err := i.MatrixEffectHandler(m, sender, args...)
	if err != nil {
		return nil, err
	}
	go func() {
		f()
		stopped.Store(true)
	}()
	return stopped, nil
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
				return style.ListSelected.Render(padFunc(s[0])) + style.ActionActive.Render(action)
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

func parseInt64Input(s string) (*int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value for int64")
	}
	return &v, nil
}

func parseDurationInput(s string) (time.Duration, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Duration(0), fmt.Errorf("invalid value for duration")
	}
	return time.Duration(v) * time.Second, nil
}
