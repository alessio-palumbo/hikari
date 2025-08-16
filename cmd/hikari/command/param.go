package command

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/alessio-palumbo/hikari/cmd/hikari/input"
	"github.com/alessio-palumbo/hikari/cmd/hikari/internal/utils"
	hlist "github.com/alessio-palumbo/hikari/cmd/hikari/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	"github.com/alessio-palumbo/lifxlan-go/pkg/matrix"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultPadding = 5

	paramInputWidth = 20
	paramCharLimit  = 5

	chainModeSingle     = "single_device"
	chainModeSequential = "chain_sequential"
	chainModeSynced     = "chain_synced"

	directionInwards  = "inwards"
	directionOutwards = "outwards"
	directionInOut    = "in-out"
	directionOutIn    = "out-in"
)

var (
	optionModes     = []string{chainModeSingle, chainModeSequential, chainModeSynced}
	optionColors    = []string{"red", "orange", "green", "yellow", "cyan", "blue", "magenta", "purple"}
	optionDirection = []string{directionInwards, directionOutwards, directionInOut, directionOutIn}
)

var colorNamesToHue = map[string]uint16{
	"red":     0,     // 0°
	"orange":  8192,  // 45°
	"yellow":  10923, // 60°
	"green":   21845, // 120°
	"cyan":    32768, // 180°
	"blue":    43690, // 240°
	"purple":  49152, // 270°
	"magenta": 54613, // 300°
}

// paramType defines a parameter for a command.
type paramType struct {
	Name         string
	InputType    input.InputType
	InputOptions []string
	Required     bool
	Description  string
	Default      any
	Validator    func(value string) (any, error)
}

type ParamItem struct {
	*paramType
	value   any
	Editing bool
	Input   input.Input
}

func (i ParamItem) FilterValue() string { return i.Name }

func (p ParamItem) ValidateValue(v string) (any, error) {
	if p.Validator != nil {
		return p.Validator(v)
	}
	return nil, nil
}

func (i ParamItem) Title() string {
	return style.ListTitle.Render(fmt.Sprintf("Setting %s", i.Name))
}

func (p ParamItem) GetValue() string {
	if p.Input != nil {
		return p.Input.Value()
	}
	return ""
}

func (p *ParamItem) SetValue() error {
	v := p.Input.Value()
	// Reset field if empty
	if v == "" {
		p.value = nil
		return nil
	}

	value, err := p.ValidateValue(v)
	if err != nil {
		return err
	}
	p.value = value
	return nil
}

func (p *ParamItem) UpdateValue(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	p.Input, cmd = p.Input.Update(msg)
	return cmd
}

func (p *ParamItem) SetEdit(v bool) {
	if v {
		p.Editing = true
		switch p.InputType {
		case input.InputText:
			p.Input = input.NewInputText(paramInputWidth, paramCharLimit, p.Description)
		case input.InputSingleSelect:
			p.Input = input.NewInputSingleSelect(p.InputOptions, paramInputWidth)
		case input.InputSingleSelectInline:
			p.Input = input.NewInputSingleSelectInline(p.InputOptions, paramInputWidth)
		case input.InputMultiSelect:
			p.Input = input.NewMultiSelect(p.InputOptions, paramInputWidth)
		}
		return
	}
	p.Editing = false
}

func SetParamValue[T any](p ParamItem) T {
	v := p.value
	if v == nil {
		v = p.Default
	}
	// Attempt direct cast first
	if val, ok := v.(T); ok {
		return val
	}

	// Reflection fallback: check if it's a pointer to T
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		// Check if *v is T
		elem := rv.Elem().Interface()
		if val, ok := elem.(T); ok {
			return val
		}
	}

	// Return zero value of T if no match
	var zero T
	return zero
}

func ParamItemsFromModel(l list.Model) []ParamItem {
	items := l.Items()
	params := make([]ParamItem, len(items))
	for i, item := range items {
		params[i] = item.(ParamItem)
	}
	return params
}

func newParamsList(params []paramType) list.Model {
	padFunc := utils.RightPadder(params, func(p paramType) int { return len(p.Name) })
	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		item, ok := listItem.(ParamItem)
		if !ok {
			return
		}

		str := item.Name
		if item.Required {
			str = str + " *"
		}
		str = fmt.Sprintf("%s-> ", padFunc(str))

		var valueStr string
		if item.Editing {
			switch item.InputType {
			case input.InputText:
				str += item.Input.View()
			case input.InputSingleSelectInline:
				str += item.Input.View()
			case input.InputSingleSelect, input.InputMultiSelect:
				str = lipgloss.NewStyle().Render(lipgloss.JoinHorizontal(lipgloss.Top, str, item.Input.View()))
			}
		} else if v := item.GetValue(); v != "" {
			valueStr = "[" + v + "]"
			str += valueStr
		} else {
			valueStr = "[not set]"
			str += valueStr
		}

		sendLabelStyle := style.ActionActive
		for _, i := range m.Items() {
			p := i.(ParamItem)
			if p.Required && p.GetValue() == "" {
				sendLabelStyle = style.ActionBlurred
				break
			}
		}

		fn := style.ListItem.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				var padding int
				if !item.Editing {
					// Add an extra padding to cater for style.ListSelected applied left-padding.
					padding = paramInputWidth + 1 - len(valueStr)
				}
				editAction := style.ActionActive.PaddingLeft(padding).Render("[E]dit ")
				return style.ListSelected.Render(lipgloss.JoinHorizontal(lipgloss.Top, s[0], editAction, sendLabelStyle.Render("[S]end")))
			}
		}

		fmt.Fprint(w, fn(str))
	}
	d := hlist.NewDelegate(renderFunc)

	f := func(i paramType) list.Item { return ParamItem{paramType: &i} }
	l := hlist.New(params, f, d)
	l.SetHeight(len(commands)*2 + 1)

	return l
}

func ValidateRequired(params ...ParamItem) error {
	for _, p := range params {
		if p.Required && p.value == nil {
			return fmt.Errorf("%s must be set", p.Name)
		}
	}
	return nil
}

func HueValidator(v string) (any, error) {
	h, err := parseFloat64Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *h < 0 || *h > 360 {
		return nil, fmt.Errorf("value out of range (0-360)")
	}
	return h, nil
}

func PercentageValidator(v string) (any, error) {
	p, err := parseFloat64Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *p < 0 || *p > 100 {
		return nil, fmt.Errorf("value out of range (0-100)")
	}
	return p, nil
}

func KelvinValidator(v string) (any, error) {
	k, err := parseUint16Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *k < 1500 || *k > 9000 {
		return nil, fmt.Errorf("value out of range (1500-9000)")
	}
	return k, nil
}

func DurationValidator(v string) (any, error) {
	d, err := parseDurationInput(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if d > 24*time.Hour {
		return nil, fmt.Errorf("duration too long")
	}
	return d, nil
}

func EffectModeValidator(v string) (any, error) {
	m, err := parseInt64Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *m < 0 || *m > 2 {
		return nil, fmt.Errorf("mode must be between 0-2")
	}
	return m, nil
}

func CyclesValidator(v string) (any, error) {
	m, err := parseInt64Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *m < 0 {
		return nil, fmt.Errorf("cycles must 0 or greater")
	}
	return m, nil
}

func PositiveIntegerValidator(v string) (any, error) {
	m, err := parseInt64Input(v)
	if err != nil {
		return nil, fmt.Errorf("invalid value, must be a number")
	}
	if *m < 1 {
		return nil, fmt.Errorf("value must 1 or greater")
	}
	return m, nil
}

func ColorListValidator(v string) (any, error) {
	for s := range strings.SplitSeq(v, ",") {
		if _, ok := colorNamesToHue[s]; !ok {
			return nil, fmt.Errorf("invalid color name: %s", s)
		}
	}
	if len(v) == 0 {
		return nil, fmt.Errorf("value must not be empty")
	}
	return v, nil
}

func ChainModeValidator(v string) (any, error) {
	switch v {
	case chainModeSequential:
		return int(matrix.ChainModeSequential), nil
	case chainModeSynced:
		return int(matrix.ChainModeSynced), nil
	default:
		return int(matrix.ChainModeNone), nil
	}
}

func DirectionValidator(v string) (any, error) {
	switch v {
	case directionOutwards:
		return int(matrix.AnimationDirectionOutwards), nil
	case directionInOut:
		return int(matrix.AnimationDirectionInOut), nil
	case directionOutIn:
		return int(matrix.AnimationDirectionOutIn), nil
	default:
		return int(matrix.AnimationDirectionInwards), nil
	}
}
