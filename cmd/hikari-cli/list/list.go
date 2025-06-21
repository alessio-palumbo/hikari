package list

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultDelegateHeight  = 1
	defaultDelegateSpacing = 0
)

func New[S ~[]E, E any](ss S, wrap func(E) list.Item, dl delegate) list.Model {
	items := make([]list.Item, len(ss))
	for i, s := range ss {
		items[i] = wrap(s)
	}

	l := list.New(items, dl, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return l
}

type RenderFunc func(w io.Writer, m list.Model, index int, listItem list.Item)

type delegate struct {
	f       RenderFunc
	height  int
	spacing int
}

func NewDelegate(f RenderFunc, opts ...DelegateOption) delegate {
	d := delegate{f: f,
		height:  defaultDelegateHeight,
		spacing: defaultDelegateSpacing,
	}

	for _, opt := range opts {
		opt(&d)
	}
	return d
}

func (d delegate) Height() int                             { return d.height }
func (d delegate) Spacing() int                            { return d.spacing }
func (d delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	d.f(w, m, index, listItem)
}
