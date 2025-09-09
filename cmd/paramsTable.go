package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableRow struct {
	KeyInput   textinput.Model
	ValueInput textinput.Model
}

type ParamsTable struct {
	Rows       []TableRow
	FocusedRow int
	FocusedCol int // 0 for key, 1 for value
	width      int
}

func NewParamsTable() ParamsTable {
	row := TableRow{
		KeyInput:   textinput.New(),
		ValueInput: textinput.New(),
	}
	row.KeyInput.Placeholder = "Key"
	row.KeyInput.Prompt = ""
	row.ValueInput.Placeholder = "Value"
	row.ValueInput.Prompt = ""
	row.KeyInput.CharLimit = 26
	row.ValueInput.CharLimit = 26
	row.KeyInput.Focus()
	return ParamsTable{
		Rows:       []TableRow{row},
		FocusedRow: 0,
		FocusedCol: 0,
		width:      0,
	}
}

func (t *ParamsTable) AddRow() {
	if len(t.Rows) >= 10 {
		return
	}
	row := TableRow{
		KeyInput:   textinput.New(),
		ValueInput: textinput.New(),
	}
	row.KeyInput.Placeholder = "Key"
	row.KeyInput.Prompt = ""
	row.ValueInput.Placeholder = "Value"
	row.ValueInput.Prompt = ""
	row.KeyInput.CharLimit = 26
	row.ValueInput.CharLimit = 26
	t.Rows = append(t.Rows, row)
	t.FocusedRow = len(t.Rows) - 1
	t.FocusedCol = 0
	t.Rows[t.FocusedRow].KeyInput.Focus()
}

func (t *ParamsTable) Update(msg tea.Msg, width int) {
	t.width = width
	if len(t.Rows) == 0 {
		t.AddRow()
	}
	row := &t.Rows[t.FocusedRow]
	if t.FocusedCol == 0 {
		row.KeyInput, _ = row.KeyInput.Update(msg)
	} else {
		row.ValueInput, _ = row.ValueInput.Update(msg)
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			if t.FocusedCol == 0 {
				// If key is focused, move to value
				t.FocusedCol = 1
				row.ValueInput.Focus()
				row.KeyInput.Blur()
			} else {
				// If value is focused, add new row
				if row.KeyInput.Value() != "" || row.ValueInput.Value() != "" {
					row.KeyInput.Blur()
					t.AddRow()
				}
			}
		case "up":
			if t.FocusedRow > 0 {
				// Remove previous focus
				prevRow := &t.Rows[t.FocusedRow]
				prevRow.KeyInput.Blur()
				prevRow.ValueInput.Blur()
				t.FocusedRow--
				t.FocusedCol = 0
				row := &t.Rows[t.FocusedRow]
				row.KeyInput.Focus()
				row.ValueInput.Blur()
				row.KeyInput.Cursor.Blink = true
			}
		case "down":
			if t.FocusedRow < len(t.Rows)-1 {
				// Remove previous focus
				prevRow := &t.Rows[t.FocusedRow]
				prevRow.KeyInput.Blur()
				prevRow.ValueInput.Blur()
				t.FocusedRow++
				t.FocusedCol = 0
				row := &t.Rows[t.FocusedRow]
				row.KeyInput.Focus()
				row.ValueInput.Blur()
				row.KeyInput.Cursor.Blink = true
			}
		}
	}
}

func (t *ParamsTable) ToMap() map[string]string {
	m := make(map[string]string)
	for _, row := range t.Rows {
		k := strings.TrimSpace(row.KeyInput.Value())
		v := strings.TrimSpace(row.ValueInput.Value())
		if k != "" {
			m[k] = v
		}
	}
	return m
}

func (t *ParamsTable) ToQueryString() string {
	var params []string
	for _, row := range t.Rows {
		k := strings.TrimSpace(row.KeyInput.Value())
		v := strings.TrimSpace(row.ValueInput.Value())
		if k != "" {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
	}
	if len(params) == 0 {
		return ""
	}
	return "?" + strings.Join(params, "&")
}

func (t *ParamsTable) SetFromQueryString(query string) {
	t.Rows = nil
	if strings.HasPrefix(query, "?") {
		query = query[1:]
	}
	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		key := ""
		val := ""
		if len(kv) > 0 {
			key = kv[0]
		}
		if len(kv) > 1 {
			val = kv[1]
		}
		row := TableRow{
			KeyInput:   textinput.New(),
			ValueInput: textinput.New(),
		}
		row.KeyInput.Placeholder = "Key"
		row.KeyInput.Prompt = ""
		row.ValueInput.Placeholder = "Value"
		row.ValueInput.Prompt = ""
		row.KeyInput.CharLimit = 26
		row.ValueInput.CharLimit = 26
		row.KeyInput.SetValue(key)
		row.ValueInput.SetValue(val)
		t.Rows = append(t.Rows, row)
	}
	if len(t.Rows) == 0 {
		t.AddRow()
	}
	// Always focus last key
	t.FocusedRow = len(t.Rows) - 1
	t.FocusedCol = 0
	t.Rows[t.FocusedRow].KeyInput.Focus()
	t.Rows[t.FocusedRow].ValueInput.Blur()
	t.Rows[t.FocusedRow].KeyInput.Cursor.Blink = true
}

func (t *ParamsTable) View() string {
	var b strings.Builder
	for i, row := range t.Rows {
		keyStyle := lipgloss.NewStyle().Width(((t.width) / 2)).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(indigo)
		valueStyle := lipgloss.NewStyle().Width(((t.width) / 2)).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(indigo)
		if i == t.FocusedRow && t.FocusedCol == 0 {
			keyStyle = keyStyle.BorderForeground(green)
		}
		if i == t.FocusedRow && t.FocusedCol == 1 {
			valueStyle = valueStyle.BorderForeground(green)
		}
		keyView := keyStyle.Render(row.KeyInput.View())
		valueView := valueStyle.Render(row.ValueInput.View())
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, keyView, valueView))
	}
	return b.String()
}
