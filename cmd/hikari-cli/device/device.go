package device

import (
	"fmt"
	"io"

	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/color"
	hlist "github.com/alessio-palumbo/hikari/cmd/hikari-cli/list"
	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/style"
	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Item implements the list.Item interface.
type Item client.Device

func (i Item) FilterValue() string {
	return i.Label
}

func (i Item) StateSphere() string {
	if !i.PoweredOn {
		return "⚫"
	} else if i.Type == client.DeviceTypeSwitch {
		return "🔘"
	}

	var r, g, b int
	if i.Color.Saturation == 0.0 {
		r, g, b = color.KelvinToRGB(int(i.Color.Kelvin))
	} else {
		r, g, b = color.HSBToRGB(i.Color.Hue, i.Color.Saturation, i.Color.Brightness)
	}

	return rgbColorBlock(r, g, b, "⬤")
}

func (i Item) Title() string {
	return style.SelectedBorder.Render(fmt.Sprintf("%s %s", i.StateSphere(), style.SelectedDevice.Render(i.Label)))
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
	renderFunc := func(w io.Writer, m list.Model, index int, listItem list.Item) {
		deviceItem, ok := listItem.(Item)
		if !ok {
			return
		}

		var str string
		if index == m.Index() {
			spStyle := style.ListSelected.Render(deviceItem.StateSphere())
			lbStyle := style.ListSelected.BorderLeft(false).Render(deviceItem.Label)
			str = fmt.Sprintf("%s%s", spStyle, lbStyle)
		} else {
			spStyle := style.ListItem.Render(deviceItem.StateSphere())
			lbStyle := style.ListItem.PaddingLeft(0).Render(deviceItem.Label)
			str = fmt.Sprintf("%s %s", spStyle, lbStyle)
		}

		fmt.Fprint(w, str)
	}
	d := hlist.NewDelegate(renderFunc, hlist.SetDelegateSpacing(1))

	f := func(i client.Device) list.Item { return Item(i) }
	l := hlist.New(devices, f, d)
	l.SetStatusBarItemName("device", "devices")
	return l
}

func rgbColorBlock(r, g, b int, text string) string {
	color := color.RGBToLipglossColor(r, g, b)
	return lipgloss.NewStyle().Foreground(color).Padding(0, 1, 0, 0).Render(text)
}
