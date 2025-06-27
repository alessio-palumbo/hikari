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
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultDeviceRefreshPeriod = 2 * time.Second
	defaultSendMessageSpinner  = 300 * time.Millisecond
)

var (
	filterExcludedBindings = []string{"enter", "q", "i", "e", "b"}
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
type msgSendDone struct{}
type tickMsg time.Time

type model struct {
	state              state
	deviceManager      *client.DeviceManager
	deviceList         list.Model
	selectedDevice     device.Item
	showDeviceInfo     bool
	commandList        list.Model
	selectedCommand    command.Item
	paramList          list.Model
	selectedParamIndex int
	errMessage         string
	lastUpdate         time.Time
	spinner            spinner.Model
	sending            bool
}

func initialModel() model {
	dm, err := client.NewDeviceManager()
	if err != nil {
		log.Fatal(err)
	}
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = style.Spinner

	return model{
		state:         stateDeviceList,
		deviceManager: dm,
		deviceList:    device.NewList(dm.GetDevices()),
		commandList:   command.NewList(),
		lastUpdate:    time.Now(),
		spinner:       s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tick(),
		m.spinner.Tick,
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
			case "enter", "e":
				if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
					m.selectedDevice = deviceItem
					m.state = stateCommandList
				}
			case "i":
				m.showDeviceInfo = !m.showDeviceInfo
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				m.deviceList, cmd = m.deviceList.Update(msg)
			}

		case stateCommandList:
			if shouldSkipBindingOnFilter(m.commandList, msg.String()) {
				m.commandList, cmd = m.commandList.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case "enter", "e":
				if commandItem, ok := m.commandList.SelectedItem().(command.Item); ok {
					m.selectedCommand = commandItem

					switch m.selectedCommand.ID {
					case "power_on", "power_off":
						message, _ := m.selectedCommand.Handler()
						m.deviceManager.Send(m.selectedDevice.Address, message)
						return m.sendMessageSpinner()
					case "set_color", "set_brightness":
						m.paramList = m.selectedCommand.NewParams()
						m.state = stateParamList
						return m, nil
					}
				}
			case "i":
				m.showDeviceInfo = !m.showDeviceInfo
			case "esc", "b":
				m.state = stateDeviceList
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				m.commandList, cmd = m.commandList.Update(msg)
			}

		case stateParamList:
			if shouldSkipBindingOnFilter(m.paramList, msg.String()) {
				m.paramList, cmd = m.paramList.Update(msg)
				return m, cmd
			}

			paramIndex := m.paramList.GlobalIndex()
			paramItem := m.paramList.Items()[paramIndex].(command.ParamItem)

			switch msg.String() {
			case "enter", "e":
				paramItem.SetEdit(true)
				m.paramList.SetItem(paramIndex, paramItem)
				m.state = stateParamEdit
			case "a":
				message, err := m.selectedCommand.Handler(command.ParamItemsFromModel(m.paramList)...)
				if err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				m.deviceManager.Send(m.selectedDevice.Address, message)
				return m.sendMessageSpinner()
			case "esc", "b":
				paramItem.SetEdit(false)
				m.paramList.SetItem(paramIndex, paramItem)
				m.state = stateCommandList
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				m.paramList, cmd = m.paramList.Update(msg)
			}

		case stateParamEdit:
			paramIndex := m.paramList.GlobalIndex()
			paramItem := m.paramList.Items()[paramIndex].(command.ParamItem)

			switch msg.String() {
			case "enter", "e":
				val := paramItem.Input.Value()
				if err := paramItem.SetValue(val); err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				fallthrough
			case "esc", "b":
				m.errMessage = ""
				paramItem.SetEdit(false)
				m.paramList.SetItem(paramIndex, paramItem)
				m.state = stateParamList
			default:
				paramItem.Input, cmd = paramItem.Input.Update(msg)
				m.paramList.SetItem(paramIndex, paramItem)
			}
		}

	case tea.WindowSizeMsg:
		m.deviceList.SetWidth(msg.Width)
		m.deviceList.SetHeight(msg.Height - 4)

	case deviceSelectedMsg:
		m.selectedDevice = device.Item(msg)
		m.state = stateCommandList

	case deviceUpdateMsg:
		cmd = m.updateDeviceList([]client.Device(msg))
		m.lastUpdate = time.Now()

	case msgSendDone:
		m.sending = false
		m.state = stateCommandList

	case tickMsg:
		switch {
		case m.state == stateDeviceList:
			return m, tea.Batch(m.refreshDevices(), m.tick())
		case time.Since(m.lastUpdate) > 5*time.Second:
			return m, tea.Batch(m.refreshDevices(), m.tick())
		default:
			return m, m.tick()
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
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
	return tea.Tick(defaultDeviceRefreshPeriod, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) sendMessageSpinner() (model, tea.Cmd) {
	m.sending = true
	return m, tea.Batch(
		m.spinner.Tick,
		tea.Tick(defaultSendMessageSpinner, func(time.Time) tea.Msg {
			return msgSendDone{}
		}),
	)
}

// updateDeviceList updates the list of devices and keeps the current selection.
func (m *model) updateDeviceList(devices []client.Device) tea.Cmd {
	// Remember current selection
	var selectedSerial client.Serial
	if selectedItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
		selectedSerial = selectedItem.Serial
	}

	items := make([]list.Item, len(devices))
	for i := range devices {
		d := device.Item(devices[i])
		items[i] = d
		if d.Serial == selectedSerial {
			m.selectedDevice = d
		}
	}

	return m.deviceList.SetItems(items)
}

func (m model) View() string {
	title := style.Title.Render("Hikari")
	switch m.state {
	case stateDeviceList:
		return m.withDeviceInfoView(fmt.Sprintf("%s\n%s\n%s\n%s",
			title,
			m.renderStartupSpinnerOrDevices(),
			style.Status.Render(fmt.Sprintf("Last updated: %s | Devices: %d",
				m.lastUpdate.Format("15:04:05"), len(m.deviceList.Items()))),
			style.Help.Render("↑/↓: navigate • enter: select device • q: quit| devices"),
		))

	case stateCommandList:
		return m.withDeviceInfoView(fmt.Sprintf("%s\n\n%s\n%s%s\n\n%s",
			title,
			m.selectedDevice.Title(),
			m.commandList.View(),
			m.renderSpinner(),
			style.Help.Render("↑/↓: navigate • enter: select • esc: back • q: quit"),
		))

	case stateParamList, stateParamEdit:
		return fmt.Sprintf("%s\n\n%s\n\n%s\n%s%s%s\n\n%s",
			title,
			m.selectedDevice.Title(),
			m.selectedCommand.Title(),
			m.paramList.View(),
			m.renderError(),
			m.renderSpinner(),
			style.Help.Render("↑/↓: navigate • enter: edit • a: apply • esc: back • q: quit"),
		)
	}

	return ""
}

func (m model) withDeviceInfoView(view string) string {
	if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok && m.showDeviceInfo {
		modal := "\n" + lipgloss.Place(20, m.deviceList.Height(),
			lipgloss.Left, lipgloss.Top,
			deviceItem.Info(),
		)

		return lipgloss.JoinHorizontal(lipgloss.Top, view, modal)
	}
	return view
}

func (m model) renderError() string {
	if m.errMessage != "" {
		return fmt.Sprintf("\n\n❌ Error: %s", m.errMessage)
	}
	return ""
}

func (m model) renderSpinner() string {
	if m.sending {
		return fmt.Sprint("\n\nSending... ", m.spinner.View())
	}
	return ""
}

func (m model) renderStartupSpinnerOrDevices() string {
	if len(m.deviceList.Items()) == 0 {
		return fmt.Sprintf("\n\nDiscovering %s\n\n", m.spinner.View())
	}
	return m.deviceList.View()
}

func main() {
	m := initialModel()
	defer m.deviceManager.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
