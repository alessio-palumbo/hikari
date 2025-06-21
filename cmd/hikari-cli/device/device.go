package device

import (
	"fmt"
	"io"

	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/color"
	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	sstyle = lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 2)

	nstyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)
)

// Item implements the list.Item interface.
type Item client.Device

func (i Item) FilterValue() string {
	return i.Label
}

func (i Item) Title() string {
	if !i.PoweredOn {
		return fmt.Sprintf("⚫ %s", i.Label)
	} else if i.Type == client.DeviceTypeSwitch {
		return fmt.Sprintf("🔘 %s", i.Label)
	}

	var r, g, b int
	if i.Color.Saturation == 0.0 {
		r, g, b = color.KelvinToRGB(int(i.Color.Kelvin))
	} else {
		r, g, b = color.HSBToRGB(i.Color.Hue, i.Color.Saturation, i.Color.Brightness)
	}

	return fmt.Sprintf("%s %s", rgbColorBlock(r, g, b, "⬤"), i.Label)
}

func (i Item) Info() string {
	content := fmt.Sprintf("%s\n\nSerial: %s\nIP: %s\nFirmware: %s",
		i.Label,
		i.Serial,
		i.Address.IP,
		i.FirmwareVersion,
	)

	if i.Type != client.DeviceTypeSwitch {
		if i.PoweredOn {
			showKelvin := i.Color.Saturation < 1
			if showKelvin {
				content += fmt.Sprintf("\n\n🔆 %.0f%% 🌡 %dK",
					i.Color.Brightness,
					i.Color.Kelvin)
			} else {
				content += fmt.Sprintf("\n\n🔆 %.0f%% 🎨 %.0f° 💧 %.0f%%",
					i.Color.Brightness,
					i.Color.Hue,
					i.Color.Saturation)
			}
		}
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 2).
		MarginTop(2).
		Width(40).
		Align(lipgloss.Center)

	return boxStyle.Render(content)
}

func NewList(devices []client.Device) list.Model {
	items := make([]list.Item, len(devices))
	for i, device := range devices {
		items[i] = Item(device)
	}

	l := list.New(items, deviceDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetStatusBarItemName("device", "devices")
	return l
}

type deviceDelegate struct{}

func (d deviceDelegate) Height() int                             { return 1 }
func (d deviceDelegate) Spacing() int                            { return 1 }
func (d deviceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d deviceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	deviceItem, ok := listItem.(Item)
	if !ok {
		return
	}

	str := deviceItem.Title()

	fn := nstyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return sstyle.Render(s...)
		}
	}

	fmt.Fprint(w, fn(str))
}

func rgbColorBlock(r, g, b int, text string) string {
	color := color.RGBToLipglossColor(r, g, b)
	return lipgloss.NewStyle().Foreground(color).Padding(0, 1, 0, 0).Bold(true).Render(text)
}
