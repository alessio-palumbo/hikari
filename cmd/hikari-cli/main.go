package main

import (
	"log"
	"time"

	"github.com/alessio-palumbo/hikari/pkg/client"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// App states
type state int

const (
	stateDeviceList state = iota
	stateDeviceSelected
	stateMessageInput
	stateResponse
)

// Bubble Tea messages
type deviceSelectedMsg client.Device
type messageResponseMsg struct {
	response string
	err      error
}
type deviceUpdateMsg []client.Device
type tickMsg time.Time

// Main model
type model struct {
	state          state
	deviceManager  *client.DeviceManager
	deviceList     list.Model
	selectedDevice client.Device
	messageInput   textinput.Model
	response       string
	err            error
	lastUpdate     time.Time
}

func initialModel() model {
	// Initialize device manager
	dm, err := client.NewDeviceManager()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)

	// Create device list
	devices := dm.GetDevices()
	items := make([]list.Item, len(devices))
	for i, device := range devices {
		items[i] = deviceItem{device: device}
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(5)
	deviceList := list.New(items, delegate, 0, 0)
	deviceList.Title = "LIFX Devices"

	// Create message input
	ti := textinput.New()
	ti.Placeholder = "Enter message to send..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return model{
		state:         stateDeviceList,
		deviceManager: dm,
		deviceList:    deviceList,
		messageInput:  ti,
		lastUpdate:    time.Now(),
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
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if selectedItem, ok := m.deviceList.SelectedItem().(deviceItem); ok {
					m.selectedDevice = selectedItem.device
					m.state = stateDeviceSelected
					m.messageInput.Focus()
					return m, nil
				}
			}
			m.deviceList, cmd = m.deviceList.Update(msg)

		case stateDeviceSelected:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "ctrl+b"))):
				m.state = stateDeviceList
				m.messageInput.Blur()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if m.messageInput.Value() != "" {
					m.state = stateMessageInput
					return m, m.sendMessage()
				}
			}
			m.messageInput, cmd = m.messageInput.Update(msg)

		case stateResponse:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "enter"))):
				m.state = stateDeviceSelected
				m.messageInput.SetValue("")
				m.response = ""
				m.err = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				return m, tea.Quit
			}
		}

	case deviceSelectedMsg:
		m.selectedDevice = client.Device(msg)
		m.state = stateDeviceSelected
		return m, nil

	case messageResponseMsg:
		m.response = msg.response
		m.err = msg.err
		m.state = stateResponse
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

func (m model) sendMessage() tea.Cmd {
	return func() tea.Msg {
		response, err := m.deviceManager.Send(m.selectedDevice.Serial, m.messageInput.Value())
		return messageResponseMsg{response: response, err: err}
	}
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

func main() {
	m := initialModel()
	defer m.deviceManager.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program: %v", err)
	}
}
