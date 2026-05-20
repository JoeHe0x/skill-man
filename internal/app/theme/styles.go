package theme

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Doc             lipgloss.Style
	AppTitle        lipgloss.Style
	AppTitleCompact lipgloss.Style
	AppVersion      lipgloss.Style
	AppPath         lipgloss.Style
	StatusBar       lipgloss.Style
	StatusBarDim    lipgloss.Style
	StatusBarSep    lipgloss.Style
	Header          lipgloss.Style
	HeaderDim       lipgloss.Style
	Panel           lipgloss.Style
	PanelTitle      lipgloss.Style
	ItemTitle       lipgloss.Style
	ItemDesc        lipgloss.Style
	ItemMeta        lipgloss.Style
	ItemBinding     lipgloss.Style
	ItemSelected    lipgloss.Style
	Footer          lipgloss.Style
	Hint            lipgloss.Style
	HintBold        lipgloss.Style
	StatusReady     lipgloss.Style
	StatusWarn      lipgloss.Style
	StatusError     lipgloss.Style
	Modal           lipgloss.Style
	ModalDanger     lipgloss.Style
	EmptyPreview    lipgloss.Style
	TabBar          lipgloss.Style
	TabActive       lipgloss.Style
	TabInactive     lipgloss.Style
	TabSep          lipgloss.Style
	TabUnderline    lipgloss.Style
	HeaderBanner    lipgloss.Style
	PanelFocused    lipgloss.Style
	PanelBlur       lipgloss.Style
	PanelTitleFocus lipgloss.Style
	PanelTitleBlur  lipgloss.Style
	FooterFlash     lipgloss.Style
	FooterContext   lipgloss.Style
	HelpKey         lipgloss.Style
	HelpDesc        lipgloss.Style
}

func NewStyles(dark bool) Styles {
	if !dark {
		return newLightStyles()
	}
	baseBorder := lipgloss.RoundedBorder()

	return Styles{
		Doc: lipgloss.NewStyle().
			Padding(0, 1),
		AppTitle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			MarginRight(1),
		AppTitleCompact: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			MarginRight(1),
		AppVersion: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginRight(2),
		AppPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
		StatusBar: lipgloss.NewStyle().
			Padding(0, 2),
		StatusBarDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		StatusBarSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")),
		HeaderDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		Panel: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("223")),
		ItemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true),
		ItemDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")),
		ItemMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")),
		ItemBinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color("109")),
		ItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
		Footer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),
		HintBold: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true),
		StatusReady: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		Modal: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("245")).
			Padding(1, 2),
		ModalDanger: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2),
		EmptyPreview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		TabBar: lipgloss.NewStyle().
			Padding(0, 2),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		TabSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		TabUnderline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),
		HeaderBanner: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		PanelFocused: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1),
		PanelBlur: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),
		PanelTitleFocus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		PanelTitleBlur: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("245")),
		FooterFlash: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		FooterContext: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),
		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}
