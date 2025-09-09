package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type request struct {
	title, desc, method, endpoint, body, params, headers string
}

func (r request) Title() string       { return r.title }
func (r request) Description() string { return r.desc }
func (r request) Method() string      { return r.method }
func (r request) Endpoint() string    { return r.endpoint }
func (r request) Body() string        { return r.body }
func (r request) Params() string      { return r.params }
func (r request) Headers() string     { return r.headers }
func (r request) FilterValue() string { return r.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc string
	)

	if i, ok := listItem.(request); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Select method style
	var methodStyle lipgloss.Style
	var methodInactiveStyle lipgloss.Style
	switch desc {
	case "GET":
		methodStyle = getMethodStyle
		methodInactiveStyle = getMethodInactiveStyle
	case "POST":
		methodStyle = postMethodStyle
		methodInactiveStyle = postMethodInactiveStyle
	case "PUT":
		methodStyle = putMethodStyle
		methodInactiveStyle = putMethodInactiveStyle
	case "PATCH":
		methodStyle = patchMethodStyle
		methodInactiveStyle = patchMethodInactiveStyle
	case "DELETE":
		methodStyle = deleteMethodStyle
		methodInactiveStyle = deleteMethodInactiveStyle
	case "INFO":
		methodStyle = infoMethodStyle
		methodInactiveStyle = infoMethodInactiveStyle
	default:
		methodStyle = otherMethodStyle
		methodInactiveStyle = otherMethodInactiveStyle
	}

	// Prevent text from exceeding list width
	textwidth := m.Width() - NormalTitleStyle.GetPaddingLeft() - NormalTitleStyle.GetPaddingRight()
	title = ansi.Truncate(title, textwidth, "…")
	var lines []string
	for i, line := range strings.Split(desc, "\n") {
		if i >= d.Height()-1 {
			break
		}
		lines = append(lines, ansi.Truncate(line, textwidth, "…"))
	}
	desc = strings.Join(lines, "\n")

	// Render method with color
	methodRender := methodStyle.Render(desc)
	methodInactiveRender := methodInactiveStyle.Render(desc)

	if index == m.Index() {
		titleRender := selectedTitle.PaddingLeft(7 - len(desc)).Render("> " + title)
		fmt.Fprintf(w, "%s%s\n", methodRender, titleRender)
	} else {
		titleRender := NormalTitleStyle.PaddingLeft(10 - len(desc)).Render(title)
		fmt.Fprintf(w, "%s%s\n", methodInactiveRender, titleRender)
	}
}
