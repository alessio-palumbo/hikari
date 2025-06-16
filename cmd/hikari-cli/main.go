package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
		// Foreground(lipgloss.Color("#FFFDF5")).
		// Background(lipgloss.Color("#25A065")).
		// Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)

	responseStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1)
)

type state int

const (
	stateDeviceList state = iota
	stateDeviceSelected
	stateParamSelected
	stateEditingParam
	stateError
)

// Bubble Tea messages
type deviceSelectedMsg client.Device
type deviceUpdateMsg []client.Device
type tickMsg time.Time

type model struct {
	state              state
	deviceManager      *client.DeviceManager
	deviceList         list.Model
	selectedDevice     client.Device
	commandRegistry    *CommandRegistry
	commandList        list.Model
	selectedCommand    *Command
	paramList          list.Model
	editInput          textinput.Model
	selectedParamIndex int
	errMessage         string
	lastUpdate         time.Time
}

func initialModel() model {
	dm, err := client.NewDeviceManager()
	if err != nil {
		log.Fatal(err)
	}

	ti := textinput.New()
	ti.Placeholder = "Enter value"
	ti.Width = 20
	ti.CharLimit = 5
	ti.Focus()

	return model{
		state:           stateDeviceList,
		deviceManager:   dm,
		deviceList:      NewDeviceList(dm.GetDevices()),
		commandRegistry: NewCommandRegistry(),
		commandList:     NewCommandList(),
		lastUpdate:      time.Now(),
		editInput:       ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.deviceList.SetWidth(msg.Width)
		m.deviceList.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateDeviceList:
			switch msg.String() {
			case "enter":
				if selectedItem, ok := m.deviceList.SelectedItem().(deviceItem); ok {
					m.selectedDevice = selectedItem.device
					m.state = stateDeviceSelected
					return m, nil
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			m.deviceList, cmd = m.deviceList.Update(msg)

		case stateDeviceSelected:
			switch msg.String() {
			case "enter":
				if cmd, ok := m.commandList.SelectedItem().(Command); ok {
					m.selectedCommand = &cmd

					// Example: Execute command with some default parameters
					switch cmd.ID {
					case "power_on", "power_off":
						message, _ := m.selectedCommand.Handler()
						m.deviceManager.Send(m.selectedDevice.Address, message)
					case "set_color", "set_brightness":
						m.paramList = NewParamsList(m.selectedCommand.ParamTypes)
						m.paramList.Title = fmt.Sprintf("Params for %s", m.selectedCommand.Name)
						m.state = stateParamSelected
						return m, nil
					}
				}
			case "esc", "ctrl+b":
				m.state = stateDeviceList
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			m.commandList, cmd = m.commandList.Update(msg)

		case stateParamSelected:
			switch msg.String() {
			case "enter":
				m.selectedParamIndex = m.paramList.Index()
				paramItem := m.paramList.Items()[m.selectedParamIndex].(paramItem)

				m.editInput.SetValue(paramItem.param.Value)
				m.editInput.Placeholder = paramItem.param.Description
				m.editInput.CursorEnd()
				m.editInput.Focus()

				m.state = stateEditingParam
				return m, nil

			case "a":
				message, err := m.selectedCommand.Handler(m.selectedCommand.ParamTypes...)
				if err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				m.deviceManager.Send(m.selectedDevice.Address, message)
				m.state = stateDeviceSelected
				return m, nil

			case "esc", "ctrl+b":
				m.state = stateDeviceSelected
				return m, nil
			}
			m.paramList, cmd = m.paramList.Update(msg)

		case stateEditingParam:
			switch msg.String() {
			case "enter":
				val := m.editInput.Value()
				paramItem := m.paramList.Items()[m.paramList.Index()].(paramItem)
				if err := paramItem.param.ValidateValue(val); err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				m.errMessage = ""
				paramItem.param.Value = val
				m.state = stateParamSelected
				return m, nil

			case "esc", "ctrl+b":
				m.state = stateParamSelected
				return m, nil
			}
			m.editInput, cmd = m.editInput.Update(msg)
		case stateError:
		}

	case deviceSelectedMsg:
		m.selectedDevice = client.Device(msg)
		m.state = stateDeviceSelected
		return m, nil

	case deviceUpdateMsg:
		m.updateDeviceList([]client.Device(msg))
		m.lastUpdate = time.Now()
		return m, nil

	case tickMsg:
		// Only refresh if we're in the device list view or it's been a while
		if m.state == stateDeviceList || time.Since(m.lastUpdate) > 5*time.Second {
			return m, tea.Batch(m.refreshDevices(), m.tick())
		}
		return m, m.tick()
	}

	return m, cmd
}

// Command to refresh device list
func (m model) refreshDevices() tea.Cmd {
	return func() tea.Msg {
		devices := m.deviceManager.GetDevices()
		return deviceUpdateMsg(devices)
	}
}

// Command for periodic updates
func (m model) tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update the device list while preserving selection
func (m *model) updateDeviceList(devices []client.Device) {
	// Remember current selection
	var selectedIndex int
	var selectedSerial client.Serial
	if selectedItem, ok := m.deviceList.SelectedItem().(deviceItem); ok {
		selectedSerial = selectedItem.device.Serial
		selectedIndex = m.deviceList.Index()
	}

	// Update list items
	items := make([]list.Item, len(devices))
	newSelectedIndex := 0
	for i, device := range devices {
		items[i] = deviceItem{device: device}
		// Try to maintain selection on the same device
		if device.Serial == selectedSerial {
			newSelectedIndex = i
		}
	}

	m.deviceList.SetItems(items)

	// Restore selection if possible
	if len(items) > 0 {
		if newSelectedIndex < len(items) {
			m.deviceList.Select(newSelectedIndex)
		} else if selectedIndex < len(items) {
			m.deviceList.Select(selectedIndex)
		}
	}

	// Update selected device if it still exists
	if !selectedSerial.IsNil() {
		for _, device := range devices {
			if device.Serial == selectedSerial {
				m.selectedDevice = device
				break
			}
		}
	}
}

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
		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s",
			titleStyle.Render("Hikari"),
			deviceTitle(m.selectedDevice),
			m.commandList.View(),
			helpStyle.Render("↑/↓: navigate • enter: send • esc: back • ctrl+c: quit"),
		)

	case stateParamSelected:
		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s",
			titleStyle.Render("Hikari"),
			selectedStyle.Render(fmt.Sprintf("Command: %s", m.selectedCommand.Name)),
			m.paramList.View(),
			helpStyle.Render("↑/↓: navigate • enter: send • esc: back • ctrl+c: quit"),
		)
	case stateEditingParam:
		param := m.selectedCommand.ParamTypes[m.selectedParamIndex]
		var err string
		if m.errMessage != "" {
			err = fmt.Sprintf("\n\n❌ Error: %s", m.errMessage)
		}
		return fmt.Sprintf(
			"\nEditing parameter: %s\n\n%s%s\n\n%s",
			param.Name,
			m.editInput.View(),
			err,
			helpStyle.Render("↑/↓: navigate • enter: send • esc: back • ctrl+c: quit"),
		)
	}

	return ""
}

func main() {
	m := initialModel()
	defer m.deviceManager.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
