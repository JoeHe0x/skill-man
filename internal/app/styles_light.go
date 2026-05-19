package app

import "github.com/charmbracelet/lipgloss"

func newLightStyles() styles {
	baseBorder := lipgloss.RoundedBorder()

	return styles{
		doc: lipgloss.NewStyle().
			Padding(0, 1),
		appTitle: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			MarginRight(1),
		appTitleCompact: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			MarginRight(1),
		appVersion: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		appPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			Italic(true),
		statusBar: lipgloss.NewStyle().
			Padding(0, 2),
		statusBarDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		statusBarSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("235")),
		headerDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		panel: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		panelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("236")),
		itemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Bold(true),
		itemDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		itemMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		itemBinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color("61")),
		itemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true),
		footer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		hintBold: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Bold(true),
		statusReady: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")),
		statusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("172")),
		statusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("160")).
			Bold(true),
		modal: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(1, 2),
		modalDanger: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("160")).
			Padding(1, 2),
		emptyPreview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		tabBar: lipgloss.NewStyle().
			Padding(0, 2),
		tabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")),
		tabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		tabSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		tabUnderline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")),
		headerBanner: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("250")).
			Padding(0, 1),
		panelFocused: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		panelBlur: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("252")).
			Padding(0, 1),
		panelTitleFocus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")),
		panelTitleBlur: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("241")),
		footerFlash: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")),
		footerContext: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		helpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")),
		helpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}
