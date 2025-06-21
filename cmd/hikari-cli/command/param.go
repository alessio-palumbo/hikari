package command

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	items := make([]list.Item, len(params))
	for i, p := range params {
		items[i] = ParamItem{&p}
	}

	// delegate := list.NewDefaultDelegate()
	// delegate.SetHeight(5)
	l := list.New(items, paramDelegate{}, 0, len(items)*2+1)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	// l.SetStatusBarItemName("setting", "settings")
	return l
}

type paramDelegate struct{}

func (d paramDelegate) Height() int                             { return 1 }
func (d paramDelegate) Spacing() int                            { return 0 }
func (d paramDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d paramDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ParamItem)
	if !ok {
		return
	}

	name := item.Name
	if item.Required {
		name = name + " *"
	}
	str := fmt.Sprintf("%s - %s", name, item.Description)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
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
