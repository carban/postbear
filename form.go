package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	title = iota
	env
	endpoint
	method
	body
)

type FormModel struct {
	width, height, focused, zzz int
	ready                       bool
	inputs                      []textinput.Model
	methodChoices               []string
	methodCursor                int
	methodSelected              bool
	bodyArea                    textarea.Model
}

func (f *FormModel) nextInput() {
	if f.focused == len(f.inputs) {
		f.focused = 0
	} else {
		f.focused = (f.focused + 1) % len(f.inputs)
	}
}

func (f *FormModel) prevInput() {
	f.focused--
	// Wrap around
	if f.focused < 0 {
		f.focused = len(f.inputs) - 1
	}
}

func (f *FormModel) setMethodSelected() {
	f.methodSelected = true
}

func (f FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func NewFormModel() *FormModel {
	var inputs []textinput.Model = make([]textinput.Model, 5)

	inputs[title] = textinput.New()
	inputs[title].Placeholder = "Title"
	inputs[title].Focus()
	inputs[title].SetCursor(0)
	inputs[title].Cursor.Blink = true

	inputs[env] = textinput.New()
	inputs[env].Placeholder = "Environment"

	inputs[method] = textinput.New()
	inputs[method].Placeholder = "Method"

	inputs[endpoint] = textinput.New()
	inputs[endpoint].Placeholder = "Endpoint"

	inputs[body] = textinput.New()
	inputs[body].Placeholder = "Body"

	mc := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	ta := textarea.New()
	ta.Placeholder = "Body"

	return &FormModel{
		width:          0,
		height:         0,
		ready:          false,
		focused:        title,
		inputs:         inputs,
		methodChoices:  mc,
		methodCursor:   0,
		methodSelected: false,
		zzz:            2,
		bodyArea:       ta,
	}
}

func (f FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(f.inputs))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
		f.bodyArea.SetHeight((msg.Height / 3) + 1)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// if f.focused == len(f.inputs)-1 {
			// 	return f, tea.Quit
			// }
			if f.focused != body {
				f.nextInput()
			}
		case "ctrl+c":
			return f, tea.Quit
		case "schift+tab":
			f.prevInput()
		case "tab":
			f.nextInput()
			if f.focused == body {
				f.bodyArea.Focus()
			}
			if f.focused == method {
				f.setMethodSelected()
			}
		case "esc":
			models[formScreen] = f
			return models[mainScreen], nil
		case "k":
			if f.focused == method {
				if f.methodCursor == len(f.methodChoices)-1 {
					f.methodCursor = 0
				} else {
					f.methodCursor++
				}
			}
		case "j":
			if f.focused == method {
				if f.methodCursor == 0 {
					f.methodCursor = len(f.methodChoices) - 1
				} else {
					f.methodCursor--
				}
			}
		case "ctrl+s":
			models[formScreen] = f
			if f.bodyArea.Value() == "" {
				f.bodyArea.SetValue("{}")
			}
			return models[mainScreen], f.NewRequest
		}

		for i := range f.inputs {
			f.inputs[i].Blur()
		}
		f.inputs[f.focused].Focus()

		// We handle errors just like any other message
		// case errMsg:
		// 	f.err = msg
		// 	return m, nil
	}

	if f.focused == body {
		f.bodyArea, _ = f.bodyArea.Update(msg)
	}

	for i := range f.inputs {
		f.inputs[i], cmds[i] = f.inputs[i].Update(msg)
	}
	return f, tea.Batch(cmds...)

}

func (f FormModel) NewRequest() tea.Msg {
	f.inputs[body].SetValue(f.bodyArea.Value())
	r := FormModel{inputs: f.inputs, methodChoices: f.methodChoices, methodCursor: f.methodCursor}
	return r
}

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

func (f FormModel) focusedStyle(state int) lipgloss.Style {
	if state == f.focused {
		return lipgloss.NewStyle().Width(f.width / 3).Padding(0).BorderForeground(hotPink).BorderStyle(lipgloss.NormalBorder())
	}
	return lipgloss.NewStyle().Width(f.width / 3).Padding(0).BorderForeground(darkGray).BorderStyle(lipgloss.NormalBorder())
}

func (f FormModel) focusedStyleArea(state int) lipgloss.Color {
	if state == f.focused {
		return hotPink
	}
	return darkGray
}

func (f FormModel) View() string {
	// inputStyle := lipgloss.NewStyle().Width(f.width / 3).Padding(1).BorderForeground(darkGray).BorderStyle(lipgloss.NormalBorder())
	// continueStyle := lipgloss.NewStyle().Foreground(darkGray)
	selectedMethodStyle := lipgloss.NewStyle().Foreground(hotPink).PaddingLeft(0)
	s := strings.Builder{}
	for i := 0; i < len(f.methodChoices); i++ {
		if f.methodCursor == i {
			s.WriteString(selectedMethodStyle.Render(fmt.Sprintf("%s ", f.methodChoices[i])))
		} else {
			// s.WriteString("  ")
			s.WriteString(fmt.Sprintf("%s ", f.methodChoices[i]))
		}

		// s.WriteString("\n")
	}

	return lipgloss.Place(
		f.width,
		f.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			"Create a new request\n",
			// continueStyle.Render(fmt.Sprintf("1 of %d", f.zzz)),
			f.focusedStyle(title).Render(f.inputs[title].View()),
			f.focusedStyle(env).Render(f.inputs[env].View()),
			f.focusedStyle(endpoint).Render(f.inputs[endpoint].View()),
			lipgloss.NewStyle().Padding(1).Render(s.String()),
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				lipgloss.NewStyle().BorderForeground(f.focusedStyleArea(body)).BorderStyle(lipgloss.NormalBorder()).Padding(0, 0).Render(f.bodyArea.View()),
				lipgloss.NewStyle().BorderForeground(f.focusedStyleArea(body)).BorderStyle(lipgloss.NormalBorder()).Padding(0, 0).Render(f.bodyArea.View()),
			),
		),
	)

}
