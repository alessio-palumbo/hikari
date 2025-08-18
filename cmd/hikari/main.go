package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync/atomic"
	"time"

	"github.com/alessio-palumbo/hikari/cmd/hikari/command"
	"github.com/alessio-palumbo/hikari/cmd/hikari/device"
	"github.com/alessio-palumbo/hikari/cmd/hikari/input"
	"github.com/alessio-palumbo/hikari/cmd/hikari/internal/version"
	"github.com/alessio-palumbo/hikari/cmd/hikari/style"
	ctrl "github.com/alessio-palumbo/lifxlan-go/pkg/controller"
	ldevice "github.com/alessio-palumbo/lifxlan-go/pkg/device"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultDeviceRefreshPeriod = 2 * time.Second
	defaultSendMessageSpinner  = 300 * time.Millisecond
	listWidth                  = 40
)

const (
	mappingQuit      = "q"
	mappingInfo      = "i"
	mappingSelect    = "enter"
	mappingSelectAlt = "e"
	mappingBack      = "left"
	mappingBackAlt   = "h"
	mappingSend      = "s"
)

var (
	filterExcludedBindings = []string{
		mappingQuit,
		mappingInfo,
		mappingSelect,
		mappingSelectAlt,
		mappingBack,
		mappingBackAlt,
		mappingSend,
	}
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
type deviceUpdateMsg []ldevice.Device
type msgSendDone struct{}
type effectStopDone struct{}
type tickMsg time.Time

type model struct {
	state              state
	deviceManager      *ctrl.Controller
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
	sending, stopping  bool
	effectStoppers     map[ldevice.Serial]*atomic.Bool
}

func initialModel() model {
	c, err := ctrl.New()
	if err != nil {
		log.Fatal(err)
	}
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = style.Spinner

	return model{
		state:          stateDeviceList,
		deviceManager:  c,
		deviceList:     device.NewList(c.GetDevices()),
		commandList:    command.NewList(),
		lastUpdate:     time.Now(),
		spinner:        s,
		effectStoppers: make(map[ldevice.Serial]*atomic.Bool),
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
		m.deviceList, cmd = m.deviceList.Update(msg)
	case tea.KeyMsg:
		switch m.state {
		case stateDeviceList:
			if shouldSkipBindingOnFilter(m.deviceList, msg.String()) {
				m.deviceList, cmd = m.deviceList.Update(msg)
				return m, cmd
			}
			switch msg.String() {
			case mappingSelect, mappingSelectAlt:
				if selectedDevice, ok := m.deviceList.SelectedItem().(device.Item); ok {
					m.selectedDevice = selectedDevice
					m.state = stateCommandList
				}
			case mappingInfo:
				m.showDeviceInfo = !m.showDeviceInfo
			case mappingQuit:
				return m, tea.Quit
			default:
				m.deviceList, cmd = m.deviceList.Update(msg)
			}

		case stateCommandList:
			switch msg.String() {
			case mappingSend:
				if commandItem, ok := m.commandList.SelectedItem().(command.Item); ok {
					m.selectedCommand = commandItem

					switch m.selectedCommand.ID {
					case "power_on", "power_off":
						message, _ := m.selectedCommand.Handler()
						m.deviceManager.Send(m.selectedDevice.Serial, message)
						return m.sendMessageSpinner()
					}
				}
			case mappingSelect, mappingSelectAlt:
				if commandItem, ok := m.commandList.SelectedItem().(command.Item); ok {
					m.selectedCommand = commandItem

					switch m.selectedCommand.ID {
					case "set_color", "set_brightness", "set_pixels":
						m.paramList = m.selectedCommand.NewParams()
						m.state = stateParamList
						return m, nil
					default:
						if m.selectedCommand.Type == command.CommandTypeEffect {
							if v, ok := m.effectStoppers[m.selectedDevice.Serial]; ok {
								v.Store(true)
								delete(m.effectStoppers, m.selectedDevice.Serial)
								return m.stopEffectSpinner()
							}
							m.paramList = m.selectedCommand.NewParams()
							m.state = stateParamList
							return m, nil
						}
					}
				}
			case mappingInfo:
				m.showDeviceInfo = !m.showDeviceInfo
			case mappingBack, mappingBackAlt:
				m.state = stateDeviceList
			case mappingQuit:
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
			case mappingSelect, mappingSelectAlt:
				paramItem.SetEdit(true, m.selectedDevice.MatrixProperties)
				m.paramList.SetItem(paramIndex, paramItem)
				m.state = stateParamEdit
			case mappingSend:
				switch m.selectedCommand.Type {
				case command.CommandTypeEffect:
					mProps := m.selectedDevice.MatrixProperties
					send := func(msg *protocol.Message) error {
						return m.deviceManager.Send(m.selectedDevice.Serial, msg)
					}
					stopped, err := m.selectedCommand.StartMatrixEffect(mProps, send, command.ParamItemsFromModel(m.paramList)...)
					if err != nil {
						m.errMessage = err.Error()
						return m, nil
					}
					m.effectStoppers[m.selectedDevice.Serial] = stopped
				default:
					message, err := m.selectedCommand.Handler(command.ParamItemsFromModel(m.paramList)...)
					if err != nil {
						m.errMessage = err.Error()
						return m, nil
					}
					m.deviceManager.Send(m.selectedDevice.Serial, message)
				}
				return m.sendMessageSpinner()
			case mappingBack, mappingBackAlt:
				paramItem.SetEdit(false)
				m.paramList.SetItem(paramIndex, paramItem)
				m.state = stateCommandList
			case mappingQuit:
				return m, tea.Quit
			default:
				m.paramList, cmd = m.paramList.Update(msg)
			}

		case stateParamEdit:
			paramIndex := m.paramList.GlobalIndex()
			paramItem := m.paramList.Items()[paramIndex].(command.ParamItem)

			switch msg.String() {
			case mappingSelect, mappingSelectAlt:
				if err := paramItem.SetValue(); err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
				paramItem.SetEdit(false)
				m.errMessage = ""
				m.state = stateParamList
			case mappingBack, mappingBackAlt:
				// Special handling for matrix input which requires directional keys.
				if paramItem.InputType == input.InputMatrixSelect {
					paramItem.UpdateValue(msg)
					break
				}

				paramItem.Input = paramItem.Input.Reset()
				_ = paramItem.SetValue()
				paramItem.SetEdit(false)
				m.errMessage = ""
				m.state = stateParamList
			default:
				paramItem.UpdateValue(msg)
			}
			m.paramList.SetItem(paramIndex, paramItem)
		}

	case tea.WindowSizeMsg:
		m.deviceList.SetWidth(msg.Width)
		m.deviceList.SetHeight(msg.Height - 4)

	case deviceUpdateMsg:
		cmd = m.updateDeviceList([]ldevice.Device(msg))
		m.lastUpdate = time.Now()

	case msgSendDone:
		m.sending = false
		m.state = stateCommandList

	case effectStopDone:
		m.stopping = false
		m.state = stateParamList

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

func (m model) stopEffectSpinner() (model, tea.Cmd) {
	m.stopping = true
	return m, tea.Batch(
		m.spinner.Tick,
		tea.Tick(defaultSendMessageSpinner, func(time.Time) tea.Msg {
			return effectStopDone{}
		}),
	)
}

// updateDeviceList updates the list of devices and keeps the current selection.
func (m *model) updateDeviceList(devices []ldevice.Device) tea.Cmd {
	var selectedSerial ldevice.Serial
	if selectedItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
		selectedSerial = selectedItem.Serial
	}

	items := make([]list.Item, len(devices))
	for i := range devices {
		d := device.Item(devices[i])
		items[i] = d
		// Update selectedDevice and its state so that it reflect in Commands and Params views.
		if d.Serial == m.selectedDevice.Serial {
			m.selectedDevice = d
		}
	}

	cmd := m.deviceList.SetItems(items)
	for i, d := range m.deviceList.VisibleItems() {
		if d.(device.Item).Serial == selectedSerial {
			m.deviceList.Select(i)
			break
		}
	}
	return cmd
}

func (m model) View() string {
	title := style.Title.Render("Hikari")
	switch m.state {
	case stateDeviceList:
		var d *device.Item
		if deviceItem, ok := m.deviceList.SelectedItem().(device.Item); ok {
			d = &deviceItem
		}
		return m.withDeviceInfoView(d, fmt.Sprintf("%s\n%s\n%s",
			title,
			m.renderStartupSpinnerOrDevices(),
			style.Status.Render(fmt.Sprintf("Last updated: %s | Devices: %d",
				m.lastUpdate.Format("15:04:05"), len(m.deviceList.Items()))),
		))

	case stateCommandList:
		return m.withDeviceInfoView(&m.selectedDevice, fmt.Sprintf("%s\n\n%s\n\n%s%s",
			title,
			m.selectedDevice.Title(),
			m.commandList.View(),
			m.renderSpinner(),
		))

	case stateParamList, stateParamEdit:
		return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s%s%s",
			title,
			m.selectedDevice.Title(),
			m.selectedCommand.Title(),
			m.paramList.View(),
			m.renderError(),
			m.renderSpinner(),
		)
	}

	return ""
}

func (m model) withDeviceInfoView(deviceItem *device.Item, view string) string {
	view = lipgloss.NewStyle().Width(listWidth).Render(view)
	if deviceItem != nil && m.showDeviceInfo {
		modal := "\n" + lipgloss.Place(0, 30,
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
	if m.stopping {
		return fmt.Sprint("\n\nStopping effect... ", m.spinner.View())
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
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		version.Print()
		os.Exit(0)
	}

	m := initialModel()
	defer m.deviceManager.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
