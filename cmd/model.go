package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type Model struct {
	lg               *lipgloss.Renderer
	styles           *Styles
	width            int
	height           int
	requestsList     list.Model
	nameField        textinput.Model
	urlField         textinput.Model
	methodField      textinput.Model
	tabs             []string
	paramsTable      ParamsTable    // Params tab
	bodyArea         textarea.Model // Body tab
	headersArea      textarea.Model // Headers tab
	responseViewport viewport.Model
	activeTab        int
	response         string
	responseTime     string
	statusCode       string
	id               string
	focused          int
	fields           []string
	spinner          spinner.Model
	message          string
	loading          bool
	tabContentWidth  int
	filepath         string
}

const (
	requestsListPanel = iota
	nameFieldPanel
	methodFieldPanel
	urlFieldPanel
	tabContentPanel
	responseViewportPanel
)

const (
	paramsTab = iota
	bodyTab
	headersTab
)

type responseMsg struct {
	response     string
	statusCode   string
	responseTime string
}

func NewModel(filepath string) Model {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.id = ""
	m.tabs = []string{"Params", "Body", "Headers"}

	m.filepath = filepath
	m.nameField = textinput.New()
	m.nameField.Cursor.Blink = false
	m.nameField.Placeholder = "Name"
	m.nameField.Focus()
	m.nameField.Prompt = " "
	m.nameField.CharLimit = 22
	m.nameField.Cursor.Blink = true

	m.urlField = textinput.New()
	m.urlField.Placeholder = "Endpoint"
	m.urlField.Focus()
	m.urlField.Prompt = " "
	m.urlField.PromptStyle.Foreground(red)
	m.urlField.Cursor.Blink = true

	m.methodField = textinput.New()
	m.methodField.Placeholder = "Method"
	m.methodField.Focus()
	m.methodField.Prompt = " "
	m.methodField.CharLimit = 6
	m.methodField.Cursor.Blink = true

	headers := `{
	"Content-Type":"application/json",
	"Accept":"*/*",
	"Accept-Encoding":"gzip, deflate, br",
	"Connection":"keep-alive"
}`
	// Load requests from the current .http file (if any)
	var items []list.Item
	if data, err := LoadHTTPFile(m.filepath); err == nil {
		for _, req := range data.Requests {
			items = append(items, request{
				title:    req.Name,
				desc:     req.Method,
				method:   req.Method,
				endpoint: req.URL,
				body:     req.Body,
				params:   req.Params,
				headers:  req.Headers,
			})
		}
	}
	if len(items) == 0 {
		items = []list.Item{
			request{title: "New Request", desc: "GET", endpoint: "", method: "GET", headers: headers, params: "", body: ""},
		}
	}
	m.requestsList = list.New(items, itemDelegate{}, 0, 0)

	m.requestsList.Title = "POSTBOY"
	m.requestsList.SetStatusBarItemName("request", "requests")
	m.requestsList.SetWidth(37)
	m.requestsList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color(red.Dark)).
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(m.requestsList.Width()).
		Border(lipgloss.ThickBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(red.Dark))
	m.requestsList.SetShowHelp(false)
	m.requestsList.SetFilteringEnabled(false)

	m.activeTab = paramsTab
	m.paramsTable = NewParamsTable()
	m.bodyArea = newTextarea()
	m.bodyArea.Placeholder = `
{ 
	"your":"body" 
}`
	m.headersArea = newTextarea()
	m.headersArea.SetValue(createHeaders())

	vp := viewport.New(m.width, m.height)
	m.responseViewport = vp

	m.focused = requestsListPanel
	m.fields = []string{"requestList", "nameField", "methodField", "urlField", "tabContent", "responseViewport"}

	m.spinner = spinner.New()
	m.spinner.Spinner = spinner.Points
	m.spinner.Style = spinnerStyle
	m.message = m.appBoundaryView("Ctrl+c to quit, Ctrl+h for help")
	m.loading = false

	return m
}

// Init is run once when the program starts
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := borderStyle.GetFrameSize()
		m.requestsList.SetSize(msg.Width-h, msg.Height-v-1)
		m.tabContentWidth = (m.width - 40 - 8) / 2
		m.bodyArea.MaxWidth = m.tabContentWidth
		m.paramsTable.width = m.tabContentWidth
		m.headersArea.MaxWidth = m.tabContentWidth
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.focused != 4 {
				m.loading = true
				m.message = m.appBoundaryMessage("Sending Request....")
				m.spinner, cmd = m.spinner.Update(msg)
				cmds = append(cmds, cmd)
				// Perform the async operation in a goroutine
				return m, func() tea.Msg {
					response, statusCode, responseTime := sendByTUI(m) // Simulate the send function
					formattedResponse := formatJSON(response)
					responseTime = responseTimeStyle.Render(responseTime)
					statusStyle := codes200Style
					statusCodeInt, _ := strconv.Atoi(statusCode)
					if statusCodeInt >= 500 {
						statusStyle = codes500Style
					} else if statusCodeInt >= 400 && statusCodeInt < 500 {
						statusStyle = codes400Style
					} else if statusCodeInt >= 300 && statusCodeInt < 400 {
						statusStyle = codes300Style
					}

					statusCode = statusStyle.Render(statusCode)
					return responseMsg{
						response:     formattedResponse,
						statusCode:   statusCode,
						responseTime: responseTime,
					}
				}
			}
		case "ctrl+h":
			cmd = tea.EnterAltScreen
			help := newHelp(m.width, m.height, m.styles, &m)
			return help, nil
		case "shift+right":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case "shift+left":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "ctrl+e":
			environment := environment(m)
			return environment, nil
		case "ctrl+s":

			m.loading = true
			m.message = m.appBoundaryMessage("Saving Request....")
			// m.spinner, cmd = m.spinner.Update(msg)
			// cmds = append(cmds, cmd)

			// Perform the async save operation in a goroutine
			return m, func() tea.Msg {
				saveFile(m)
				return saveMsg{
					success: true,
					message: fmt.Sprintf("Request Saved in %s Successfully!", m.filepath),
				}
			}

		case "tab":
			m.focused = (m.focused + 1) % len(m.fields)
			m.message = m.appBoundaryView("Ctrl+c to quit, Ctrl+h for help")
		case "shift+tab":
			m.focused = m.focused - 1
			if m.focused < 0 {
				m.focused = len(m.fields) - 1
			}
			m.message = m.appBoundaryView("Ctrl+c to quit, Ctrl+h for help")
		case "n":
			// Add a new empty request and select it
			if m.focused == requestsListPanel {
				headers := `{
	"Content-Type":"application/json",
	"Accept":"*/*",
	"Accept-Encoding":"gzip, deflate, br",
	"Connection":"keep-alive"
}`

				newReq := request{
					title:    "New Request",
					desc:     "GET",
					method:   "GET",
					endpoint: "",
					body:     "",
					params:   "",
					headers:  headers,
				}
				m.requestsList.InsertItem(len(m.requestsList.Items()), newReq)
				m.requestsList.Select(len(m.requestsList.Items()) - 1)
				// Update fields to match new request
				m.nameField.SetValue(newReq.title)
				m.methodField.SetValue(strings.ToUpper(newReq.method))
				m.urlField.SetValue(newReq.endpoint)
				m.bodyArea.SetValue(newReq.body)
				// m.paramsTable.SetValue(newReq.params)
				m.headersArea.SetValue(newReq.headers)
				m.paramsTable = NewParamsTable()
				m.paramsTable.width = m.tabContentWidth
				return m, nil
			}
		}
	}

	m.sizeInputs()

	// Handle custom messages for async tasks
	switch msg := msg.(type) {
	case responseMsg:
		m.response = msg.response
		m.responseTime = msg.responseTime
		m.statusCode = msg.statusCode
		m.loading = false
		m.message = m.appBoundaryMessage("Request Sent!")

		wrappedContent := wordwrap.String(m.response, m.responseViewport.Width)
		m.responseViewport.SetContent(wrappedContent)
		m.responseViewport.GotoTop()
	case saveMsg:
		m.loading = false
		m.message = m.appBoundaryMessage(msg.message)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	// Update based on focus
	switch m.focused {
	case requestsListPanel:
		var cmdTemp tea.Cmd
		m.requestsList, cmdTemp = m.requestsList.Update(msg)
		cmds = append(cmds, cmdTemp)
		// If the index changed, update all fields to match the selected request
		if item, ok := m.requestsList.SelectedItem().(request); ok {
			m.nameField.SetValue(item.Title())
			m.methodField.SetValue(strings.ToUpper(item.Method()))
			m.urlField.SetValue(item.Endpoint())
			// m.paramsTable.SetValue(item.Params())
			m.bodyArea.SetValue(item.Body())
			m.headersArea.SetValue(item.Headers())
			// Sync paramsTable to selected request
			m.paramsTable.SetFromQueryString("") // Clear first
			if idx := strings.Index(item.Endpoint(), "?"); idx != -1 {
				m.paramsTable.SetFromQueryString(item.Endpoint()[idx:])
			}
		}
	case nameFieldPanel:
		m.nameField, cmd = m.nameField.Update(msg)
		cmds = append(cmds, cmd)
		// Sync change to requestsList
		if idx := m.requestsList.Index(); idx >= 0 {
			if item, ok := m.requestsList.SelectedItem().(request); ok {
				item.title = m.nameField.Value()
				m.requestsList.SetItem(idx, item)
			}
		}
	case methodFieldPanel:
		m.methodField, cmd = m.methodField.Update(msg)
		m.methodField.SetValue(strings.ToUpper(m.methodField.Value()))
		cmds = append(cmds, cmd)
		// Sync change to requestsList
		if idx := m.requestsList.Index(); idx >= 0 {
			if item, ok := m.requestsList.SelectedItem().(request); ok {
				item.method = m.methodField.Value()
				item.desc = m.methodField.Value()
				m.requestsList.SetItem(idx, item)
			}
		}
	case urlFieldPanel:
		m.urlField, cmd = m.urlField.Update(msg)
		cmds = append(cmds, cmd)
		m.urlField.CursorEnd()
		// Sync change to requestsList
		if idx := m.requestsList.Index(); idx >= 0 {
			if item, ok := m.requestsList.SelectedItem().(request); ok {
				item.endpoint = m.urlField.Value()
				m.requestsList.SetItem(idx, item)
			}
		}
		// Parse params from URL and update paramsTable
		urlVal := m.urlField.Value()
		paramsQuery := ""
		if idx := strings.Index(urlVal, "?"); idx != -1 {
			paramsQuery = urlVal[idx:]
		}
		m.paramsTable.SetFromQueryString(paramsQuery)
	case tabContentPanel:
		// With TABLE
		if m.activeTab == paramsTab {
			// Always focus last key when entering params tab, but allow typing in value if FocusedCol == 1
			if len(m.paramsTable.Rows) > 0 {
				for i, row := range m.paramsTable.Rows {
					if i == m.paramsTable.FocusedRow {
						if m.paramsTable.FocusedCol == 0 {
							row.KeyInput.Focus()
							row.ValueInput.Blur()
							row.KeyInput.Cursor.Blink = true
						} else {
							row.ValueInput.Focus()
							row.KeyInput.Blur()
							row.ValueInput.Cursor.Blink = true
						}
					} else {
						row.KeyInput.Blur()
						row.ValueInput.Blur()
					}
				}
			}
			m.paramsTable.Update(msg, m.tabContentWidth)
			// 	// Sync params to URL field
			paramsQuery := m.paramsTable.ToQueryString()
			baseUrl := m.urlField.Value()
			if idx := strings.Index(baseUrl, "?"); idx != -1 {
				baseUrl = baseUrl[:idx]
			}
			if paramsQuery != "" {
				m.urlField.SetValue(baseUrl + paramsQuery)
			} else {
				m.urlField.SetValue(baseUrl)
			}
			// Sync change to requestsList
			if idx := m.requestsList.Index(); idx >= 0 {
				if item, ok := m.requestsList.SelectedItem().(request); ok {
					item.params = m.paramsTable.ToQueryString()
					// item.params = m.tabContent[paramsTab].Value()
					item.endpoint = m.urlField.Value()
					m.requestsList.SetItem(idx, item)
				}
			}
		} else {
			if m.activeTab == bodyTab {
				m.bodyArea.Focus()
				m.bodyArea, cmd = m.bodyArea.Update(msg)
			} else {
				m.headersArea.Focus()
				m.headersArea, cmd = m.headersArea.Update(msg)
			}
			cmds = append(cmds, cmd)
			if idx := m.requestsList.Index(); idx >= 0 {
				if item, ok := m.requestsList.SelectedItem().(request); ok {
					if m.activeTab == bodyTab {
						item.body = m.bodyArea.Value()
					} else if m.activeTab == headersTab {
						item.headers = m.headersArea.Value()
					}
					m.requestsList.SetItem(idx, item)
				}
			}
		}

	case responseViewportPanel:
		// m.responseViewport, cmd = m.responseViewport.Update(msg)
		// cmds = append(cmds, cmd)
	}

	updateViewport := true
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "up" || keyMsg.String() == "down" {
			if m.focused != responseViewportPanel {
				updateViewport = false
			}
		}
	}
	if _, ok := msg.(tea.MouseMsg); ok {
		updateViewport = true
	}

	if updateViewport {
		m.responseViewport, cmd = m.responseViewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	// Combine all commands into a single tea.Cmd

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	var footer string

	doc := strings.Builder{}
	var renderedTabs []string

	for i, t := range m.tabs {
		var style lipgloss.Style
		isActive := i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		renderedTabs = append(renderedTabs, style.Render(t))
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	tabStyle := borderStyle
	if m.focused == tabContentPanel {
		if m.activeTab == 1 {
			m.bodyArea.Focus()
			m.headersArea.Blur()
		} else if m.activeTab == 2 {
			m.headersArea.Focus()
			m.bodyArea.Blur()
		} else {
			m.bodyArea.Blur()
			m.headersArea.Blur()
		}
		tabStyle = focusedBorder
	} else {
		m.bodyArea.Blur()
		m.headersArea.Blur()
	}
	tabView := m.headersArea.View()
	// With TABLE
	if m.activeTab == 0 {
		tabView = m.paramsTable.View()
	} else if m.activeTab == 1 {
		tabView = m.bodyArea.View()
	}

	tabContent := lipgloss.NewStyle().
		Width(m.tabContentWidth).
		Height(m.height - 8).
		Render(tabView)

	// With NO TABLE
	// tabContent = lipgloss.NewStyle().
	// 	Width(m.tabContentWidth).
	// 	Height(m.height - 8).
	// 	Render(m.tabContent[m.activeTab].View())

	combined := lipgloss.JoinVertical(lipgloss.Left, tabRow, tabContent)

	// Now, wrap the entire combined layout in a border.
	finalPanel := tabStyle.Render(combined)

	doc.WriteString(finalPanel)

	requestListStyle := borderStyle
	if m.focused == requestsListPanel {
		requestListStyle = focusedBorder
	}
	requestListBorderPanel := requestListStyle.Render(m.requestsList.View())

	nameStyle := borderStyle
	if m.focused == nameFieldPanel {
		m.nameField.Focus()
		nameStyle = focusedBorder

	} else {
		m.nameField.Blur()
	}
	nameInput := nameStyle.Width(25).Height(1).Render(m.nameField.View())

	// Render the Method input field
	methodStyle := borderStyle
	if m.focused == methodFieldPanel {
		m.methodField.Focus()
		methodStyle = focusedBorder
	} else {
		m.methodField.Blur()
	}
	methodInput := methodStyle.Width(9).Height(1).Render(m.methodField.View())

	// Render the URL input field
	urlStyle := borderStyle.Foreground(red)
	if m.focused == urlFieldPanel {
		m.urlField.Focus()
		urlStyle = focusedBorder.Foreground(red)
	} else {
		m.urlField.Blur()
	}
	urlInputWidth := m.width - 40 - 9 - 25 - (8 + 3)
	urlValue := m.urlField.Value()
	if urlInputWidth > 0 && len(urlValue)+5 > urlInputWidth {
		urlValue = urlValue[len(urlValue)+5-urlInputWidth:]
	} else if urlInputWidth <= 0 {
		urlValue = ""
	}
	m.urlField.SetValue(urlValue)
	urlInput := urlStyle.Width(urlInputWidth).Height(1).Render(m.urlField.View())

	// Render the Response
	responseStyle := borderStyle
	if m.focused == responseViewportPanel {
		responseStyle = focusedBorder
	} else {
		m.urlField.Blur()
	}

	requestPanel := doc.String()

	m.responseViewport.Height = m.height - 7
	m.responseViewport.Width = m.tabContentWidth - 1

	var responsePanel string
	if m.loading {
		spinnerView := m.spinner.View()
		responsePanel = responseStyle.Width(m.responseViewport.Width).Height(m.responseViewport.Height).Render(responseTitleStyle.Render(" Response: ") + " " + spinnerView + "\n" + m.responseViewport.View())
	} else {
		responsePanel = responseStyle.Width(m.responseViewport.Width).Height(m.responseViewport.Height).Render(responseTitleStyle.Render(" Response: ") + m.statusCode + m.responseTime + "\n" + m.responseViewport.View())
	}

	mainPanel := lipgloss.JoinHorizontal(lipgloss.Left, requestPanel, responsePanel)

	// Final Mounting Views
	topPanel := lipgloss.JoinHorizontal(lipgloss.Left, nameInput, methodInput, urlInput)

	body := lipgloss.JoinVertical(lipgloss.Top, topPanel, mainPanel)

	fullBodyWithList := lipgloss.JoinHorizontal(lipgloss.Left, requestListBorderPanel, body)

	if m.loading {
		spinnerView := m.spinner.View()
		footer = " " + spinnerView + " " + m.appBoundaryMessage(m.message)
	} else {
		footer = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render(" "+m.message), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
	}

	return m.styles.Base.Render(fullBodyWithList + "\n" + footer)
}

func (m *Model) sizeInputs() {
	m.bodyArea.SetWidth(int(float64(m.width)*0.5) - 2)
	m.bodyArea.SetHeight(m.height - 8)
	m.headersArea.SetWidth(int(float64(m.width)*0.5) - 2)
	m.headersArea.SetHeight(m.height - 8)
}

func saveAllToHTTPFile(m Model) error {
	// Collect all requests from the list
	var requests []HTTPRequest
	items := m.requestsList.Items()
	for i := 0; i < len(items); i++ {
		item := items[i]
		if req, ok := item.(request); ok {
			reqData := HTTPRequest{
				Name:    req.Title(),
				Method:  req.Method(),
				URL:     req.Endpoint(),
				Headers: req.Headers(),
				Body:    req.Body(),
				Params:  req.Params(),
			}
			requests = append(requests, reqData)
		}
	}

	// Load global variables
	globalVars := LoadGlobalVarsFromHTTPFile(m.filepath)

	data := &HTTPFileData{
		Requests:   requests,
		GlobalVars: globalVars,
	}
	return SaveHTTPFile(data, m.filepath)
}

// Replace the save() function to also save .http file
func saveFile(m Model) {
	_ = saveAllToHTTPFile(m)
}
