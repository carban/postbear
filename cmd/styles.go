package cmd

import "github.com/charmbracelet/lipgloss"

const maxWidth = 80

var (
	yellow = lipgloss.AdaptiveColor{Light: "#d79921", Dark: "#d79921"}
	indigo = lipgloss.AdaptiveColor{Light: "#458588", Dark: "#458588"}
	green  = lipgloss.AdaptiveColor{Light: "#8ec07c", Dark: "#8ec07c"}
	white  = lipgloss.AdaptiveColor{Light: "#d1d1d1ff", Dark: "#e5e5e5ff"}
)

var (
	// status style
	boldStyle     = lipgloss.NewStyle().Bold(true)
	blackColor    = lipgloss.Color("#000000")
	codes200Style = boldStyle.Foreground(blackColor).Background(lipgloss.Color("#82FFA1")).Padding(0, 1)
	codes500Style = boldStyle.Foreground(blackColor).Background(lipgloss.Color("#FF0000")).Padding(0, 1)
	codes400Style = boldStyle.Foreground(blackColor).Background(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	codes300Style = boldStyle.Foreground(blackColor).Background(lipgloss.Color("#4e98f8ff")).Padding(0, 1)

	// method style
	otherMethodColor  = lipgloss.Color("#d205cfff")
	getMethodColor    = lipgloss.Color("#12da00ff")
	postMethodColor   = lipgloss.Color("#05a2eaff")
	putMethodColor    = lipgloss.Color("#ffd000ff")
	patchMethodColor  = lipgloss.Color("#ff6f00ff")
	deleteMethodColor = lipgloss.Color("#ff0000ff")
	infoMethodColor   = lipgloss.Color("#42d6fbff")

	otherMethodStyle  = boldStyle.Foreground(blackColor).Background(otherMethodColor).Padding(0, 1)
	getMethodStyle    = boldStyle.Foreground(blackColor).Background(getMethodColor).Padding(0, 1)
	postMethodStyle   = boldStyle.Foreground(blackColor).Background(postMethodColor).Padding(0, 1)
	putMethodStyle    = boldStyle.Foreground(blackColor).Background(putMethodColor).Padding(0, 1)
	patchMethodStyle  = boldStyle.Foreground(blackColor).Background(patchMethodColor).Padding(0, 1)
	deleteMethodStyle = boldStyle.Foreground(blackColor).Background(deleteMethodColor).Padding(0, 1)
	infoMethodStyle   = boldStyle.Foreground(blackColor).Background(infoMethodColor).Padding(0, 1)

	otherMethodInactiveColor  = lipgloss.Color("#790877ff")
	getMethodInactiveColor    = lipgloss.Color("#128308ff")
	postMethodInactiveColor   = lipgloss.Color("#095274ff")
	putMethodInactiveColor    = lipgloss.Color("#7b660aff")
	patchMethodInactiveColor  = lipgloss.Color("#833f0bff")
	deleteMethodInactiveColor = lipgloss.Color("#900c0cff")
	infoMethodInactiveColor   = lipgloss.Color("#2c697eff")

	otherMethodInactiveStyle  = boldStyle.Foreground(blackColor).Background(otherMethodInactiveColor).Padding(0, 1)
	getMethodInactiveStyle    = boldStyle.Foreground(blackColor).Background(getMethodInactiveColor).Padding(0, 1)
	postMethodInactiveStyle   = boldStyle.Foreground(blackColor).Background(postMethodInactiveColor).Padding(0, 1)
	putMethodInactiveStyle    = boldStyle.Foreground(blackColor).Background(putMethodInactiveColor).Padding(0, 1)
	patchMethodInactiveStyle  = boldStyle.Foreground(blackColor).Background(patchMethodInactiveColor).Padding(0, 1)
	deleteMethodInactiveStyle = boldStyle.Foreground(blackColor).Background(deleteMethodInactiveColor).Padding(0, 1)
	infoMethodInactiveStyle   = boldStyle.Foreground(blackColor).Background(infoMethodInactiveColor).Padding(0, 1)
)

var (
	NormalTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#777777", Dark: "#777777"})

	selectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(white).
			Foreground(white)
)

var (
	spinnerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#E03535"))
	responseTimeStyle = lipgloss.NewStyle().Foreground(blackColor).Background(lipgloss.Color("#4e98f8ff"))
)

var (
	borderStyle = lipgloss.NewStyle().
			Padding(0, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(indigo)

	focusedBorder      = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(yellow)
	responseTitleStyle = lipgloss.NewStyle().Background(indigo).Foreground(lipgloss.Color("230"))
)

type Styles struct {
	Base,
	HeaderText,
	HeaderDecoration,
	Status,
	StatusHeader,
	Highlight,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(0, 0, 0, 0)
	s.HeaderText = lg.NewStyle().
		Foreground(yellow).
		Bold(true).
		Padding(0, 1, 0, 0)
	s.HeaderDecoration = lg.NewStyle().
		Foreground(indigo).
		Padding(0, 1, 0, 0)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

var (
	cursorStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color(indigo.Dark))
	cursorLineStyle         = lipgloss.NewStyle().Background(lipgloss.Color(green.Dark)).Foreground(lipgloss.Color("000"))
	placeholderStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	endOfBufferStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(indigo.Dark))
	focusedPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("000"))
)

var (
	inactiveTabStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, false).BorderForeground(green).Padding(0, 2).Margin(0, 1)
	activeTabStyle   = inactiveTabStyle.Border(lipgloss.ThickBorder(), false, false, true, false).Foreground(green)
)

func (m help) appTopLabel(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render("####  "+text), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
}

func (m help) appBottomLabel(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render("<--- "+text), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render("|||| "+text), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
}

func (m Model) appBoundaryMessage(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render(" "+text), lipgloss.WithWhitespaceChars("/"), lipgloss.WithWhitespaceForeground(yellow))
}

func (m env) appTopLabel(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render("####  "+text), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
}

func (m env) appBottomLabel(text string) string {
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Left, m.styles.HeaderText.Render("<--- "+text), lipgloss.WithWhitespaceChars("|"), lipgloss.WithWhitespaceForeground(indigo))
}
