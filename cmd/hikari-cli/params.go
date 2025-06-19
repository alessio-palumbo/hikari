package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// ParamType defines a parameter for a command.
type ParamType struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Default     any
	Value       string
	Validator   func(value string) error
}

func (p ParamType) ValidateValue(v string) error {
	if p.Validator != nil {
		return p.Validator(v)
	}
	return nil
}

type paramItem struct {
	param *ParamType
}

func (i paramItem) FilterValue() string { return i.param.Name }

func (i paramItem) Title() string {
	label := i.param.Name
	if i.param.Required {
		label = label + " *"
	}
	return label
}
func (i paramItem) Description() string {
	if i.param.Value != "" {
		return i.param.Value
	}
	return i.param.Description
}

func NewParamsList(params []ParamType) list.Model {
	items := make([]list.Item, len(params))
	for i := range params {
		items[i] = paramItem{param: &params[i]}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 30, len(items)*7)
	l.SetShowHelp(false)
	l.SetStatusBarItemName("setting", "settings")
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
