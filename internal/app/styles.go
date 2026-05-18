package app

import "github.com/charmbracelet/lipgloss"

type styles struct {
	doc             lipgloss.Style
	appTitle        lipgloss.Style
	appTitleCompact lipgloss.Style
	appVersion      lipgloss.Style
	appPath         lipgloss.Style
	statusBar       lipgloss.Style
	statusBarDim    lipgloss.Style
	statusBarSep    lipgloss.Style
	header          lipgloss.Style
	headerDim       lipgloss.Style
	panel           lipgloss.Style
	panelTitle      lipgloss.Style
	itemTitle       lipgloss.Style
	itemDesc        lipgloss.Style
	itemMeta        lipgloss.Style
	itemBinding     lipgloss.Style
	itemSelected    lipgloss.Style
	footer          lipgloss.Style
	hint            lipgloss.Style
	hintBold        lipgloss.Style
	statusReady     lipgloss.Style
	statusWarn      lipgloss.Style
	statusError     lipgloss.Style
	modal           lipgloss.Style
	modalDanger     lipgloss.Style
	emptyPreview    lipgloss.Style
	tabBar          lipgloss.Style
	tabActive       lipgloss.Style
	tabInactive     lipgloss.Style
	tabSep          lipgloss.Style
	headerBanner    lipgloss.Style
}

func newStyles() styles {
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
			Foreground(lipgloss.Color("245")).
			MarginRight(2),
		appPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
		statusBar: lipgloss.NewStyle().
			Padding(0, 2),
		statusBarDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		statusBarSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")),
		headerDim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		panel: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		panelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("223")),
		itemTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true),
		itemDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")),
		itemMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")),
		itemBinding: lipgloss.NewStyle().
			Foreground(lipgloss.Color("109")),
		itemSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
		footer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),
		hintBold: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true),
		statusReady: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		statusWarn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		statusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		modal: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("245")).
			Padding(1, 2),
		modalDanger: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2),
		emptyPreview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		tabBar: lipgloss.NewStyle().
			Padding(0, 2),
		tabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		tabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		tabSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		headerBanner: lipgloss.NewStyle().
			Border(baseBorder).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
	}
}
