package command

import (
	"fmt"
	"io"
	"time"

	hlist "github.com/alessio-palumbo/hikari/cmd/hikari-cli/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/style"
	"github.com/charmbracelet/bubbles/list"
)

// paramType defines a parameter for a command.
type paramType struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Default     any
	Validator   func(value string) error
	value       *string
}

type ParamItem struct {
	*paramType
}

func (i ParamItem) FilterValue() string { return i.Name }

func (p ParamItem) ValidateValue(v string) error {
	if p.Validator != nil {
		return p.Validator(v)
	}
	return nil
}

func (p ParamItem) GetValue() string {
	if p.paramType != nil && p.value != nil {
		return *p.value
	}
	return ""
}
func (p *ParamItem) SetValue(v string) {
	pv := &v
	p.value = pv
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
	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		item, ok := listItem.(ParamItem)
		if !ok {
			return
		}

		name := item.Name
		if item.Required {
			name = name + " *"
		}
		str := fmt.Sprintf("%s - %s", name, item.Description)

		fn := style.ListItem.Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return style.ListSelected.Render(s...)
			}
		}

		fmt.Fprint(w, fn(str))
	}
	d := hlist.NewDelegate(renderFunc)

	f := func(i paramType) list.Item { return ParamItem{&i} }
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
