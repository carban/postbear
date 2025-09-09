package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var footer string

type env struct {
	width       int
	height      int
	styles      *Styles
	returnModel Model
	content     textarea.Model
}

func environment(m Model) env {

	env := env{
		width:       m.width,
		height:      m.height,
		styles:      m.styles,
		returnModel: m,
		content:     newTextarea(),
	}
	gv := LoadGlobalVarsFromHTTPFile(m.filepath)
	jsonStr := "{}"
	if len(gv) > 0 {
		b, err := json.MarshalIndent(gv, "", "  ")
		if err == nil {
			jsonStr = string(b)
		}
	}

	env.content.SetValue(jsonStr)
	env.content.SetWidth(env.width - 2)
	env.content.SetHeight(env.height - 5)
	env.content.Placeholder = `
	{
	"Key":"Value",
}`

	footer = env.appBottomLabel("Ctrl+s to save variables, <ESC> to go back")

	return env

}

func (en *env) sizeInputs() {
	en.content.SetWidth(en.width - 2)
	en.content.SetHeight(en.height - 5)
}

func (en env) Init() tea.Cmd {
	return nil
}

func (en env) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	en.content.Focus()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return en, tea.Quit
		}
		if msg.String() == "esc" {
			en.returnModel.height = en.height
			en.returnModel.width = en.width
			return en.returnModel, nil
		}
		if msg.String() == "ctrl+s" {
			// Save env variables to .http file
			var globalVars map[string]string
			err := json.Unmarshal([]byte(en.content.Value()), &globalVars)
			if err != nil {
				footer = en.appBottomLabel("Error: Invalid JSON for environment variables")
			} else {
				// Load all requests from the current .http file (if any)
				data := &HTTPFileData{GlobalVars: globalVars}
				if loaded, err := LoadHTTPFile(en.returnModel.filepath); err == nil {
					data.Requests = loaded.Requests // preserve requests
				}
				// Only update the globalVars in the .http file, keep requests unchanged
				if err := SaveHTTPFile(data, en.returnModel.filepath); err != nil {
					footer = en.appBottomLabel("Error saving .http file")
				} else {
					footer = en.appBottomLabel(fmt.Sprintf("Environment variables saved to %s file!", en.returnModel.filepath))
				}
			}
		} else {
			footer = en.appBottomLabel("Ctrl+s to Save, <ESC> to go back")
		}

	case tea.WindowSizeMsg:
		en.width = msg.Width
		en.height = msg.Height
	}

	en.sizeInputs()

	en.content, cmd = en.content.Update(msg)
	cmds = append(cmds, cmd)

	return en, tea.Batch(cmds...)
}

func (en env) View() string {

	header := en.appTopLabel("POSTBOY Environment Varibales")

	body := borderStyle.Width(en.width - 2).Height(en.height - 4).Render(en.content.View())
	return en.styles.Base.Render(header + "\n" + body + "\n" + footer)
}
