package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)

	responseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1)
)

func (m model) View() string {
	switch m.state {
	case stateDeviceList:
		return fmt.Sprintf("%s\n%s\n%s\n%s",
			titleStyle.Render("Hikari"),
			m.deviceList.View(),
			statusStyle.Render(fmt.Sprintf("Last updated: %s | Devices: %d",
				m.lastUpdate.Format("15:04:05"), len(m.deviceList.Items()))),
			helpStyle.Render("↑/↓: navigate • enter: select device • q: quit"),
		)

	case stateDeviceSelected:
		status := "🟢 On"
		if !m.selectedDevice.PoweredOn {
			status = "🔴 Off"
		}

		return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n\n%s",
			titleStyle.Render("Send Message"),
			selectedStyle.Render(fmt.Sprintf("Selected: %s", m.selectedDevice.Label)),
			statusStyle.Render(fmt.Sprintf("Status: %s | Serial: %s", status, m.selectedDevice.Serial)),
			m.messageInput.View(),
			helpStyle.Render("enter: send message • esc: back to device list • ctrl+c: quit"),
		)

	case stateMessageInput:
		return fmt.Sprintf("%s\n\n%s",
			titleStyle.Render("Sending Message..."),
			statusStyle.Render("Please wait..."),
		)

	case stateResponse:
		var content string
		if m.err != nil {
			content = fmt.Sprintf("❌ Error: %v", m.err)
		} else {
			content = fmt.Sprintf("✅ %s", m.response)
		}

		return fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("Response"),
			responseStyle.Render(content),
			helpStyle.Render("enter/esc: back to message input • ctrl+c: quit"),
		)
	}

	return ""
}
