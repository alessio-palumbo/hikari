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
	"github.com/alessio-palumbo/hikari/cmd/hikari/internal/utils"
	hlist "github.com/alessio-palumbo/hikari/cmd/hikari/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	ldevice "github.com/alessio-palumbo/lifxlan-go/pkg/device"
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
		Type:        CommandTypeSetter,
		Description: "Turn the device on",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return messages.SetPowerOn(), nil
		},
		ParamTypes: []paramType{},
	},
	{
		ID:          "power_off",
		Name:        "Power Off",
		Type:        CommandTypeSetter,
		Description: "Turn the device off",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			return messages.SetPowerOff(), nil
		},
		ParamTypes: []paramType{},
	},
	{
		ID:          "set_color",
		Name:        "Set Color",
		Type:        CommandTypeSetter,
		Description: "Change device color (HSB + Kelvin)",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}
			return messages.SetColor(SetParamValue[*float64](params[0]), SetParamValue[*float64](params[1]), SetParamValue[*float64](params[2]),
				SetParamValue[*uint16](params[3]), SetParamValue[time.Duration](params[4]), enums.LightWaveformLIGHTWAVEFORMSAW), nil
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
		Type:        CommandTypeSetter,
		Description: "Adjust device brightness",
		Handler: func(params ...ParamItem) (*protocol.Message, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}
			return messages.SetColor(nil, nil, SetParamValue[*float64](params[0]), nil,
				SetParamValue[time.Duration](params[1]), enums.LightWaveformLIGHTWAVEFORMSAW,
			), nil
		},
		ParamTypes: []paramType{
			{Name: "brightness", InputType: input.InputText, Required: true, Description: "Brightness (0-100)", Validator: PercentageValidator},
			{Name: "duration", InputType: input.InputText, Required: false, Description: "Transition seconds", Validator: DurationValidator},
		},
	},
	{
		ID:          "waterfall_effect",
		Name:        "Waterfall Effect",
		Type:        CommandTypeEffect,
		Description: "Apply given colors sequentially row by row",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}

			var colors []packets.LightHsbk
			for c := range strings.SplitSeq(SetParamValue[string](params[3]), ",") {
				colors = append(colors, packets.LightHsbk{
					Hue: colorNamesToHue[c], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
				})
			}

			return func() error {
				return matrix.Waterfall(
					m,
					send,
					SetParamValue[int64](params[1]),
					SetParamValue[int](params[2]),
					matrix.ParseChainMode(SetParamValue[int](params[0])),
					colors...,
				)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputSingleSelectInline, InputOptions: optionModes, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: ChainModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition", Validator: PositiveIntegerValidator, Default: int64(100)},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "colors", InputType: input.InputMultiSelect, InputOptions: optionColors, Required: true, Description: "Colors of the waterfall", Validator: ColorListValidator},
		},
	},
	{
		ID:          "rockets_effect",
		Name:        "Rockets Effect",
		Type:        CommandTypeEffect,
		Description: "Apply colors to a single pixel row by row",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}

			var colors []packets.LightHsbk
			for c := range strings.SplitSeq(SetParamValue[string](params[3]), ",") {
				colors = append(colors, packets.LightHsbk{
					Hue: colorNamesToHue[c], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
				})
			}

			return func() error {
				return matrix.Rockets(
					m,
					send,
					SetParamValue[int64](params[1]),
					SetParamValue[int](params[2]),
					matrix.ParseChainMode(SetParamValue[int](params[0])),
					colors...,
				)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputSingleSelectInline, InputOptions: optionModes, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: ChainModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition", Validator: PositiveIntegerValidator, Default: int64(100)},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "colors", InputType: input.InputMultiSelect, InputOptions: optionColors, Required: true, Description: "Colors of the rocket", Validator: ColorListValidator},
		},
	},
	{
		ID:          "snake_effect",
		Name:        "Snake Effect",
		Type:        CommandTypeEffect,
		Description: "Simulate a slithering snake",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}

			return func() error {
				return matrix.Snake(
					m,
					send,
					SetParamValue[int64](params[1]),
					SetParamValue[int](params[2]),
					matrix.ParseChainMode(SetParamValue[int](params[0])),
					SetParamValue[int](params[3]),
					packets.LightHsbk{
						Hue: colorNamesToHue[SetParamValue[string](params[4])], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
					},
				)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputSingleSelectInline, InputOptions: optionModes, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: ChainModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition (default 100)", Validator: PositiveIntegerValidator, Default: int64(100)},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "size", InputType: input.InputText, Required: false, Description: "The size of the snake (default 4)", Validator: PositiveIntegerValidator, Default: 4},
			{Name: "color", InputType: input.InputSingleSelect, InputOptions: optionColors, Required: true, Description: "Color of the snake", Validator: ColorListValidator},
		},
	},
	{
		ID:          "worm_effect",
		Name:        "Worm Effect",
		Type:        CommandTypeEffect,
		Description: "Simulate a crawling worm",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}

			return func() error {
				return matrix.Worm(
					m,
					send,
					SetParamValue[int64](params[1]),
					SetParamValue[int](params[2]),
					matrix.ParseChainMode(SetParamValue[int](params[0])),
					SetParamValue[int](params[3]),
					packets.LightHsbk{
						Hue: colorNamesToHue[SetParamValue[string](params[4])], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
					},
				)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputSingleSelectInline, InputOptions: optionModes, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: ChainModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition (default 100)", Validator: PositiveIntegerValidator, Default: int64(100)},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "size", InputType: input.InputText, Required: false, Description: "The size of the snake (default 4)", Validator: PositiveIntegerValidator, Default: 4},
			{Name: "color", InputType: input.InputSingleSelect, InputOptions: optionColors, Required: true, Description: "Color of the worm", Validator: ColorListValidator},
		},
	},
	{
		ID:          "concentric_frames_effect",
		Name:        "Concentric Frames Effect",
		Type:        CommandTypeEffect,
		Description: "Iterates according to the given direction drawing frames of variadic sizes",
		MatrixEffectHandler: func(m *matrix.Matrix, send matrix.SendFunc, params ...ParamItem) (func() error, error) {
			if err := ValidateRequired(params...); err != nil {
				return nil, err
			}

			var color *packets.LightHsbk
			if v := SetParamValue[*string](params[4]); v != nil {
				color = &packets.LightHsbk{
					Hue: colorNamesToHue[*v], Saturation: math.MaxUint16, Brightness: math.MaxUint16, Kelvin: 3500,
				}
			}
			return func() error {
				return matrix.ConcentricFrames(
					m,
					send,
					SetParamValue[int64](params[1]),
					SetParamValue[int](params[2]),
					matrix.ParseChainMode(SetParamValue[int](params[0])),
					matrix.ParseAnimationDirection(SetParamValue[int](params[3])),
					color,
				)
			}, nil
		},
		ParamTypes: []paramType{
			{Name: "mode", InputType: input.InputSingleSelectInline, InputOptions: optionModes, Required: false, Description: "0-(No chain), 1-(Chain sequential), 2-(Chain synced)", Validator: ChainModeValidator},
			{Name: "send_interval", InputType: input.InputText, Required: false, Description: "Ms pause between transition (default 200)", Validator: PositiveIntegerValidator, Default: int64(200)},
			{Name: "cycles", InputType: input.InputText, Required: false, Description: "Times the animation runs for (0 = forever)", Validator: CyclesValidator},
			{Name: "direction", InputType: input.InputSingleSelect, InputOptions: optionDirection, Required: false, Description: "The direction of the animation", Validator: DirectionValidator},
			{Name: "color", InputType: input.InputSingleSelect, InputOptions: optionColors, Required: false, Description: "Color of the frames", Validator: ColorListValidator},
		},
	},
}

type commandType int

const (
	CommandTypeSetter commandType = iota
	CommandTypeEffect
)

// Command represents a backend command with metadata
type Command struct {
	ID                  string
	Name                string
	Type                commandType
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
func (i Item) StartMatrixEffect(mProps ldevice.MatrixProperties, send matrix.SendFunc, args ...ParamItem) (*atomic.Bool, error) {
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
	padFunc := utils.RightPadder(commands, func(c Command) int { return len(c.Name) })
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
