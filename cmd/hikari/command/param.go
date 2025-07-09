package command

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"time"

	hlist "github.com/alessio-palumbo/hikari/cmd/hikari/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	defaultPadding = 5

	paramInputWidth = 20
	paramCharLimit  = 5
)

// paramType defines a parameter for a command.
type paramType struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Default     any
	Validator   func(value string) error
}

type ParamItem struct {
	*paramType
	value   *string
	Editing bool
	Input   textinput.Model
}

func (i ParamItem) FilterValue() string { return i.Name }

func (p ParamItem) ValidateValue(v string) error {
	if p.Validator != nil {
		return p.Validator(v)
	}
	return nil
}

func (i ParamItem) Title() string {
	return style.ListTitle.Render(fmt.Sprintf("Setting %s", i.Name))
}

func (p ParamItem) GetValue() string {
	if p.paramType != nil && p.value != nil {
		return *p.value
	}
	return ""
}
func (p *ParamItem) SetValue(v string) error {
	// Reset field if empty
	if v == "" {
		p.value = nil
		return nil
	}

	if err := p.ValidateValue(v); err != nil {
		return err
	}
	p.value = &v
	return nil
}

func (p *ParamItem) SetEdit(v bool) {
	if v {
		p.Editing = true
		p.Input = textinput.New()
		p.Input.Prompt = ""
		p.Input.Width = paramInputWidth
		p.Input.CharLimit = paramCharLimit
		p.Input.Placeholder = p.Description
		p.Input.Focus()
		return
	}
	p.Editing = false
}

func ParamItemsFromModel(l list.Model) []ParamItem {
	items := l.Items()
	params := make([]ParamItem, len(items))
	for _, i := range items {
		params = append(params, i.(ParamItem))
	}
	return params
}

func newParamsList(params []paramType) list.Model {
	padFunc := rightPadder(params, func(p paramType) int { return len(p.Name) })
	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		item, ok := listItem.(ParamItem)
		if !ok {
			return
		}

		str := item.Name
		if item.Required {
			str = str + " *"
		}

		var valueStr string
		if item.Editing {
			valueStr = item.Input.View()
		} else if v := item.GetValue(); v != "" {
			valueStr = "[" + v + "]"
		} else {
			valueStr = "[not set]"
		}

		str = fmt.Sprintf("%s-> %s", padFunc(str), valueStr)

		sendLabelStyle := style.ActionFocused
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
					padding = paramInputWidth + 1 - len(valueStr)
				}
				editAction := style.ActionFocused.PaddingLeft(padding).Render("[E]dit")
				return style.ListSelected.Render(s[0], editAction, sendLabelStyle.Render("[S]end"))
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

func HueValidator(v string) error {
	h, err := parseFloat64Input(v)
	if err != nil {
		return fmt.Errorf("invalid value, must be a number")
	}
	if *h < 0 || *h > 360 {
		return fmt.Errorf("value out of range (0-360)")
	}
	return nil
}

func PercentageValidator(v string) error {
	p, err := parseFloat64Input(v)
	if err != nil {
		return fmt.Errorf("invalid value, must be a number")
	}
	if *p < 0 || *p > 100 {
		return fmt.Errorf("value out of range (0-100)")
	}
	return nil
}

func KelvinValidator(v string) error {
	k, err := parseUint16Input(v)
	if err != nil {
		return fmt.Errorf("invalid value, must be a number")
	}
	if *k < 1500 || *k > 9000 {
		return fmt.Errorf("value out of range (1500-9000)")
	}
	return nil
}

func DurationValidator(v string) error {
	d, err := parseDurationInput(v)
	if err != nil {
		return fmt.Errorf("invalid value, must be a number")
	}
	if d > 24*time.Hour {
		return fmt.Errorf("duration too long")
	}
	return nil
}

func rightPadder[S ~[]E, E any](ss S, lenFunc func(E) int) func(s string) string {
	longest := slices.MaxFunc(ss, func(a, b E) int {
		return cmp.Compare(lenFunc(a), lenFunc(b))
	})
	maxPadding := lenFunc(longest) + defaultPadding
	return func(s string) string {
		return fmt.Sprintf("%-*s", maxPadding, s)
	}
}
