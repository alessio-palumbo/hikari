package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/command"
	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/device"
	"github.com/alessio-palumbo/hikari/cmd/hikari-cli/style"
	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultDeviceRefreshPeriod = 2 * time.Second
)

var (
	filterExcludedBindings = []string{"enter", "q"}
)

type state int

const (
	stateDeviceList state = iota
	stateCommandList
	stateParamList
	stateParamEdit
	stateError
)

// Bubble Tea messages
type deviceSelectedMsg client.Device
type deviceUpdateMsg []client.Device
type tickMsg time.Time

type model struct {
	state               state
	deviceManager       *client.DeviceManager
	deviceRefreshPeriod time.Duration
	deviceList          list.Model
	selectedDevice      device.Item
	showDeviceInfo      bool
	commandList         list.Model
	selectedCommand     command.Item
	paramList           list.Model
	editInput           textinput.Model
	selectedParamIndex  int
	errMessage          string
	lastUpdate          time.Time
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
		state:               stateDeviceList,
		deviceManager:       dm,
		deviceRefreshPeriod: defaultDeviceRefreshPeriod,
		deviceList:          device.NewList(dm.GetDevices()),
		commandList:         command.NewList(),
		lastUpdate:          time.Now(),
		editInput:           ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tick(),
	)
}

func shouldSkipBindingOnFilter(l list.Model, keypress string) bool {
	return l.FilterState() == list.Filtering && slices.Contains(filterExcludedBindings, keypress)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case list.FilterMatchesMsg:
		switch m.state {
		case stateDeviceList:
			m.deviceList, cmd = m.deviceList.Update(msg)
		case stateCommandList:
			m.commandList, cmd = m.commandList.Update(msg)
		case stateParamList:
			m.paramList, cmd = m.paramList.Update(msg)
		}
	case tea.KeyMsg:
		switch m.state {
		case stateDeviceList:
			if shouldSkipBindingOnFilter(m.deviceList, msg.String()) {
				m.deviceList, cmd = m.deviceList.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
					m.selectedDevice = deviceItem
					m.state = stateCommandList
					m.deviceList.ResetFilter()
					return m, nil
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "i":
				if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
					m.selectedDevice = deviceItem
					m.showDeviceInfo = !m.showDeviceInfo
					return m, nil
				}
			}
			m.deviceList, cmd = m.deviceList.Update(msg)

		case stateCommandList:
			if shouldSkipBindingOnFilter(m.commandList, msg.String()) {
				m.commandList, cmd = m.commandList.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				if commandItem, ok := m.commandList.SelectedItem().(command.Item); ok {
					m.selectedCommand = commandItem

					switch m.selectedCommand.ID {
					case "power_on", "power_off":
						message, _ := m.selectedCommand.Handler()
						m.deviceManager.Send(m.selectedDevice.Address, message)
					case "set_color", "set_brightness":
						m.paramList = m.selectedCommand.NewParams()
						m.state = stateParamList
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

		case stateParamList:
			if shouldSkipBindingOnFilter(m.paramList, msg.String()) {
				m.paramList, cmd = m.paramList.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter":
				m.selectedParamIndex = m.paramList.Index()
				paramItem := m.paramList.Items()[m.selectedParamIndex].(command.ParamItem)

				m.editInput.SetValue(paramItem.GetValue())
				m.editInput.Placeholder = paramItem.Description
				m.editInput.CursorEnd()
				m.editInput.Focus()

				m.state = stateParamEdit
				return m, nil

			case "a":
				message, err := m.selectedCommand.Handler(command.ParamItemsFromModel(m.paramList)...)
				if err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				m.deviceManager.Send(m.selectedDevice.Address, message)
				m.state = stateCommandList
				return m, nil

			case "esc", "ctrl+b", "backspace":
				m.state = stateCommandList
				return m, nil
			}
			m.paramList, cmd = m.paramList.Update(msg)

		case stateParamEdit:
			switch msg.String() {
			case "enter":
				val := m.editInput.Value()
				paramItem := m.paramList.Items()[m.paramList.Index()].(command.ParamItem)
				if err := paramItem.ValidateValue(val); err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				m.errMessage = ""
				paramItem.SetValue(val)
				m.state = stateParamList
				return m, nil

			case "esc", "ctrl+b":
				m.state = stateParamList
				return m, nil
			}
			m.editInput, cmd = m.editInput.Update(msg)
		case stateError:
		}

	case tea.WindowSizeMsg:
		m.deviceList.SetWidth(msg.Width)
		m.deviceList.SetHeight(msg.Height - 4)
		return m, nil

	case deviceSelectedMsg:
		m.selectedDevice = device.Item(msg)
		m.state = stateCommandList
		return m, nil

	case deviceUpdateMsg:
		m.updateDeviceList([]client.Device(msg))
		m.lastUpdate = time.Now()
		return m, nil

	case tickMsg:
		// Only refresh if we're in the device list view or it's been a while
		switch {
		case m.state == stateDeviceList && m.deviceList.FilterState() == list.Unfiltered:
			return m, tea.Batch(m.refreshDevices(), m.tick())
		case m.state != stateDeviceList && time.Since(m.lastUpdate) > 5*time.Second:
			return m, tea.Batch(m.refreshDevices(), m.tick())
		default:
			return m, m.tick()
		}
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
	return tea.Tick(m.deviceRefreshPeriod, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update the device list while preserving selection
func (m *model) updateDeviceList(devices []client.Device) {
	// Remember current selection
	var selectedIndex int
	var selectedSerial client.Serial
	if selectedItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
		selectedIndex = m.deviceList.Index()
		selectedSerial = selectedItem.Serial
	}

	// Update list items
	items := make([]list.Item, len(devices))
	newSelectedIndex := 0
	for i, d := range devices {
		items[i] = device.Item(d)
		// Try to maintain selection on the same device
		if d.Serial == selectedSerial {
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
		for _, d := range devices {
			if d.Serial == selectedSerial {
				m.selectedDevice = device.Item(d)
				break
			}
		}
	}
}

func (m model) View() string {
	title := style.Title.Render("Hikari")
	switch m.state {
	case stateDeviceList:
		deviceView := fmt.Sprintf("%s\n%s\n%s\n%s",
			title,
			m.deviceList.View(),
			style.Status.Render(fmt.Sprintf("Last updated: %s | Devices: %d",
				m.lastUpdate.Format("15:04:05"), len(m.deviceList.Items()))),
			style.Help.Render("↑/↓: navigate • enter: select device • q: quit| devices"),
		)
		var modal string
		if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok && m.showDeviceInfo {
			modal = "\n" + lipgloss.Place(20, m.deviceList.Height(),
				lipgloss.Left, lipgloss.Top,
				deviceItem.Info(),
			)

			return lipgloss.JoinHorizontal(lipgloss.Top, deviceView, modal)
		}
		return deviceView

	case stateCommandList:
		return fmt.Sprintf("%s\n\n%s\n%s\n\n%s",
			title,
			m.selectedDevice.Title(),
			m.commandList.View(),
			style.Help.Render("↑/↓: navigate • enter: select • esc: back • q: quit"),
		)

	case stateParamList:
		return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n\n%s",
			title,
			m.selectedDevice.Title(),
			m.selectedCommand.Title(),
			m.paramList.View(),
			style.Help.Render("↑/↓: navigate • enter: edit • esc: back • q: quit"),
		)
	case stateParamEdit:
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s\n\n%s%s\n\n%s",
			title,
			m.selectedDevice.Title(),
			m.paramList.Items()[m.paramList.Index()].(command.ParamItem).Title(),
			m.editInput.View(),
			m.renderError(),
			style.Help.Render("↑/↓: navigate • enter: set • esc: back • q: quit"),
		)
	}

	return ""
}

func (m model) renderError() string {
	if m.errMessage != "" {
		return fmt.Sprintf("\n\n❌ Error: %s", m.errMessage)
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
