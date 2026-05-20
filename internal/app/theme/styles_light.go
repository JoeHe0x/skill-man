package theme

import "github.com/charmbracelet/lipgloss"

func newLightStyles() Styles {
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
			Foreground(lipgloss.Color("238")),
		AppPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			Italic(true),
		StatusBar: lipgloss.NewStyle().
			Padding(0, 2),
		StatusBarDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		StatusBarSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("235")),
		HeaderDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		Panel: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("236")),
		ItemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Bold(true),
		ItemDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		ItemMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		ItemBinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color("61")),
		ItemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true),
		Footer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		Hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		HintBold: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Bold(true),
		StatusReady: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")),
		StatusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("172")),
		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("160")).
			Bold(true),
		Modal: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(1, 2),
		ModalDanger: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("160")).
			Padding(1, 2),
		EmptyPreview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		TabBar: lipgloss.NewStyle().
			Padding(0, 2),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		TabSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		TabUnderline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")),
		HeaderBanner: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		PanelFocused: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		PanelBlur: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("252")).
			Padding(0, 1),
		PanelTitleFocus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")),
		PanelTitleBlur: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("241")),
		FooterFlash: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")),
		FooterContext: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")),
		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}
