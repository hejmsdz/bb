package main

import (
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hejmsdz/bb/model"
	"github.com/hejmsdz/bb/prs"
	"github.com/pkg/browser"
)

const listHeight = 14

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	paginationStyle       = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle             = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	ignoredStyle          = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	approvesStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	requestedChangesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	updatesStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	infoToastStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	successToastStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorToastStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type rootModel struct {
	list        list.Model
	prs         model.PrsModel
	ignores     model.IgnoresModel
	autoUpdate  model.AutoUpdateModel
	whatChanged model.WhatChangedModel
	localRepos  map[string]string
	quitting    bool
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		m.prs.Init(),
		m.ignores.Init(),
		m.autoUpdate.Init(),
	)
}

func NewInfoToast(text string) tea.Cmd {
	return func() tea.Msg {
		return model.MsgShowToast{Text: text, Style: infoToastStyle}
	}
}

func NewToast(text string, isOk bool) tea.Cmd {
	return func() tea.Msg {
		style := successToastStyle
		if !isOk {
			style = errorToastStyle
		}
		return model.MsgShowToast{Text: text, Style: style}
	}
}

func NewErrorToast(text string) tea.Cmd {
	return func() tea.Msg {
		return model.MsgShowToast{Text: text, Style: errorToastStyle}
	}
}

func OpenBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		browser.OpenURL(url)
		return nil
	}
}

func CopyToClipboard(str string, m rootModel) tea.Cmd {
	clipboard.WriteAll(str)
	return NewInfoToast("Copied '" + str + "'")
}

func UpdateListView(m *rootModel) {
	prItems := make([]list.Item, 0)
	for _, pr := range m.prs.Prs {
		if m.ignores.IsHidden(pr) {
			continue
		}

		prItems = append(prItems, PullRequestItem{
			pr,
			m.whatChanged.WhatChanged(pr),
			m.ignores.IsIgnored(pr),
		})
	}
	m.list.SetItems(prItems)
	if m.list.Cursor() >= len(prItems) {
		m.list.Select(len(prItems) - 1)
	}
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case model.MsgUpdateListView:
		UpdateListView(&m)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "r":
			return m, m.prs.StartLoadingPrs

		case ".":
			cmd := m.ignores.ToggleShowIgnored()
			return m, cmd
		}

		sel, ok := m.list.SelectedItem().(PullRequestItem)
		if !ok {
			return m, nil
		}

		switch keypress := msg.String(); keypress {
		case "i":
			cmd := m.ignores.ToggleIgnore(sel.Pr)
			return m, cmd

		case "d":
			cmd := m.whatChanged.DismissChanges(sel.Pr)
			return m, cmd

		case "u":
			return m, CopyToClipboard(sel.Pr.Url, m)

		case "b":
			return m, CopyToClipboard(sel.Pr.Branch, m)

		case "c":
			return m, Checkout(sel.Pr, m)

		case "P":
			return m, PullOrigin(sel.Pr, m)

		case "enter":
			return m, OpenBrowser(sel.Pr.Url)
		}
	}

	var listCmd, prsCmd, autoUpdateCmd, whatChangedCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.prs, prsCmd = m.prs.Update(msg)
	m.autoUpdate, autoUpdateCmd = m.autoUpdate.Update(msg)
	m.whatChanged, whatChangedCmd = m.whatChanged.Update(msg)

	return m, tea.Batch(listCmd, prsCmd, autoUpdateCmd, whatChangedCmd)
}

func (m rootModel) View() string {
	return m.list.View()
}

func main() {
	config := ReadConfig()
	c := prs.CreateBitbucketClient(config.Bitbucket)
	interval := time.Duration(config.UpdateIntervalMinutes) * time.Minute

	const defaultWidth = 20

	l := list.New(make([]list.Item, 0), model.NewItemDelegate(), defaultWidth, listHeight)
	l.Title = "Pull requests"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := rootModel{
		list:        l,
		prs:         model.NewPrsModel(c),
		ignores:     model.NewIgnoresModel(),
		autoUpdate:  model.NewAutoUpdateModel(interval),
		whatChanged: model.NewWhatChangedModel(),
		localRepos:  config.LocalRepositoryPaths,
	}

	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
