package command

import (
	"fmt"
	"io"
	"strings"
	"time"

	hlist "github.com/alessio-palumbo/hikari/cmd/hikari-cli/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/style"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	defaultPaddingAfterName = 5
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
		p.Input.Width = 20
		p.Input.CharLimit = 5
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
	// Calculate longest name for padding value.
	var longestName int
	for _, p := range params {
		l := len(p.Name)
		if l > longestName {
			longestName = l
		}
	}

	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		item, ok := listItem.(ParamItem)
		if !ok {
			return
		}

		str := item.Name
		if item.Required {
			str = str + " *"
		}
		padding := longestName + defaultPaddingAfterName - len(str)

		var valueStr string
		if item.Editing {
			valueStr = item.Input.View()
		} else if v := item.GetValue(); v != "" {
			valueStr = "[" + v + "]"
		} else {
			valueStr = "[not set]"
		}

		str = fmt.Sprintf("%s%s-> %s", str, strings.Repeat(" ", padding), valueStr)

		fn := style.ListItem.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return style.ListSelected.Render(s...)
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
