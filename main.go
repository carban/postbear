package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var models []tea.Model

const (
	mainScreen = iota
	formScreen
)

type MainModel struct {
	width, height int
	ready         bool
	requestsList  list.Model
	requestArea   viewport.Model
	response      viewport.Model
	spinner       spinner.Model
}

type request struct {
	title, env, method, headers, endpoint, body string
}

func (r request) Title() string       { return r.title }
func (r request) Env() string         { return r.env }
func (r request) Method() string      { return r.method }
func (r request) Headers() string     { return r.headers }
func (r request) Body() string        { return r.body }
func (r request) FilterValue() string { return r.title }

func (m MainModel) Init() tea.Cmd {
	return nil
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(request)
	if !ok {
		return
	}
	selectedItemStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Background(lipgloss.Color("#ffffff"))

	str := fmt.Sprintf("%d. %s", index+1, i.title)
	fn := lipgloss.NewStyle().PaddingLeft(4).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func (d itemDelegate) SelectedItem(m list.Model) list.Item {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		return nil
	}

	return items[i].(request)
}

func NewMainModel() *MainModel {

	li := list.New([]list.Item{
		request{title: "Foo Request", env: "fooapi.com", method: "GET", headers: "", endpoint: "/api/users", body: `{"data":"12222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222"}`},
		request{title: "Just a request", env: "fooapi.com", method: "GET", headers: "", endpoint: "/api/users", body: `"Nothing"`},
		request{title: "xdxd", env: "fooapi.com", method: "DELETE", headers: "", endpoint: "/api/users/5", body: `{"data":{"id":"1","name":"John","lastname":"Doe","username":"JohnxDoe11","birthdate":"1990-01-01","age":30,"gender":"Male","phone":"+63 791 675 8914","email":"foo@example.com","country":"USA","height":170,"weight":70}}`},
	}, itemDelegate{}, 0, 0)
	li.Title = "POSTBOY"

	spinner := spinner.New()
	spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	return &MainModel{
		width:        0,
		height:       0,
		ready:        false,
		requestsList: li,
		spinner:      spinner,
	}
}

var docStyle = lipgloss.NewStyle().BorderForeground(lipgloss.Color("33")).BorderStyle(lipgloss.NormalBorder()).Margin(2)

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.requestsList.SetSize(msg.Width-h, msg.Height-v)
		// m.requestArea.SetWidth(msg.Width - h)
		// m.requestArea.SetHeight(msg.Height - v*2)
		headerHeight := lipgloss.Height(m.headerView("", ""))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		// if !m.ready {
		// 	// Since this program is using the full size of the viewport we
		// 	// need to wait until we've received the window dimensions before
		// 	// we can initialize the viewport. The initial dimensions come in
		// 	// quickly, though asynchronously, which is why we wait for them
		// 	// here.
		// 	m.ready = true
		// }

		m.requestArea = viewport.New(msg.Width-h*33, msg.Height-verticalMarginHeight)
		m.requestArea.YPosition = headerHeight
		m.response = viewport.New(msg.Width-h*33, msg.Height-verticalMarginHeight)
		m.response.YPosition = headerHeight
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "n":
			models[mainScreen] = m
			models[formScreen] = NewFormModel()
			return models[formScreen].Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		case "r":
			m.runCall()
		}
	case FormModel:
		form := msg
		newForm := request{
			title:    form.inputs[title].Value(),
			env:      form.inputs[env].Value(),
			method:   form.methodChoices[form.methodCursor],
			endpoint: form.inputs[endpoint].Value(),
			body:     form.inputs[body].Value(),
		}
		return m, m.requestsList.InsertItem(len(m.requestsList.Items()), newForm)
	}
	m.requestsList, cmd = m.requestsList.Update(msg)
	return m, cmd
}

func (m *MainModel) runCall() {

	m.ready = false
	currentRequest := m.requestsList.SelectedItem().(request)
	fullurl := fmt.Sprintf("https://%s%s", currentRequest.env, currentRequest.endpoint)

	var req *http.Request
	var err error

	if currentRequest.body != "" {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			m.response.SetContent("error")
		}
		req, _ = http.NewRequest(currentRequest.method, fullurl, bytes.NewBuffer(bodyBytes))
	} else {
		req, _ = http.NewRequest(currentRequest.method, fullurl, nil)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		m.response.SetContent(string(err.Error()))
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		m.response.SetContent("error")
	}
	m.ready = true
	m.response.SetContent(FormatJSON(string(bodyBytes)))
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderForeground(lipgloss.Color("#00ff00")).BorderStyle(b).Width(16).AlignHorizontal(lipgloss.Center)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

var lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))

func FormatJSON(jsonStr string) string {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "error Unmarshal"
	}

	formattedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "error MarshalIndent"
	}

	// Apply styling using Lip Gloss and ANSI escape codes
	styledJSON := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(string(formattedJSON))

	// Add ANSI escape codes for specific elements (e.g., keys, values)
	styledJSON = strings.ReplaceAll(styledJSON, `"`, "\x1b[32m\"\x1b[0m") // Green quotes
	styledJSON = strings.ReplaceAll(styledJSON, "{", "\x1b[33m{\x1b[0m")  // Yellow brackets
	styledJSON = strings.ReplaceAll(styledJSON, "}", "\x1b[33m}\x1b[0m")  // Yellow brackets
	styledJSON = strings.ReplaceAll(styledJSON, ":", "\x1b[31m:\x1b[0m")
	styledJSON = strings.ReplaceAll(styledJSON, ",", "\x1b[36m,\x1b[0m")

	return string(styledJSON)
}

func (m MainModel) headerView(env, met string) string {
	titleEnv := titleStyle.Render(env)
	titleMethod := titleStyle.Render(met)
	line := strings.Repeat("─", max(0, m.requestArea.Width-lipgloss.Width(titleEnv)-lipgloss.Width(titleMethod)))
	return lipgloss.JoinHorizontal(lipgloss.Center, titleMethod, titleEnv, lineStyle.Margin(2, 0).Render(line))
}

func (m MainModel) footerView() string {
	info := infoStyle.Render("100")
	line := strings.Repeat("─", max(0, m.requestArea.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, lineStyle.Render(line), info)
}

func (m MainModel) View() string {
	currentRequest := m.requestsList.SelectedItem().(request)
	m.requestArea.SetContent(FormatJSON(currentRequest.Body()))
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Left,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, docStyle.Render(m.requestsList.View())),
			lipgloss.JoinVertical(
				lipgloss.Center,
				fmt.Sprintf("%s\n%s\n%s",
					m.headerView(currentRequest.env, currentRequest.method),
					m.requestArea.View(),
					m.footerView()),
			),
			lipgloss.JoinVertical(
				lipgloss.Center,
				fmt.Sprintf("%s\n%s\n%s",
					m.headerView("", ""),
					m.response.View(),
					m.footerView()),
			)))
}

func main() {
	m := NewMainModel()
	f := NewFormModel()
	models = []tea.Model{m, f}
	file, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	p := tea.NewProgram(models[mainScreen], tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
